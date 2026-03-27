package main

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	dbpkg "xyrTools/passwordManage/passwordManageServer/pkg/db"
	"xyrTools/passwordManage/passwordManageServer/pkg/encryption"
	"xyrTools/passwordManage/passwordManageServer/pkg/mail"
	respMessage "xyrTools/passwordManage/passwordManageServer/pkg/responseStd"
)

// record 定义结构体，供数据更新函数使用
type record struct {
	id           int
	choseEncrypt string
	key          string
}

// totp数据结构
type TotpRequestData struct {
	SecretKey     string `json:"secretKey"`
	RefreshPeriod int    `json:"refreshPeriod"`
	AppName       string `json:"appName"`
	Username      string `json:"username"`
	Digits        int    `json:"digits"`
	Enabled       bool   `json:"enabled"`
}

// totp整体设置
type TotpAllSet struct {
	SecretKey     string `json:"secretKey"`
	RefreshPeriod int    `json:"refreshPeriod"`
	AppName       string `json:"appName"`
	Username      string `json:"username"`
	Digits        int    `json:"digits"`
	Enabled       bool   `json:"enabled"`
	SetupStatus   bool   `json:"setupStatus"`
}

// 数据库存储登录告警邮件设置
type AlertMailStorage struct {
	Data          json.RawMessage `json:"data"`
	EncryptSecret bool            `json:"encryptSecret"`
	Enabled       bool            `json:"enabled"`
}

// 白名单配置
type WhitelistConfig struct {
	Enabled                bool     `json:"enabled"`
	Whitelist              []string `json:"whitelist"`
	ActionOutsideWhitelist string   `json:"actionOutsideWhitelist"`
}

// deriveSharedSecret 从私钥和公钥派生共享密钥
func deriveSharedSecret(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) ([]byte, error) {
	x, _ := privateKey.Curve.ScalarMult(publicKey.X, publicKey.Y, privateKey.D.Bytes())
	secret := x.Bytes()
	// 如果不足 66 字节，手动填充前导零
	if len(secret) < 66 {
		paddedSecret := make([]byte, 66)
		copy(paddedSecret[66-len(secret):], secret)
		secret = paddedSecret
	}
	// 强制将最后一个字节的前 7 比特置零
	secret[65] &= 0b10000000
	return secret, nil
}

// deriveAESKey 从共享密钥派生 AES 密钥
func deriveAESKey(sharedSecret []byte) ([]byte, error) {
	hasher := sha256.New()
	hasher.Write(sharedSecret)
	aesKey := hasher.Sum(nil)
	return aesKey[:32], nil // 取前 32 字节作为 AES-256 密钥
}

// updateSavedData 更新已保存的数据
// 参数：
// - newAesKey: 新的AES密钥
// - oldAesKey: 旧的AES密钥
// - dbPath: 数据库路径
func updateSavedData(newAesKey, oldAesKey []byte, dbPath string) {
	fmt.Println("[+] 开始更新数据库中已保存的数据")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("[-] 连接数据库失败: %v\n", err)
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, chose_encrypt, key FROM input_data`)
	if err != nil {
		fmt.Printf("[-] 查询数据失败: %v\n", err)
		return
	}
	defer rows.Close()

	var (
		totalCount   int
		updatedCount int
		skipCount    int
		batchSize    = 500
	)

	// 缓存批次
	batch := make([]record, 0, batchSize)

	for rows.Next() {
		var r record
		if err := rows.Scan(&r.id, &r.choseEncrypt, &r.key); err != nil {
			fmt.Printf("[-] 扫描行数据失败: %v\n", err)
			skipCount++
			continue
		}
		batch = append(batch, r)
		totalCount++

		// 满一个批次就处理
		if len(batch) >= batchSize {
			uc, sc := processBatch(db, batch, oldAesKey, newAesKey)
			updatedCount += uc
			skipCount += sc
			batch = batch[:0]
		}
	}

	// 处理最后不足批次的数据
	if len(batch) > 0 {
		uc, sc := processBatch(db, batch, oldAesKey, newAesKey)
		updatedCount += uc
		skipCount += sc
	}

	if err := rows.Err(); err != nil {
		fmt.Printf("[-] 遍历记录时出错: %v\n", err)
	}

	fmt.Printf("[+] 数据库更新完成，共查询 %d 条，成功更新 %d 条，跳过 %d 条\n",
		totalCount, updatedCount, skipCount)
}

// processBatch 每个批次事务处理
// 参数：
// - db: 数据库连接对象
// - batch: 记录批次
// - oldAesKey: 旧的AES密钥
// - newAesKey: 新的AES密钥
// 返回值：
// - 成功更新的记录数
// - 跳过的记录数
func processBatch(db *sql.DB, batch []record, oldAesKey, newAesKey []byte) (int, int) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("[-] 开始事务失败: %v\n", err)
		return 0, len(batch)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var updatedCount, skipCount int
	for _, r := range batch {
		// 解密
		var decryptedChoseEncrypt, decryptedKey string
		if r.choseEncrypt != "" {
			val, err := decryptField(r.choseEncrypt, oldAesKey)
			if err != nil {
				fmt.Printf("[-] 解密 choseEncrypt 失败(id=%d): %v\n", r.id, err)
				skipCount++
				continue
			}
			decryptedChoseEncrypt = val
		}
		if r.key != "" {
			val, err := decryptField(r.key, oldAesKey)
			if err != nil {
				fmt.Printf("[-] 解密 key 失败(id=%d): %v\n", r.id, err)
				skipCount++
				continue
			}
			decryptedKey = val
		}

		// 加密
		var encryptedChoseEncrypt, encryptedKey string
		if decryptedChoseEncrypt != "" {
			encryptedChoseEncrypt, err = encryptField(decryptedChoseEncrypt, newAesKey)
			if err != nil {
				fmt.Printf("[-] 加密 choseEncrypt 失败(id=%d): %v\n", r.id, err)
				skipCount++
				continue
			}
		}
		if decryptedKey != "" {
			encryptedKey, err = encryptField(decryptedKey, newAesKey)
			if err != nil {
				fmt.Printf("[-] 加密 key 失败(id=%d): %v\n", r.id, err)
				skipCount++
				continue
			}
		}

		// 更新
		_, err = tx.Exec(`UPDATE input_data 
			SET chose_encrypt = ?, 
			    "key" = ?, 
			    updated_at = CURRENT_TIMESTAMP 
			WHERE id = ?`,
			encryptedChoseEncrypt, encryptedKey, r.id)
		if err != nil {
			fmt.Printf("[-] 更新记录失败(id=%d): %v\n", r.id, err)
			skipCount++
			continue
		}
		updatedCount++
	}

	if err := tx.Commit(); err != nil {
		fmt.Printf("[-] 提交事务失败: %v\n", err)
		return 0, len(batch)
	}
	return updatedCount, skipCount
}

// decryptField 辅助函数：解密单个字段
// 参数：
// - encryptedStr: 加密的字符串
// - key: AES密钥
// 返回值：
// - 解密后的明文
// - 错误信息
func decryptField(encryptedStr string, key []byte) (string, error) {
	// 解析 JSON 格式的加密数据 {"iv":"...","data":"..."}
	var encryptedData struct {
		IV   string `json:"iv"`
		Data string `json:"data"`
	}

	// 解析 JSON
	if err := json.Unmarshal([]byte(encryptedStr), &encryptedData); err != nil {
		return "", fmt.Errorf("解析加密数据JSON失败: %w", err)
	}

	// 解析 IV 和加密数据
	nonce, err := base64.StdEncoding.DecodeString(encryptedData.IV)
	if err != nil {
		return "", fmt.Errorf("解析 IV 失败: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData.Data)
	if err != nil {
		return "", fmt.Errorf("解析密文失败: %w", err)
	}

	// 使用现有的AES解密函数
	plaintext, err := encryption.AesDecryptData(nonce, ciphertext, key)
	if err != nil {
		return "", fmt.Errorf("AES解密失败: %w", err)
	}
	fmt.Printf("[+] 解密成功: %s\n", string(plaintext))

	return string(plaintext), nil
}

// encryptField 辅助函数：加密单个字段
// 参数：
// - plaintext: 明文
// - key: AES密钥
// 返回值：
// - 加密后的JSON字符串
// - 错误信息
func encryptField(plaintext string, key []byte) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// 使用现有的AES加密函数
	nonce, ciphertext, err := encryption.AesEncryptData([]byte(plaintext), key)
	if err != nil {
		return "", fmt.Errorf("AES加密失败: %w", err)
	}

	// 将 IV 和密文转换为 Base64
	ivBase64 := base64.StdEncoding.EncodeToString(nonce)
	ciphertextBase64 := base64.StdEncoding.EncodeToString(ciphertext)

	// 构造符合系统格式的JSON数据
	encryptedData := struct {
		IV   string `json:"iv"`
		Data string `json:"data"`
	}{IV: ivBase64, Data: ciphertextBase64}

	jsonData, err := json.Marshal(encryptedData)
	if err != nil {
		return "", fmt.Errorf("序列化加密数据失败: %w", err)
	}

	return string(jsonData), nil
}

// writeErrorResponse 辅助函数：写入错误响应
// 参数：
// - w: HTTP响应写入器
// - message: 错误消息
// - err: 错误对象
func writeErrorResponse(w http.ResponseWriter, message string, err error) {
	fullMessage := message + ": " + err.Error()
	w.WriteHeader(http.StatusInternalServerError)
	response := respMessage.CommonResponse{
		Code:    http.StatusInternalServerError,
		Message: fullMessage,
		Data:    nil,
	}
	json.NewEncoder(w).Encode(response)
}

func ensureTotpColumnsExist(db *sql.DB, table string, totpColumns []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	exists, err := dbpkg.ColumnsExist(db, table, totpColumns)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("数据库列 %s 存在性检查失败: %v", "totp_enabled", err)
	}
	if len(exists) > 0 {
		if err := dbpkg.AddColumns(db, table, exists); err != nil {
			tx.Rollback()
			return fmt.Errorf("添加列 %s 失败: %v", "totp_enabled", err)
		}
	}
	err = tx.Commit() // 成功，提交事务
	if err != nil {
		tx.Rollback() // 提交失败，回滚
		return err
	}
	return nil
}

// 初始化同步用户配置
func initSyncUserConfig(r *http.Request, username string) error {
	// 从数据库获取用户个性化配置
	db, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}
	defer db.Close()
	var alertMailSetsBytes []byte
	if err := db.QueryRow(`SELECT alertMailSets FROM userMeta_data WHERE username = ?`, username).Scan(&alertMailSetsBytes); err != nil {
		return fmt.Errorf("查询失败: %w", err)
	}

	var storage AlertMailStorage
	if err := json.Unmarshal(alertMailSetsBytes, &storage); err != nil {
		return fmt.Errorf("解析 StorageData 失败: %w", err)
	}

	var mailCfg mail.GlobalConfig
	var dataBytes []byte

	// 判断是否加密
	if storage.EncryptSecret {
		encryptedBytes, err := base64.StdEncoding.DecodeString(string(storage.Data[1 : len(storage.Data)-1])) // 去掉引号
		if err != nil {
			return fmt.Errorf("base64解码失败: %w", err)
		}
		decrypted, err := encryption.DpapiDecrypt(encryptedBytes, encryption.DpapiUserScope)
		if err != nil {
			return fmt.Errorf("解密失败: %w", err)
		}
		dataBytes = decrypted
	} else {
		dataBytes = storage.Data
	}

	// 解析配置
	if err := json.Unmarshal(dataBytes, &mailCfg); err != nil {
		return fmt.Errorf("解析邮件配置失败: %w", err)
	}
	// 初始化邮件配置
	globalMail.SetGlobalConfig(mailCfg)

	// 加载白名单配置
	if err := loadWhitelistConfig(username); err != nil {
		return fmt.Errorf("加载白名单配置失败: %w", err)
	}

	//获取数据
	return nil
}

func loadWhitelistConfig(username string) error {
	// 从数据库获取白名单配置
	db, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		return fmt.Errorf("数据库异常: %w", err)
	}
	defer db.Close()
	exists, err := dbpkg.TableExists(db, "globalWhiteList")
	if err != nil {
		return fmt.Errorf("数据库异常: %w", err)
	}
	if !exists {
		return fmt.Errorf("表不存在！")
	}
	var (
		whiteList string
	)
	if err := db.QueryRow(`SELECT whitelist_enabled,whitelist_ips,whitelist_action FROM globalWhiteList WHERE username = ?`, username).Scan(&whitelistCfg.Enabled, &whiteList, &whitelistCfg.ActionOutsideWhitelist); err != nil {
		return fmt.Errorf("查询失败: %w", err)
	}
	whitelistCfg.Whitelist = strings.Split(whiteList, ",")
	return nil
}

// 清除已同步的用户配置、白名单配置、全局邮件配置、登录状态
func initSyncUserConfigCleanup(username string) {
	statusMux.Lock()
	defer statusMux.Unlock()
	status, ok := loginStatusList[username]
	if !ok || status == nil {
		return
	}
	if status.StopClearCh != nil {
		close(status.StopClearCh)
	}
	status.isPwdAuthed = false
	status.StopClearCh = nil
	loginStatusList[username] = status
	// 清除whiteListcfg
	whitelistCfg.Enabled = false
	whitelistCfg.Whitelist = []string{}
	whitelistCfg.ActionOutsideWhitelist = ""
	// 清除globalMailCfg
	globalMail.ClearGlobalConfig()
}
