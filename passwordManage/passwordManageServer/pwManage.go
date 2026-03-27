/*
 * @Author: 小鱼
 * @Date: 2025-07-03 15:27:13
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-10-30 15:15:28
 * @FilePath: \passwordManageServer\pwManage.go
 * @Description:
 *
 * Copyright (c) 2025 by 小鱼, All Rights Reserved.
 */
package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"xyrTools/passwordManage/passwordManageServer/pkg/dataHandler"
	"xyrTools/passwordManage/passwordManageServer/pkg/dataStructs"
	dbpkg "xyrTools/passwordManage/passwordManageServer/pkg/db"
	"xyrTools/passwordManage/passwordManageServer/pkg/encryption"
	"xyrTools/passwordManage/passwordManageServer/pkg/mail"
	"xyrTools/passwordManage/passwordManageServer/pkg/middleware"
	"xyrTools/passwordManage/passwordManageServer/pkg/otherFunc"
	respMessage "xyrTools/passwordManage/passwordManageServer/pkg/responseStd"
	"xyrTools/passwordManage/passwordManageServer/pkg/tools"

	"github.com/h2non/filetype"
)

// register 处理用户注册请求
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func register(w http.ResponseWriter, r *http.Request) {

	// 多用户未完善，临时改单用户
	// 打开 SQLite 数据库查询用户数量
	sqliteDb, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库连接失败")
		return
	}
	defer sqliteDb.Close()

	var userCount int
	err = sqliteDb.QueryRow("SELECT COUNT(*) FROM userMeta_data").Scan(&userCount)
	if err != nil && err.Error() != "no such table: userMeta_data" {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库查询失败")
		return
	}

	// 检查是否已有用户
	if userCount > 0 {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "已有用户注册，目前单用户模式下不允许再次注册")
		return
	}
	// 上为单用户临时加限制，完善后删除

	// 处理数据，解析json获取iv和密文，通过iv和协商的密钥tmpSessionAESKey解密密文
	// 解析请求体中的 JSON 数据
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "读取注册内容失败")
		return
	}
	defer r.Body.Close()

	// 将 JSON 数据解析到结构体
	var encryptedRequestData dataStructs.EncryptedRequestData
	if err = json.Unmarshal(body, &encryptedRequestData); err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "解析加密注册信息失败")
		return
	}

	iv := encryptedRequestData.IV
	registerDataCrypted := encryptedRequestData.EncryptedData

	// 解密数据
	decryptedData, err := encryption.AesDecryptData(iv, registerDataCrypted, tmpSessionAESKey)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "解密注册信息失败")
		return
	}
	// 解析注册数据
	var decryptedRegisterData dataStructs.DecryptedRegisterData
	if err = json.Unmarshal(decryptedData, &decryptedRegisterData); err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "解析 JSON 注册数据失败")
		return
	}

	// 生成rsaKey
	// 生成 RSA 密钥对
	rsaPrivateKey, rsaPublicKey, err := encryption.GenerateRsaKeyPair_BaseRand()
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "生成密钥对失败")
		return
	}

	// 使用公钥和 customInfo 生成 AES 密钥
	aesKey, err := encryption.GenerateAESKey_BaseInfo(rsaPublicKey, decryptedRegisterData.CustomInfo)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "生成 AES 密钥失败")
		return
	}
	// 输出生成的 AES 密钥
	//fmt.Printf("生成的 AES 密钥: %x\n", aesKey)

	userName := decryptedRegisterData.Username
	userPasswd := decryptedRegisterData.Password
	userPasswdHash, _ := otherFunc.GeneratePasswordHash(userPasswd)
	userInfo := decryptedRegisterData.CustomInfo
	rsaPrivKeyPemStr := encryption.PrivateKeyToPEM_Str(rsaPrivateKey)
	rsaPublicKeyPemStr := encryption.PublicKeyToPEM_Str(rsaPublicKey)
	aesKeyBase64 := base64.StdEncoding.EncodeToString(aesKey)

	// 需要加密保存到数据库的信息
	encryptedUserData := dataStructs.EncryptedUserData{
		UserInfo:   userInfo,
		UserPasswd: userPasswd,
		AESKey:     aesKey,
	}
	// 将结构体转换为 JSON 格式
	userMetaJsonData, err := json.Marshal(encryptedUserData)
	if err != nil {

		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据转换为 JSON 格式失败")
		return
	}
	// 加密用户重要数据

	nonce, userCryptedData, err := encryption.AesEncryptData(userMetaJsonData, aesKey)
	if err != nil {

		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "加密重要数据失败")
		return
	}
	// 保存数据到数据库

	err = dbpkg.SaveEncryptedUserDataToDB(sqliteDb, userName, userPasswdHash, rsaPublicKeyPemStr, nonce, userCryptedData)
	if err != nil {

		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "保存数据到数据库失败")
		return
	}

	// 生成pem文件用于返回客户端
	combinedPEM := fmt.Sprintf(
		"%s\n"+
			"%s\n"+
			"-----BEGIN AES KEY-----\n%s\n-----END AES KEY-----\n",
		rsaPrivKeyPemStr, rsaPublicKeyPemStr, aesKeyBase64)

	// 使用 AES 对 PEM 文件内容加密
	iv, encryptedPEMData, err := encryption.AesEncryptData([]byte(combinedPEM), tmpSessionAESKey)

	if err != nil {

		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "AES 加密 PEM 数据失败")
		return
	}

	// 构建响应数据
	encryptedResponse := struct {
		IV             string `json:"iv"`
		EncryptedPEM   string `json:"encryptedPEM"`
		RegisterStatus dataStructs.RegisterResponse
	}{
		IV:             base64.StdEncoding.EncodeToString(iv),               // 将 IV 编码为 Base64
		EncryptedPEM:   base64.StdEncoding.EncodeToString(encryptedPEMData), // 加密的 PEM 数据编码为 Base64
		RegisterStatus: dataStructs.RegisterResponse{Status: http.StatusOK, Message: "注册成功"},
	}

	respMessage.SendSuccessResponse(w, "Success", encryptedResponse)
}

// 登录处理
// login 处理用户登录请求
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func login(w http.ResponseWriter, r *http.Request) {
	ipInfo := r.Context().Value("ipInfo").(middleware.IPInfo)
	fmt.Printf("登录请求来自: %s (%v)\n", ipInfo.ClientIP, ipInfo.Chain)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "读取登录内容错误")
		return
	}
	defer r.Body.Close()

	var loginPwdRequest dataStructs.LoginPwdRequest
	if err = json.Unmarshal(body, &loginPwdRequest); err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "读取登录内容错误")
		return
	}

	// 解密数据
	decryptedData, err := encryption.AesDecryptData(loginPwdRequest.IV, loginPwdRequest.EncryptedData, tmpSessionAESKey)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "解密登录信息失败")
		return
	}

	var decryptedLoginData dataStructs.DecryptedLoginData
	if err = json.Unmarshal(decryptedData, &decryptedLoginData); err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "解析登录信息失败")
		return
	}

	sqliteDB, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "连接数据库失败")
		return
	}
	defer sqliteDB.Close()

	realPwdHash, err := dbpkg.GetPasswordHash(sqliteDB, decryptedLoginData.Username)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusOK, "用户名或密码错误")
		return
	}

	isTotpSetting := otherFunc.IsTotpSetting(sqliteDB, decryptedLoginData.Username)
	isLoginSuccess := otherFunc.VerifyPassword(realPwdHash, decryptedLoginData.Password)

	if !isLoginSuccess {
		respMessage.SendErrorResponse(w, http.StatusOK, "用户名或密码错误！")
		return
	}

	// 初始化用户配置
	err = initSyncUserConfig(r, decryptedLoginData.Username)
	if err != nil {
		fmt.Printf("[-] 初始化用户配置失败: %v\n", err)
	}
	// 更新登录状态为已认证密码
	mu.Lock()
	status, ok := loginStatusList[decryptedLoginData.Username]
	if !ok || status == nil {
		status = &loginStatus{}
		loginStatusList[decryptedLoginData.Username] = status
	}
	status.isPwdAuthed = true
	status.StopClearCh = make(chan struct{})
	mu.Unlock()
	if isTotpSetting {
		// 如果启用了totp，设置1分钟超时清除登录状态
		go func() {
			userStatus := loginStatusList[decryptedLoginData.Username]
			select {
			case <-time.After(time.Minute): // 超时时间
				initSyncUserConfigCleanup(decryptedLoginData.Username)
				return
			case <-userStatus.StopClearCh:
				return
			}
		}()

		cookie := http.Cookie{
			Name:   "sessionId",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		}
		http.SetCookie(w, &cookie)
		respMessage.SendSuccessResponse(w, "TOTP", nil)
		return
	}

	// 白名单和session处理逻辑
	if whitelistCfg.Enabled {
		isInWhitelist, err := tools.IsIPInWhitelist(ipInfo.ClientIP, whitelistCfg.Whitelist)
		if err != nil {
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "检查白名单失败")
			return
		}
		if !isInWhitelist {
			if whitelistCfg.ActionOutsideWhitelist == "block" {
				cookie := http.Cookie{
					Name:   "sessionId",
					Value:  "",
					Path:   "/",
					MaxAge: -1,
				}
				http.SetCookie(w, &cookie)
				respMessage.SendErrorResponse(w, http.StatusForbidden, "非法请求！")
				return
			} else if whitelistCfg.ActionOutsideWhitelist == "alert" {
				mailBody := fmt.Sprintf("%s用户 %s 登录成功，登录IP为 %s 不在白名单中！",
					time.Now().Format("2006-01-02 15:04:05"), decryptedLoginData.Username, ipInfo.ClientIP)
				go func() {
					err = globalMail.Email(globalMail.MailConfig.To, "登录告警", mailBody)
					if err != nil {
						fmt.Printf("[-] 发送登录告警邮件失败: %v\n", err)
					}
				}()
			}
		}
	}

	// 会话处理
	valid, checkErr := sessionStore.CheckSessionValid(r, w)
	if checkErr != nil {
		goto newSession
	}
	if valid {
		userSession, sessionErr := sessionStore.Get(r, "sessionId")
		if sessionErr != nil || userSession == nil {
			goto newSession
		}
		sessionStore.Save(r, w, userSession)
		respMessage.SendSuccessResponse(w, "ok", nil)
		return
	}

newSession:
	usrSession, err := sessionStore.New(r, "sessionId")
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "创建会话失败")
		return
	}
	usrSession.Values["username"] = decryptedLoginData.Username
	usrSession.Values["loginTime"] = time.Now()
	if err = sessionStore.Save(r, w, usrSession); err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "创建会话失败")
		return
	}
	respMessage.SendSuccessResponse(w, "Success", nil)
}

// logout 处理用户退出登录
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func logout(w http.ResponseWriter, r *http.Request) {
	// 从请求上下文中获取用户名
	cookie, err := r.Cookie("sessionId")
	if err != nil {
		cookie = &http.Cookie{
			Name:   "sessionId", // Cookie 名称
			Value:  "",          // 清空 Cookie 值
			Path:   "/",         // 设置 Cookie 路径
			MaxAge: -1,          // 设置 MaxAge 为 -1 让浏览器删除 Cookie
		}
		http.SetCookie(w, cookie)
		respMessage.SendErrorResponse(w, http.StatusUnauthorized, "session失效！")
		return
	}
	sessionInfo, err := sessionStore.GetSessionInfo(r)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "获取session信息失败")
		return
	}
	//fmt.Println(sessionInfo)

	// 清除登录标志
	initSyncUserConfigCleanup(sessionInfo["username"].(string))
	// 清除session
	sessionStore.DeleteSession(cookie.Value)
	// 清除密钥信息
	userAesKeyHex = ""
	// 清除cookie
	cookie = &http.Cookie{
		Name:   "sessionId", // Cookie 名称
		Value:  "",          // 清空 Cookie 值
		Path:   "/",         // 设置 Cookie 路径
		MaxAge: -1,          // 设置 MaxAge 为 -1 让浏览器删除 Cookie
	}
	http.SetCookie(w, cookie)
	// 设置Clear-Site-Data头，清除所有站点数据
	w.Header().Set("Clear-Site-Data", "*")
	//返回成功
	respMessage.SendSuccessResponse(w, "ok", nil)
}

// saveSecret 保存加密数据
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func saveSecret(w http.ResponseWriter, r *http.Request) {
	// 限制20MB
	defer r.Body.Close()
	const maxBodySize = 20 * 1024 * 1024 // 10MB
	if r.ContentLength > maxBodySize {
		respMessage.SendErrorResponse(w, http.StatusRequestEntityTooLarge, "请求数据最大20M！")
		return
	}

	body, err := io.ReadAll(r.Body)
	if len(body) == 0 {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "请求数据为空！")
		return
	}
	if r.Method == "POST" {
		// 读取请求体
		if err != nil {
			respMessage.SendErrorResponse(w, http.StatusBadRequest, "无法读取请求体")
			return
		}
		// 打印请求体中的数据
		//fmt.Println("接收到的数据:", string(body))
		var encryptedRequestData dataStructs.EncryptedRequestData
		err = json.Unmarshal(body, &encryptedRequestData)
		if err != nil {
			respMessage.SendErrorResponse(w, http.StatusBadRequest, "无法解析请求体中的 JSON 数据")
			return
		}
		//fmt.Println("IV:", encryptedRequestData.IV)
		//fmt.Println("Data:", encryptedRequestData.EncryptedData)
		//keystr := "c81ca881b2340b509b23258066e67056f0e07ff54ecfb4dedf5b20c0ae5a70a1"
		//aesKeyTest, err := hex.DecodeString(keystr)
		// if err != nil {
		// 	http.Error(w, "密钥格式错误：", http.StatusBadRequest)
		// 	return
		// }
		addData, err := encryption.AesDecryptData(encryptedRequestData.IV, encryptedRequestData.EncryptedData, tmpSessionAESKey)
		if err != nil {
			respMessage.SendErrorResponse(w, http.StatusBadRequest, "服务端解密失败")
			return
		}
		//fmt.Println("解密后的数据:", string(addData))

		//判断传过来的数据中，是否存在Id，有则更新，无则新增
		type upDateId struct {
			ID json.RawMessage `json:"id,omitempty"` // 仅解析id字段
		}
		var isUpdate bool = false
		var dataId int
		var updateID upDateId
		if err = json.Unmarshal(addData, &updateID); err == nil && len(updateID.ID) > 0 {
			// ID字段存在且能正确解析时视为更新
			if err = json.Unmarshal(updateID.ID, &dataId); err == nil {
				isUpdate = true
			}
		}
		var input dataHandler.InputData
		// 带 keyFile 的数据
		if err = json.Unmarshal(addData, &input); err != nil {
			respMessage.SendErrorResponse(w, http.StatusBadRequest, "解析数据失败")
			return
		}
		//dataHandler.AnalyzeDecryptedDataPrint(input)

		if isUpdate {
			if err = dbpkg.UpdateData(input, cfg.Database.SQLite.SqliteDbPath, dataId); err != nil {
				respMessage.SendErrorResponse(w, http.StatusInternalServerError, "更新数据出错")
				return
			}
		} else {
			// 保存新增数据
			if err = dbpkg.SaveToDatabase(input, cfg.Database.SQLite.SqliteDbPath); err != nil {
				respMessage.SendErrorResponse(w, http.StatusInternalServerError, "保存数据出错")
				return
			}
		}

		_, err = tools.AnalyzePasswordStrength(cfg.Database.SQLite.SqliteDbPath, "input_data", userAesKeyHex)
		if err != nil {
			//log.Fatalf("分析密钥强度出错: %v", err)
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据已保存，但服务端无主密钥，分析密钥强度失败")
			return
		}
		respMessage.SendSuccessResponse(w, "Success", nil)

	} else {
		respMessage.SendErrorResponse(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}
}

// aes临时密钥协商
// getAesKey 获取AES会话密钥，通过ECC密钥交换生成共享密钥
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func getAesKey(w http.ResponseWriter, r *http.Request) {
	// ECC密钥交换
	// 定义接收客户端公钥结构体
	var request struct {
		PublicKeyBase64 string `json:"clientPublicKeyBase64"`
	}
	// 解析请求体内容
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "请求异常！")
		return
	}
	// 提取接收到的客户端公钥
	clientPublicKeyBase64 := request.PublicKeyBase64
	if clientPublicKeyBase64 == "" {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "请求异常！")
		return
	}
	// 导入客户端公钥，生成公钥对象
	clientPublicKey, err := encryption.ImportEccPublicKey(clientPublicKeyBase64)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "服务端异常！")
		return
	}
	// 生成服务器密钥对
	serverPrivateKey, serverPpublicKey, err := encryption.GenerateEccKeyPair()
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "服务端异常！")
		return
	}
	// 导出服务器公钥并进行 Base64 编码
	serverPublicKeyBase64, err := encryption.ExportEccPublicKey(serverPpublicKey)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "服务端异常！")
		return
	}

	// 计算共享秘密
	sharedSecret, err := deriveSharedSecret(serverPrivateKey, clientPublicKey)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	// 将共享密钥转换为十六进制字符串打印（调试用）
	//hexSecret := hex.EncodeToString(sharedSecret)
	//fmt.Println("共享密钥 (Hex):", hexSecret)
	// 共享密钥生成aes密钥
	aesKey, err := deriveAESKey(sharedSecret)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	tmpSessionAESKey = aesKey
	// 返回服务器的公钥
	respMessage.SendSuccessResponse(w, "ok", serverPublicKeyBase64)
	// 计算aesKey的hash并打印（调试对比用）
	//hash := sha256.Sum256(aesKey)
	//fmt.Printf("AES Key Hash: %x\n", hash)
	// 打印 AES 会话密钥（仅用于调试）
	//fmt.Printf("Derived AES Key: %x\n", aesKey)
}

// checkSession 检查会话是否有效
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func checkSession(w http.ResponseWriter, r *http.Request) {
	// 从请求上下文中获取用户名
	username := r.Context().Value("username").(string)
	data := map[string]interface{}{
		"username": username,
		"isLogin":  true,
	}
	respMessage.SendSuccessResponse(w, "ok", data)
}

// queryData 查询密码数据
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func queryData(w http.ResponseWriter, r *http.Request) {
	// 解析基础分页参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 获取查询参数
	queryParams := r.URL.Query()
	appName := queryParams.Get("appName")
	userName := queryParams.Get("userName")
	urlParam := queryParams.Get("url")
	tags := queryParams.Get("tags")

	// 初始化数据库连接
	dbQuery, dbErr := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if dbErr != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库连接失败！")
		return
	}
	defer dbQuery.Close()

	// 动态构建查询条件
	baseQuery := `
        SELECT
            id, app_name, is_app_name_encrypted, username, is_username_encrypted,
            input_type, password, key_file, url, is_url_encrypted,
            notes, is_notes_encrypted, tags, is_tags_encrypted, strength,
            chose_encrypt, key, count
        FROM input_data`

	countQuery := "SELECT COUNT(*) FROM input_data"
	conditions := []string{}
	args := []any{}
	// 添加应用名查询条件
	if appName != "" {
		conditions = append(conditions, "app_name LIKE ? AND is_app_name_encrypted = 0")
		args = append(args, "%"+appName+"%")
	}

	// 添加用户名查询条件
	if userName != "" {
		conditions = append(conditions, "username LIKE ? AND is_username_encrypted = 0")
		args = append(args, "%"+userName+"%")
	}

	// 添加 URL 查询条件
	if urlParam != "" {
		conditions = append(conditions, "url LIKE ? AND is_url_encrypted = 0")
		args = append(args, "%"+urlParam+"%")
	}

	// 处理标签查询（多标签 OR 关系）
	if tags != "" {
		tagList := strings.Split(tags, ",")
		tagConditions := []string{}
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagConditions = append(tagConditions, "tags LIKE ? AND is_tags_encrypted = 0")
				args = append(args, "%"+tag+"%")
			}
		}
		if len(tagConditions) > 0 {
			conditions = append(conditions, "("+strings.Join(tagConditions, " OR ")+")")
		}
	}

	// 合并 WHERE 条件
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// 最终查询语句
	fullQuery := baseQuery + whereClause + " ORDER BY ROWID ASC LIMIT ? OFFSET ?"

	// 添加分页参数
	queryArgs := append(args, pageSize, offset)

	// 执行分页查询
	rows, err := dbQuery.Query(fullQuery, queryArgs...)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "查询失败！")
		return
	}
	defer rows.Close()

	// 处理查询结果
	var data []dataHandler.QueryData
	for rows.Next() {
		var record dataHandler.QueryData
		var keyFile sql.NullString

		err = rows.Scan(
			&record.Id, &record.AppName.Raw, &record.IsAppNameEncrypted,
			&record.Username.Raw, &record.IsUsernameEncrypted, &record.InputType,
			&record.Password.Raw, &keyFile, &record.URL.Raw, &record.IsUrlEncrypted,
			&record.Notes.Raw, &record.IsNotesEncrypted, &record.Tags.Raw,
			&record.IsTagsEncrypted, &record.Strength, &record.ChoseEncrypt.Raw,
			&record.Key.Raw, &record.Count,
		)
		if err != nil {
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据解析失败！")
			return
		}

		if keyFile.Valid {
			//record.KeyFile.File = &keyFile.String
		} else {
			record.KeyFile.File = nil
		}
		data = append(data, record)
	}

	// 查询带条件的总记录数
	var total int
	err = dbQuery.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "查询失败！")
		return
	}

	// 返回JSON响应
	response := struct {
		Data  []dataHandler.QueryData `json:"data"`
		Total int                     `json:"total"`
	}{
		Data:  data,
		Total: total,
	}
	respMessage.SendSuccessResponse(w, "Success", response)
}

// deleteRecords 删除密码记录
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func deleteRecords(w http.ResponseWriter, r *http.Request) {
	// 从请求上下文中获取用户名
	//username := r.Context().Value("username").(string)
	// 解析请求体中的 JSON 数据
	var req struct {
		Ids []int `json:"ids"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "解析请求出错")
		return
	}

	// 检查是否有要删除的记录
	if len(req.Ids) == 0 {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "请求数据异常！")
		return
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库连接失败！")
		return
	}
	defer db.Close()

	// 构建 SQL 删除语句
	placeholders := make([]string, len(req.Ids))
	for i := range req.Ids {
		placeholders[i] = "?"
	}
	query := fmt.Sprintf("DELETE FROM input_data WHERE id IN (%s)", strings.Join(placeholders, ","))

	// 将 []int 转换为 []any
	args := make([]any, len(req.Ids))
	for i, id := range req.Ids {
		args[i] = id
	}
	// 执行删除操作
	_, err = db.Exec(query, args...)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "删除失败！")
		return
	}
	// 返回成功响应
	respMessage.SendSuccessResponse(w, "Success", nil)
}

// setting 处理系统设置保存
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func setting(w http.ResponseWriter, r *http.Request) {
	// 从请求上下文中获取用户名
	//username := r.Context().Value("username").(string)
	// 获取请求体内容
	var settingRequest dataStructs.SettingRequest
	if err := json.NewDecoder(r.Body).Decode(&settingRequest); err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "错误的请求体！")
		return
	}

	//解析配置
	switch settingRequest.SetType {
	// 系统名称
	case "systemName":
		var systemNameSettings dataStructs.SystemNameSettings
		if err := json.Unmarshal(settingRequest.Data, &systemNameSettings); err != nil {
			respMessage.SendErrorResponse(w, http.StatusBadRequest, "数据解析失败！")
			return
		}
		tmpRunCfg.SystemName = systemNameSettings.SysName
		respMessage.SendSuccessResponse(w, "Success", tmpRunCfg.SystemName)
		return

	// 系统安全
	case "security":
		var securitySettings dataStructs.SecuritySettings
		if err := json.Unmarshal(settingRequest.Data, &securitySettings); err != nil {
			respMessage.SendErrorResponse(w, http.StatusBadRequest, "系统安全解析失败！")
			return
		}
		tmpRunCfg.Timeout = securitySettings.Timeout
		cfg.Session.MaxAge = securitySettings.Timeout
		sessionStore.SetMaxAge(securitySettings.Timeout)

		sessionStore.UpdateSession(r, w, "sessionId")
		respMessage.SendSuccessResponse(w, "Success", tmpRunCfg.Timeout)
		return
	// 天气API
	case "weatherAPI":
		var weatherSettings dataStructs.WeatherAPISettings
		if err := json.Unmarshal(settingRequest.Data, &weatherSettings); err != nil {
			respMessage.SendErrorResponse(w, http.StatusBadRequest, "天气API解析失败！")
			return
		}
		tmpRunCfg.WeatherApiKey = weatherSettings.APIKey
		respMessage.SendSuccessResponse(w, "Success", tmpRunCfg.WeatherApiKey)
	default:
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "未知的设置类型！")
		return
	}

	// 返回成功响应
	respMessage.SendSuccessResponse(w, "Success", nil)
}

// saveKeyToCS 保存密钥到服务端
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func saveKeyToCS(w http.ResponseWriter, r *http.Request) {
	// 只允许 POST 请求
	if r.Method != http.MethodPost {
		respMessage.SendErrorResponse(w, http.StatusMethodNotAllowed, "只允许 POST 请求")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	// 定义接收数据的结构体
	type KeyData struct {
		AESKey       string `json:"aesKey"`
		AESKeyBase64 string `json:"aesKeyBase64"`
	}
	var keyData KeyData
	// 解析 JSON 数据
	err = json.Unmarshal(body, &keyData)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "解析数据失败")
		return
	}

	// 验证数据有效性
	if keyData.AESKey == "" || keyData.AESKeyBase64 == "" {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "密钥数据不能为空")
		return
	}

	// 为内存中的密钥赋值
	userAesKeyHex = keyData.AESKey
	//userAesKeyBase64 = keyData.AESKeyBase64

	// 返回
	respMessage.SendSuccessResponse(w, "Success", "密钥同步成功")

}

// 清除服务端密钥配置
// clearKey 清除服务端缓存的AES密钥
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func clearKey(w http.ResponseWriter, r *http.Request) {
	// 只允许 GET 请求
	if r.Method != http.MethodGet {
		respMessage.SendErrorResponse(w, http.StatusMethodNotAllowed, "只允许 GET 请求")
		return
	}

	// 清除内存中的密钥
	userAesKeyHex = ""

	// 返回成功响应
	response := map[string]interface{}{
		"result":  "success",
		"message": "服务端密钥配置已清除",
	}

	respMessage.SendSuccessResponse(w, "Success", response)
}

// 更新统计数据
// updateStats 更新密码统计信息
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func updateStats(w http.ResponseWriter, r *http.Request) {
	// 只允许GET请求
	if r.Method != http.MethodGet {
		respMessage.SendErrorResponse(w, http.StatusMethodNotAllowed, "只允许 GET 请求")
		return
	}
	if userAesKeyHex == "" {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "服务端密钥为空，功能不可用！")
		return
	}
	responseData, err := tools.AnalyzePasswordStrength(cfg.Database.SQLite.SqliteDbPath, "input_data", userAesKeyHex)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "统计数据更新失败")
		return
	} else {
		respMessage.SendSuccessResponse(w, "Success", responseData)
	}
}

// 下载密钥文件
// downloadKeyFile 下载密钥文件
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func downloadKeyFile(w http.ResponseWriter, r *http.Request) {
	// 只允许GET请求
	if r.Method != http.MethodGet {
		respMessage.SendErrorResponse(w, http.StatusMethodNotAllowed, "只允许 GET 请求")
		return
	}
	// 获取get请求中的id参数
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "缺少id参数")
		return
	}
	// 将id转换为整数
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadGateway, "id参数格式错误")
		return
	}

	// 初始化数据库连接
	db, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadGateway, "数据库连接失败")
		return
	}
	defer db.Close()

	// 查询key_file、chose_encrypt和key字段值
	var keyFile, choseEncrypt, key sql.NullString
	err = db.QueryRow("SELECT key_file, chose_encrypt, key FROM input_data WHERE id = ?", id).Scan(&keyFile, &choseEncrypt, &key)
	if err != nil {
		if err == sql.ErrNoRows {
			respMessage.SendErrorResponse(w, http.StatusBadGateway, "未找到对应记录")
			return
		} else {
			respMessage.SendErrorResponse(w, http.StatusBadGateway, "查询失败")
			return
		}
	}

	// 检查key_file是否有效
	if !keyFile.Valid {
		respMessage.SendErrorResponse(w, http.StatusBadGateway, "key_file值为空")
		return
	}

	// 准备返回的数据结构，包含解密所需的全部字段
	result := map[string]string{
		"key_file":      keyFile.String,
		"chose_encrypt": choseEncrypt.String, // 即使为空也返回
		"key":           key.String,          // 即使为空也返回
	}
	// 返回数据
	respMessage.SendSuccessResponse(w, "Success", result)
}

// 更新密码
// updatePassword 更新用户密码
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func updatePassword(w http.ResponseWriter, r *http.Request) {
	// 只允许POST请求
	if r.Method != http.MethodPost {
		respMessage.SendErrorResponse(w, http.StatusMethodNotAllowed, "只允许 POST 请求")
		return
	}
	// 检查主密钥是否配置
	if userAesKeyHex == "" {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "主密钥未配置！")
		return
	}
	// 读取请求体数据
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "读取请求体失败")
		return
	}
	defer r.Body.Close()
	// 将 JSON 数据解析到结构体
	var encryptedRequestData dataStructs.EncryptedRequestData
	if err = json.Unmarshal(body, &encryptedRequestData); err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("请求数据异常：%v", err))
		return
	}

	iv := encryptedRequestData.IV
	registerDataCrypted := encryptedRequestData.EncryptedData

	// 解密数据
	decryptedData, err := encryption.AesDecryptData(iv, registerDataCrypted, tmpSessionAESKey)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("请求数据异常： %v", err))
		return
	}

	// 解析请求体数据
	var updatePwdData struct {
		Username    string `json:"username"`
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err = json.Unmarshal(decryptedData, &updatePwdData); err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}
	// 添加非空检查
	if updatePwdData.Username == "" {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "用户名不能为空")
		return
	}
	if updatePwdData.OldPassword == "" {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "旧密码不能为空")
		return
	}
	if updatePwdData.NewPassword == "" {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "新密码不能为空")
		return
	}
	// 从数据库获取hash
	sqliteDB, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库连接失败")
		return
	}
	defer sqliteDB.Close()
	realPwdHash, err := dbpkg.GetPasswordHash(sqliteDB, updatePwdData.Username)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库查询失败")
		return
	}

	// 旧密码验证
	isLoginSuccess := otherFunc.VerifyPassword(realPwdHash, updatePwdData.OldPassword)
	if !isLoginSuccess {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "校验失败！")
		return
	}
	// 生成新密码hash
	newPwdHash, err := otherFunc.GeneratePasswordHash(updatePwdData.NewPassword)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "服务端异常")
		return
	}
	var userMeta dbpkg.UserMeta
	// 查询数据库获取data
	userMeta, err = dbpkg.GetUserMeta(sqliteDB, updatePwdData.Username)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库查询失败")
		return
	}

	// 十六进制密钥转为byte
	userAesKeyByte, err := hex.DecodeString(userAesKeyHex)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "服务端异常")
		return
	}

	// 解密原数据
	decryptedData, err = encryption.AesDecryptData(userMeta.Nonce, userMeta.Ciphertext, userAesKeyByte)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "服务端异常")
		return
	}
	var encryptedUserData dataStructs.EncryptedUserData
	if err = json.Unmarshal(decryptedData, &encryptedUserData); err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "服务端异常")
		return
	}
	// 改密
	encryptedUserData.UserPasswd = updatePwdData.NewPassword
	userMetaJsonData, err := json.Marshal(encryptedUserData)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "服务端异常")
		return
	}

	// 重新加密数据
	nonce, reEncryptedData, err := encryption.AesEncryptData(userMetaJsonData, userAesKeyByte)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "服务端异常")
		return
	}
	// 更新数据库
	err = dbpkg.UpdateRegisterInfo(sqliteDB, updatePwdData.Username, newPwdHash, userMeta.PublicKey, nonce, reEncryptedData)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "服务端异常")
		return
	}
	// 改密成功
	respMessage.SendSuccessResponse(w, "Success", "密码更新成功")

}

// 获取临时配置，前端初始化部分设置
// getTmpCfg 获取临时配置信息
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func getTmpCfg(w http.ResponseWriter, r *http.Request) {
	// 只允许GET请求
	if r.Method != http.MethodGet {
		respMessage.SendErrorResponse(w, http.StatusMethodNotAllowed, "只允许 GET 请求")
		return
	}
	// 返回tmpRunCfg
	respMessage.SendSuccessResponse(w, "Success", tmpRunCfg)
}

// changeMasterKey 修改用户主密钥
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func changeMasterKey(w http.ResponseWriter, r *http.Request) {
	// 从请求上下文中获取用户名
	//username := r.Context().Value("username").(string)
	// 只允许POST请求
	if r.Method != http.MethodPost {
		respMessage.SendErrorResponse(w, http.StatusMethodNotAllowed, "只允许 POST 请求")
		return
	}

	// 限制请求大小
	defer r.Body.Close()
	const maxBodySize = 10 * 1024 * 102 // 10MB
	if r.ContentLength > maxBodySize {
		respMessage.SendErrorResponse(w, http.StatusRequestEntityTooLarge, "请求数据过大！")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if len(body) == 0 {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "请求数据为空！")
		return
	}
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "无法读取请求体："+err.Error())
		return
	}

	// 解析加密的请求数据
	var encryptedRequestData dataStructs.EncryptedRequestData
	err = json.Unmarshal(body, &encryptedRequestData)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "无法解析请求体中的 JSON 数据："+err.Error())
		return
	}

	// 解密请求数据
	decryptedData, err := encryption.AesDecryptData(encryptedRequestData.IV, encryptedRequestData.EncryptedData, tmpSessionAESKey)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "解密请求数据失败："+err.Error())
		return
	}

	// 打印解密后的数据
	//fmt.Println("解密后的请求数据:", string(decryptedData))

	// 从解密后的数据中获取oldMasterKey
	var changeMasterKeyData struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		OldMasterKey string `json:"oldMasterKey"`
	}
	err = json.Unmarshal(decryptedData, &changeMasterKeyData)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "解析解密后的数据失败："+err.Error())
		return
	}

	oldMasterKey := changeMasterKeyData.OldMasterKey
	if oldMasterKey == "" {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "原主密钥不能为空")
		return
	}

	// 从数据库中获取数据，userMeta_data表中的nonce和ciphertext
	sqliteDb, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库连接失败："+err.Error())
		return
	}
	defer sqliteDb.Close()

	username := changeMasterKeyData.Username
	// 获取的用户元数据
	userMeta, err := dbpkg.GetUserMeta(sqliteDb, username)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "获取用户信息失败："+err.Error())
		return
	}

	// 尝试用传过来的oldMasterKey解密数据
	// 十六进制字符串转字节数组
	oldMasterKeyBytes, err := hex.DecodeString(oldMasterKey)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "密钥格式错误："+err.Error())
		return
	}
	decryptedUserData, err := encryption.AesDecryptData(userMeta.Nonce, userMeta.Ciphertext, oldMasterKeyBytes)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusUnauthorized, "原主密钥不正确，解密失败："+err.Error())
		return
	}

	//打印解密的数据
	//fmt.Println("解密后的用户数据:", string(decryptedUserData))
	var trustUserData dataStructs.EncryptedUserData
	err = json.Unmarshal(decryptedUserData, &trustUserData)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "解析解密后的数据失败："+err.Error())
		return
	}

	// 检查信息是否一致
	trustAESKeyHex := hex.EncodeToString(trustUserData.AESKey)
	if trustUserData.UserPasswd != changeMasterKeyData.Password || trustAESKeyHex != changeMasterKeyData.OldMasterKey {
		respMessage.SendErrorResponse(w, http.StatusUnauthorized, "信息有误，验证失败")
		return
	}

	// 生成 RSA 密钥对
	newRsaPrivateKey, newRsaPublicKey, err := encryption.GenerateRsaKeyPair_BaseRand()
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "生成密钥对失败")
		return
	}

	// 使用公钥和 customInfo 生成 AES 密钥
	newAesKey, err := encryption.GenerateAESKey_BaseInfo(newRsaPublicKey, trustUserData.UserInfo)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "生成 AES 密钥失败")
		return
	}
	// 输出生成的 AES 密钥
	//fmt.Printf("生成的 AES 密钥: %x\n", newAesKey)

	// 这里已经在前面定义了username变量，不需要重复定义
	// username := changeMasterKeyData.Username
	userPasswd := trustUserData.UserPasswd
	userPasswdHash, _ := otherFunc.GeneratePasswordHash(userPasswd)
	//userInfo := trustUserData.UserInfo
	rsaPrivKeyPemStr := encryption.PrivateKeyToPEM_Str(newRsaPrivateKey)
	rsaPublicKeyPemStr := encryption.PublicKeyToPEM_Str(newRsaPublicKey)
	aesKeyBase64 := base64.StdEncoding.EncodeToString(newAesKey)

	updateUserData := dataStructs.EncryptedUserData{
		UserInfo:   trustUserData.UserInfo,
		UserPasswd: trustUserData.UserPasswd,
		AESKey:     newAesKey,
	}
	// 将结构体转换为 JSON 格式
	userMetaJsonData, err := json.Marshal(updateUserData)
	if err != nil {

		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据转换为 JSON 格式失败")
		return
	}
	// 加密用户重要数据
	nonce, userCryptedData, err := encryption.AesEncryptData(userMetaJsonData, newAesKey)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "加密重要数据失败")
		return
	}

	// 更新数据库
	// 调用UpdateRegisterInfo函数更新数据库中的nonce和ciphertext
	err = dbpkg.UpdateRegisterInfo(sqliteDb, username, userPasswdHash, rsaPublicKeyPemStr, nonce, userCryptedData)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "更新数据库失败："+err.Error())
		return
	}

	// 生成pem文件用于返回客户端
	combinedPEM := fmt.Sprintf(
		"%s\n"+
			"%s\n"+
			"-----BEGIN AES KEY-----\n%s\n-----END AES KEY-----\n",
		rsaPrivKeyPemStr, rsaPublicKeyPemStr, aesKeyBase64)

	// 使用 AES 对 PEM 文件内容加密
	iv, encryptedPEMData, err := encryption.AesEncryptData([]byte(combinedPEM), tmpSessionAESKey)

	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "AES 加密 PEM 数据失败")
		return
	}
	go updateSavedData(newAesKey, oldMasterKeyBytes, cfg.Database.SQLite.SqliteDbPath)

	// 构建响应数据
	encryptedResponse := struct {
		IV             string `json:"iv"`
		EncryptedPEM   string `json:"encryptedPEM"`
		RegisterStatus dataStructs.RegisterResponse
	}{
		IV:             base64.StdEncoding.EncodeToString(iv),               // 将 IV 编码为 Base64
		EncryptedPEM:   base64.StdEncoding.EncodeToString(encryptedPEMData), // 加密的 PEM 数据编码为 Base64
		RegisterStatus: dataStructs.RegisterResponse{Status: http.StatusOK, Message: "主密钥更新成功"},
	}

	// 解密成功，返回成功响应
	respMessage.SendSuccessResponse(w, "原主密钥验证成功", encryptedResponse)
}

// getAvatarAndUser 获取用户名和头像信息
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func getAvatarAndUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 从请求上下文中获取用户名
	username := r.Context().Value("username").(string)

	// 连接数据库
	db, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库连接失败")
		return
	}
	defer db.Close()

	// 查询头像路径和大小
	var avatarPath string
	var avatarSize int64
	query := "SELECT avatar_path, avatar_size FROM userMeta_data WHERE username = ?"
	err = db.QueryRow(query, username).Scan(&avatarPath, &avatarSize)

	// 处理查询结果
	if err != nil {
		// 判断错误类型
		if err == sql.ErrNoRows {
			// 用户存在但没有头像记录，返回默认空头像
			w.WriteHeader(http.StatusOK)
			respMessage.SendSuccessResponse(w, "Success", map[string]interface{}{
				"username": username,
				"avatar":   "",
				"size":     0,
			})
			return
		} else if strings.Contains(err.Error(), "no such column") {
			// 字段不存在错误，返回默认空头像
			respMessage.SendSuccessResponse(w, "Success", map[string]interface{}{
				"username": username,
				"avatar":   "",
				"size":     0,
			})
			return
		}
		// 其他数据库错误
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "获取用户信息失败: "+err.Error())
		return
	}

	// 初始化头像Base64编码
	avatarBase64 := ""

	// 如果头像路径存在，读取文件并进行Base64编码
	if avatarPath != "" {
		// 读取图片文件
		imageData, err := os.ReadFile(avatarPath)
		if err != nil {
			// 文件读取失败时返回空头像，但仍返回成功状态
			log.Printf("读取头像文件失败: %v", err)
		} else {
			// 成功读取文件，进行Base64编码
			avatarBase64 = base64.StdEncoding.EncodeToString(imageData)
		}
	}

	respMessage.SendSuccessResponse(w, "Success", map[string]interface{}{
		"username": username,
		"avatar":   avatarBase64, // 返回Base64编码后的头像数据
		"size":     avatarSize,   // 返回头像大小
	})
}

// 更新头像
// updateAvatar 更新用户头像
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func updateAvatar(w http.ResponseWriter, r *http.Request) {
	// 设置响应类型为JSON
	w.Header().Set("Content-Type", "application/json")

	// 定义接收前端数据的结构体
	type AvatarRequestData struct {
		Filename string `json:"filename"`
		Size     int64  `json:"size"`
		Data     string `json:"data"` // Base64编码的图片数据
	}

	// 从请求上下文中获取用户名
	username := r.Context().Value("username").(string)

	// 解析前端传来的JSON数据
	var requestData AvatarRequestData
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := respMessage.CommonResponse{
			Code:    http.StatusBadRequest,
			Message: "解析请求数据失败: " + err.Error(),
			Data:    nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 验证数据
	if requestData.Filename == "" || requestData.Data == "" {
		w.WriteHeader(http.StatusBadRequest)
		response := respMessage.CommonResponse{
			Code:    http.StatusBadRequest,
			Message: "文件名或图片数据不能为空",
			Data:    nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 创建头像存储目录
	avatarDir := "avatars"
	err = os.MkdirAll(avatarDir, 0755)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := respMessage.CommonResponse{
			Code:    http.StatusInternalServerError,
			Message: "创建头像存储目录失败: " + err.Error(),
			Data:    nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 生成唯一文件名，避免覆盖
	fileExt := ""
	if dotIdx := strings.LastIndex(requestData.Filename, "."); dotIdx != -1 {
		fileExt = requestData.Filename[dotIdx:]
	}
	timestamp := time.Now().UnixNano()
	uniqueFilename := fmt.Sprintf("%s_%d%s", username, timestamp, fileExt)
	filePath := filepath.Join(avatarDir, uniqueFilename)

	// 解码Base64数据
	imageData, err := base64.StdEncoding.DecodeString(requestData.Data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := respMessage.CommonResponse{
			Code:    http.StatusBadRequest,
			Message: "解码图片数据失败: " + err.Error(),
			Data:    nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 使用filetype库检查文件类型
	kind, err := filetype.Match(imageData)
	if err != nil || kind == filetype.Unknown {
		w.WriteHeader(http.StatusBadRequest)
		response := respMessage.CommonResponse{
			Code:    http.StatusBadRequest,
			Message: "不允许的文件类型！",
			Data:    nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 只允许图片类型
	if !filetype.IsImage(imageData) {
		w.WriteHeader(http.StatusBadRequest)
		response := respMessage.CommonResponse{
			Code:    http.StatusBadRequest,
			Message: "不允许的文件类型！",
			Data:    nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 根据检测到的文件类型修正文件后缀
	correctExt := kind.Extension
	// 重新生成带正确后缀的文件名
	timestamp = time.Now().UnixNano()
	uniqueFilename = fmt.Sprintf("%s_%d.%s", username, timestamp, correctExt)
	filePath = filepath.Join(avatarDir, uniqueFilename)

	// 保存图片文件
	err = os.WriteFile(filePath, imageData, 0644)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := respMessage.CommonResponse{
			Code:    http.StatusInternalServerError,
			Message: "保存图片文件失败: " + err.Error(),
			Data:    nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 连接数据库
	db, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := respMessage.CommonResponse{
			Code:    http.StatusInternalServerError,
			Message: "数据库连接失败: " + err.Error(),
			Data:    nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer db.Close()

	// 检查用户是否存在旧的头像记录，如果有则删除旧头像文件
	var oldAvatarPath string
	query := `SELECT avatar_path FROM userMeta_data WHERE username = ?`
	err = db.QueryRow(query, username).Scan(&oldAvatarPath)
	// 只处理查询成功且存在旧头像路径的情况，忽略其他错误
	if err == nil && oldAvatarPath != "" {
		// 检查文件是否存在
		if _, err := os.Stat(oldAvatarPath); err == nil {
			// 文件存在，尝试删除
			if removeErr := os.Remove(oldAvatarPath); removeErr != nil {
				log.Printf("删除旧头像文件失败: %v", removeErr)
				// 删除失败不阻止更新过程，继续执行
			}
		}
	}

	// 检查userMeta_data表是否存在avatar_path和avatar_size字段
	var hasAvatarPathColumn, hasAvatarSizeColumn bool
	query = `SELECT COUNT(*) FROM pragma_table_info('userMeta_data') WHERE name = ?`
	err = db.QueryRow(query, "avatar_path").Scan(&hasAvatarPathColumn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := respMessage.CommonResponse{
			Code:    http.StatusInternalServerError,
			Message: "检查头像路径字段失败: " + err.Error(),
			Data:    nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	err = db.QueryRow(query, "avatar_size").Scan(&hasAvatarSizeColumn)
	if err != nil {
		writeErrorResponse(w, "检查头像大小字段失败", err)
		return
	}

	// 如果字段不存在，添加字段
	if !hasAvatarPathColumn {
		_, err = db.Exec(`ALTER TABLE userMeta_data ADD COLUMN avatar_path TEXT`)
		if err != nil {
			writeErrorResponse(w, "添加头像路径字段失败", err)
			return
		}
	}

	if !hasAvatarSizeColumn {
		_, err = db.Exec(`ALTER TABLE userMeta_data ADD COLUMN avatar_size INTEGER`)
		if err != nil {
			writeErrorResponse(w, "添加头像大小字段失败", err)
			return
		}
	}

	// 检查用户是否存在
	var exists bool
	query = `SELECT EXISTS(SELECT 1 FROM userMeta_data WHERE username = ?)`
	err = db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		writeErrorResponse(w, "查询用户失败", err)
		return
	}

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		response := respMessage.CommonResponse{
			Code:    http.StatusNotFound,
			Message: "用户不存在",
			Data:    nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 更新用户头像信息
	query = `UPDATE userMeta_data SET avatar_path = ?, avatar_size = ? WHERE username = ?`
	_, err = db.Exec(query, filePath, requestData.Size, username)
	if err != nil {
		writeErrorResponse(w, "更新头像信息失败", err)
		return
	}

	// 返回成功响应
	respMessage.SendSuccessResponse(w, "头像更新成功", nil)
}

// setTotp 设置双因素认证TOTP
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func setTotp(w http.ResponseWriter, r *http.Request) {
	// 定义需使用变量
	var (
		body          []byte
		requestData   TotpRequestData                // 解析后的TOTP请求数据
		qrCodeBase64  string                         // 二维码数据base64编码
		qrCodeData    []byte                         // 二维码数据
		respMsg       string                         // 响应消息中的message
		uri           string                         // TOTP URI
		respData      = make(map[string]interface{}) // 响应数据
		err           error
		sqliteDb      *sql.DB
		response      respMessage.CommonResponse
		decryptedData []byte
	)
	// 设置响应类型为JSON
	w.Header().Set("Content-Type", "application/json")

	// 从请求上下文中获取用户名
	currentUsername := r.Context().Value("username").(string)
	defer r.Body.Close()
	// 读取请求体数据
	body, err = io.ReadAll(r.Body)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "读取请求体失败")
		return
	}

	// 将 JSON 数据解析到EncryptedRequestData结构体
	var encryptedRequestData dataStructs.EncryptedRequestData
	if err = json.Unmarshal(body, &encryptedRequestData); err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "解析加密数据失败: "+err.Error())
		return
	}

	// 解密数据
	decryptedData, err = encryption.AesDecryptData(encryptedRequestData.IV, encryptedRequestData.EncryptedData, tmpSessionAESKey)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "解密数据失败: "+err.Error())
		return
	}

	// 解析解密后的数据到TotpRequestData结构体
	if err = json.Unmarshal(decryptedData, &requestData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response = respMessage.CommonResponse{
			Code:    http.StatusBadRequest,
			Message: "解析请求数据失败: " + err.Error(),
			Data:    nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 连接数据库
	sqliteDb, err = sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库连接失败: "+err.Error())
		return
	}
	defer sqliteDb.Close()

	// var totpColumns = map[string]string{
	// 	"totp_enabled":        "INTEGER DEFAULT 0",
	// 	"totp_secret":         "TEXT",
	// 	"totp_refresh_period": "INTEGER DEFAULT 30",
	// 	"totp_digits":         "INTEGER DEFAULT 6",
	// 	"totp_app_name":       "TEXT DEFAULT 'PasswordManager'",
	// 	"totp_username":       "TEXT",
	// 	"totp_setup_status":   "INTEGER DEFAULT 0",
	// }
	var totpColumns = []string{
		"totp_enabled INTEGER DEFAULT 0",
		"totp_secret TEXT",
		"totp_refresh_period INTEGER DEFAULT 30",
		"totp_digits INTEGER DEFAULT 6",
		"totp_app_name TEXT DEFAULT 'PasswordManager'",
		"totp_username TEXT",
		"totp_setup_status INTEGER DEFAULT 0",
	}
	// 检查userMeta_data表中是否存在TOTP相关字段，如果不存在则添加
	err = dbpkg.CheckAndCreateColumns(sqliteDb, "userMeta_data", totpColumns)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库异常: "+err.Error())
		return
	}

	// 关闭Totp
	if !requestData.Enabled {
		// 更新数据库中的TOTP设置
		_, err = sqliteDb.Exec(`
			UPDATE userMeta_data
			SET totp_enabled = 0 WHERE username = ?
		`, currentUsername)
		if err != nil {
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "切换Totp状态失败: "+err.Error())
			return
		}
		respData["enabled"] = false
		respData["qrCode"] = ""
		respData["secretKey"] = ""
		respData["setupStatus"] = false
		respData["refreshPeriod"] = ""
		respData["digits"] = ""
		respData["appName"] = ""
		respData["username"] = ""
		respMsg = "已禁用TOTP功能"
		goto ret
	}

	if requestData.SecretKey == "" || requestData.AppName == "" || requestData.Username == "" || requestData.Digits == 0 || requestData.RefreshPeriod == 0 {
		respData["setupStatus"] = false
	} else {
		respData["setupStatus"] = true
	}

	// 更新数据库中的TOTP设置
	_, err = sqliteDb.Exec(`
		UPDATE userMeta_data
		SET totp_enabled = 1,
		    totp_secret = ?,
		    totp_refresh_period = ?,
		    totp_digits = ?,
		    totp_app_name = ?,
		    totp_setup_status = ?,
		    totp_username = ?
		WHERE username = ?
	`, requestData.SecretKey, requestData.RefreshPeriod, requestData.Digits, requestData.AppName, respData["setupStatus"], requestData.Username, currentUsername)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "更新TOTP设置失败: "+err.Error())
		return
	}

	if !respData["setupStatus"].(bool) {
		respMsg = "TOTP参数缺失！"
		goto ret
	}

	// 生成TOTP URI
	uri = otherFunc.GenerateTotpURI(
		requestData.SecretKey,
		requestData.AppName,
		requestData.Username,
		requestData.Digits,
		requestData.RefreshPeriod,
	)

	// 生成二维码图像
	qrCodeData, err = otherFunc.GenerateQRCode(uri)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "生成二维码失败: "+err.Error())
		return
	}

	// 将二维码数据进行Base64编码
	qrCodeBase64 = base64.StdEncoding.EncodeToString(qrCodeData)
	// 更新数据库设置状态
	_, err = sqliteDb.Exec(`
		UPDATE userMeta_data
		SET totp_setup_status = 1
		WHERE username = ?
	`, currentUsername)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "更新TOTP设置失败: "+err.Error())
		return
	}
	respData["qrCode"] = qrCodeBase64
	respData["secretKey"] = requestData.SecretKey
	respData["enabled"] = true
	respData["setupStatus"] = true
	respData["refreshPeriod"] = requestData.RefreshPeriod
	respData["digits"] = requestData.Digits
	respData["appName"] = requestData.AppName
	respData["username"] = requestData.Username
	respMsg = "TOTP设置成功"

ret:
	// 返回响应
	respMessage.SendSuccessResponse(w, respMsg, map[string]interface{}{
		"qrCode":        respData["qrCode"], // Base64编码后的二维码数据
		"secretKey":     respData["secretKey"],
		"enabled":       respData["enabled"],
		"setupStatus":   respData["setupStatus"],
		"refreshPeriod": respData["refreshPeriod"],
		"digits":        respData["digits"],
		"appName":       respData["appName"],
		"username":      respData["username"],
	})
}

// getSecuritySettings 获取安全设置信息
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func getSecuritySettings(w http.ResponseWriter, r *http.Request) {
	// 设置响应类型为JSON
	w.Header().Set("Content-Type", "application/json")

	// 从请求中获取用户名
	username := r.Context().Value("username").(string)

	// 从请求URL中获取mod参数
	mod := r.URL.Query().Get("mod")

	// 连接数据库
	db, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库连接失败")
		return
	}
	defer db.Close()

	// 根据mod参数获取配置
	switch mod {
	case "totp":
		// 从数据库中查询用户的TOTP设置
		var totpEnabled, totpSetupStatus sql.NullInt64
		var totpSecret, totpAppName, totpUsername sql.NullString
		var totpRefreshPeriod, totpDigits sql.NullInt64

		err = db.QueryRow(`
			SELECT totp_enabled, totp_secret, totp_refresh_period, totp_digits, totp_app_name, totp_username, totp_setup_status
			FROM userMeta_data
			WHERE username = ?
		`, username).Scan(
			&totpEnabled, &totpSecret, &totpRefreshPeriod, &totpDigits, &totpAppName, &totpUsername, &totpSetupStatus,
		)

		// 处理查询结果
		if err != nil {
			// 如果记录不存在，返回默认的未启用状态
			if err == sql.ErrNoRows {
				respMessage.SendSuccessResponse(w, "获取TOTP设置成功", map[string]interface{}{
					"totpset": map[string]interface{}{
						"enabled":       false,
						"setupStatus":   false,
						"secretKey":     "",
						"refreshPeriod": 30,
						"digits":        6,
						"appName":       "PasswordManager",
						"username":      username,
					},
				})
				return
			}

			// 其他错误
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "查询TOTP设置失败: "+err.Error())
			return
		}

		// 处理可能为NULL的值，设置默认值
		secretKey := ""
		if totpSecret.Valid {
			secretKey = totpSecret.String
		}

		appName := "PasswordManager"
		if totpAppName.Valid && totpAppName.String != "" {
			appName = totpAppName.String
		}

		userNameForTotp := username
		if totpUsername.Valid && totpUsername.String != "" {
			userNameForTotp = totpUsername.String
		}

		refreshPeriod := int64(30)
		if totpRefreshPeriod.Valid && totpRefreshPeriod.Int64 > 0 {
			refreshPeriod = totpRefreshPeriod.Int64
		}

		digits := int64(6)
		if totpDigits.Valid && totpDigits.Int64 > 0 {
			digits = totpDigits.Int64
		}

		enabled := false
		if totpEnabled.Valid && totpEnabled.Int64 == 1 {
			enabled = true
		}

		setupStatus := false
		if totpSetupStatus.Valid && totpSetupStatus.Int64 == 1 {
			setupStatus = true
		}

		// 构建成功响应
		totpData := map[string]interface{}{
			"enabled":       enabled,
			"setupStatus":   setupStatus,
			"secretKey":     secretKey,
			"refreshPeriod": refreshPeriod,
			"digits":        digits,
			"appName":       appName,
			"username":      userNameForTotp,
		}

		respMessage.SendSuccessResponse(w, "Success", map[string]interface{}{
			"totpset": totpData,
		})

	case "mailAlert":
		var mailAlertStorage AlertMailStorage
		var mailConfig mail.GlobalConfig
		var decryptedJSON []byte
		// 从数据库中查询用户的邮件提醒设置
		var alertMailSetBytes []byte
		err = db.QueryRow(`
			SELECT alertMailSets
			FROM userMeta_data
			WHERE username = ?
		`, username).Scan(&alertMailSetBytes)
		if err != nil {
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "邮件告警配置查询失败: "+err.Error())
			return
		}
		// 获取原数据
		if err := json.Unmarshal(alertMailSetBytes, &mailAlertStorage); err != nil {
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "邮件告警配置解析失败: "+err.Error())
			return
		}
		if mailAlertStorage.EncryptSecret {
			// 解base64再转byte
			base64Str := string(mailAlertStorage.Data)
			if len(base64Str) >= 2 && base64Str[0] == '"' && base64Str[len(base64Str)-1] == '"' {
				base64Str = base64Str[1 : len(base64Str)-1]
			}

			encryptedBytes, err := base64.StdEncoding.DecodeString(base64Str)
			if err != nil {
				respMessage.SendErrorResponse(w, http.StatusInternalServerError, "base64 解码失败: "+err.Error())
				return
			}

			decryptedJSON, err = encryption.DpapiDecrypt(encryptedBytes, encryption.DpapiUserScope)
			if err != nil {
				respMessage.SendErrorResponse(w, http.StatusInternalServerError, "DPAPI 解密失败: "+err.Error())
				return
			}

		} else {
			var tmpResult map[string]json.RawMessage
			err := json.Unmarshal(alertMailSetBytes, &tmpResult)
			if err != nil {
				respMessage.SendErrorResponse(w, http.StatusInternalServerError, "JSON 解析失败: "+err.Error())
				return
			}
			// 提取嵌套的 "data" 部分
			dataJson, ok := tmpResult["data"]
			if !ok {
				respMessage.SendErrorResponse(w, http.StatusInternalServerError, "JSON 解析失败: 缺少 'data' 键")
				return
			}
			decryptedJSON = dataJson
		}
		// 解析获取mail配置
		if err = json.Unmarshal(decryptedJSON, &mailConfig); err != nil {
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "邮件告警配置解析失败: "+err.Error())
			return
		}
		//拼接告警配置
		mailAlertData := mail.MailAlertConfig{
			GlobalConfig:  mailConfig,
			EncryptSecret: mailAlertStorage.EncryptSecret,
			Enabled:       mailAlertStorage.Enabled,
		}

		respMessage.SendSuccessResponse(w, "Success", mailAlertData)

	case "ipWhitelist":
		var ipListStr string
		var whiteListCfg WhitelistConfig
		// 从数据库中查询用户的白名单设置
		err = db.QueryRow(`
			SELECT whitelist_enabled, whitelist_ips, whitelist_action
			FROM globalWhiteList
			WHERE username = ?
		`, username).Scan(&whiteListCfg.Enabled, &ipListStr, &whiteListCfg.ActionOutsideWhitelist)
		if err != nil {
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "白名单配置查询失败: "+err.Error())
			return
		}
		whiteListCfg.Whitelist = strings.Split(ipListStr, ",")
		// 构建成功响应
		respMessage.SendSuccessResponse(w, "Success", whiteListCfg)

	default:
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "不支持的模块: "+mod)
	}
}

// 登录发送邮件告警
// 功能：当有用户登录时，发送邮件告警通知
// 参数：
// loginAlert 发送登录告警邮件
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func loginAlert(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	// 返回成功响应
	respMessage.SendSuccessResponse(w, "登录告警邮件发送成功", nil)
}

// setLoginAlertSettings 设置登录告警配置
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func setLoginAlertSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		respMessage.SendErrorResponse(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	username := r.Context().Value("username").(string)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "无法读取请求体")
		return
	}
	defer r.Body.Close()

	var encReq dataStructs.EncryptedRequestData
	if err := json.Unmarshal(body, &encReq); err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "请求体格式错误")
		return
	}

	// 解密数据
	decryptedData, err := encryption.AesDecryptData(encReq.IV, encReq.EncryptedData, tmpSessionAESKey)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据解密失败")
		return
	}

	// 解析解密后的数据
	var req mail.MailAlertConfig
	if err := json.Unmarshal(decryptedData, &req); err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "解析数据失败")
		return
	}

	// 提取mail配置数据
	mailSettingsJSON, _ := json.Marshal(req.GlobalConfig)
	storage := AlertMailStorage{
		EncryptSecret: req.EncryptSecret,
		Enabled:       req.Enabled,
	}

	// 是否加密
	if req.EncryptSecret {
		encrypted, err := encryption.DpapiEncrypt(mailSettingsJSON, encryption.DpapiUserScope)
		if err != nil {
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "DPAPI加密失败")
			return
		}
		storage.Data = json.RawMessage(`"` + base64.StdEncoding.EncodeToString(encrypted) + `"`)
	} else {
		storage.Data = mailSettingsJSON
	}

	storageJSON, err := json.Marshal(storage)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据序列化失败")
		return
	}

	db, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库连接失败")
		return
	}
	defer db.Close()
	noExist, err := dbpkg.ColumnsExist(db, "userMeta_data", []string{"alertMailSets BLOB"})
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	if len(noExist) > 0 {
		_ = dbpkg.AddColumns(db, "userMeta_data", noExist)
	}

	if _, err := db.Exec("UPDATE userMeta_data SET alertMailSets=? WHERE username=?", storageJSON, username); err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}

	globalMail.SetGlobalConfig(mail.GlobalConfig(req.GlobalConfig))

	respMessage.SendSuccessResponse(w, "Success", req.Enabled)
}

// 验证TOTP
// verifyTotp 验证TOTP验证码
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func verifyTotp(w http.ResponseWriter, r *http.Request) {
	// 获取请求体并解密
	var encryptedRequestData dataStructs.EncryptedRequestData
	err := json.NewDecoder(r.Body).Decode(&encryptedRequestData)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusOK, "请求体格式错误: "+err.Error())
		return
	}

	// 使用会话密钥解密数据
	decryptedData, err := encryption.AesDecryptData(encryptedRequestData.IV, encryptedRequestData.EncryptedData, tmpSessionAESKey)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusOK, "解密数据失败: "+err.Error())
		return
	}
	var requestData struct {
		Username string `json:"username"`
		Code     string `json:"code"`
	}
	err = json.Unmarshal(decryptedData, &requestData)
	if err != nil {

		respMessage.SendErrorResponse(w, http.StatusOK, "请求体格式错误: "+err.Error())
		return
	}
	mu.RLock()
	status, ok := loginStatusList[requestData.Username]
	mu.RUnlock()
	// 未登录或未认证密码
	if !ok || !status.isPwdAuthed {
		// 清除cookie
		cookie := http.Cookie{
			Name:     "sessionId",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Unix(0, 0),
		}
		http.SetCookie(w, &cookie)
		respMessage.SendErrorResponse(w, http.StatusOK, "请先密码认证")
		return
	}

	// 验证totp
	valid := otherFunc.VerifyTotp(requestData.Code, requestData.Username, cfg.Database.SQLite.SqliteDbPath)
	if !valid {
		respMessage.SendErrorResponse(w, http.StatusOK, "TOTP验证失败")
		return
	}
	// 发送信号停止清除
	close(status.StopClearCh)
	status.StopClearCh = nil

	// 验证session是否有效
	valid, checkErr := sessionStore.CheckSessionValid(r, w)
	// 无session
	if checkErr != nil {
		// 新建session
		goto newSession
	}
	if valid {
		// session失效，更新session
		userSession, sessionErr := sessionStore.Get(r, "sessionId")
		if sessionErr != nil || userSession == nil {
			// 更新session失败，新建session
			goto newSession
		}
		sessionStore.Save(r, w, userSession)
		respMessage.SendSuccessResponse(w, "ok", nil)
		return
	}

newSession:
	// 创建会话
	usrSession, err := sessionStore.New(r, "sessionId")
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusOK, "创建会话失败")
		return
	}

	// 将用户名和登录时间添加到会话数据中
	usrSession.Values["username"] = requestData.Username
	usrSession.Values["loginTime"] = time.Now()
	//fmt.Print(usrSession.Values)

	// 保存会话
	sessionSaveErr := sessionStore.Save(r, w, usrSession)
	if sessionSaveErr != nil {
		respMessage.SendErrorResponse(w, http.StatusOK, "创建会话失败")
		return
	}
	respMessage.SendSuccessResponse(w, "Success", nil)
	return
}

func enableTotp(w http.ResponseWriter, r *http.Request) {
	// 设置响应类型为JSON
	w.Header().Set("Content-Type", "application/json")

	// 获取会话信息
	sessionInfo, err := sessionStore.GetSessionInfo(r)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "会话异常！")
		return
	}

	username, ok := sessionInfo["username"].(string)
	if !ok {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "会话异常！")
		return
	}
	// 连接数据库
	db, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库连接失败！")
		return
	}
	defer db.Close()

	// 检查请求方法，如果是POST，处理前端传递的数据
	if r.Method == "POST" {
		// 读取请求体数据
		body, err := io.ReadAll(r.Body)
		if err != nil {
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据异常！")
			return
		}
		defer r.Body.Close()

		// 将JSON数据解析到结构体
		var requestData struct {
			Enabled bool `json:"enabled"`
		}
		if err = json.Unmarshal(body, &requestData); err != nil {
			// 解析失败可能是因为请求体为空，继续查询操作
			log.Printf("解析请求数据失败，继续查询操作: %v", err)
		} else {
			// 如果解析成功且包含enabled字段，更新数据库
			var enabledValue int
			if requestData.Enabled {
				enabledValue = 1
			} else {
				enabledValue = 0
			}

			_, err = db.Exec(`
				UPDATE userMeta_data
				SET totp_enabled = ?
				WHERE username = ?
			`, enabledValue, username)
			if err != nil {
				respMessage.SendErrorResponse(w, http.StatusInternalServerError, "更新TOTP启用状态失败: "+err.Error())
				return
			}
		}
	}

	// 查询数据库获取最新状态
	var result TotpAllSet
	err = db.QueryRow("SELECT totp_secret, totp_refresh_period, totp_app_name, totp_username, totp_digits, totp_enabled ,totp_setup_status FROM userMeta_data WHERE username = ?", username).Scan(&result.SecretKey, &result.RefreshPeriod, &result.AppName, &result.Username, &result.Digits, &result.Enabled, &result.SetupStatus)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "查询数据库失败！")
		return
	}
	respMessage.SendSuccessResponse(w, "Success", result)
	return
}

func setWhitelistSettings(w http.ResponseWriter, r *http.Request) {
	// 获取用户名
	sessionInfo, err := sessionStore.GetSessionInfo(r)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "会话异常！")
		return
	}
	username, ok := sessionInfo["username"].(string)
	if !ok {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "会话异常！")
		return
	}
	var encryptedRequestData dataStructs.EncryptedRequestData
	err = json.NewDecoder(r.Body).Decode(&encryptedRequestData)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "请求体格式错误: "+err.Error())
		return
	}
	// 解密
	decryptedData, err := encryption.AesDecryptData(encryptedRequestData.IV, encryptedRequestData.EncryptedData, tmpSessionAESKey)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "解密数据失败: "+err.Error())
		return
	}

	// 解析解密后的JSON数据
	var whitelistConfig WhitelistConfig
	err = json.Unmarshal(decryptedData, &whitelistConfig)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "解析解密数据失败: "+err.Error())
		return
	}
	// 检查IP格式异常返回
	if !tools.IpFormatCheck(&whitelistConfig.Whitelist) {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "IP格式错误")
		return
	}
	// 检查ActionOutsideWhitelist是否为"alert"或"block"
	if whitelistConfig.ActionOutsideWhitelist != "alert" && whitelistConfig.ActionOutsideWhitelist != "block" {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "非法数据！")
		return
	}
	// 写入数据库
	sqliteDB, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据库连接失败")
		return
	}
	defer sqliteDB.Close()

	tx, err := sqliteDB.Begin()
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "开始事务失败: "+err.Error())
		return
	}

	// 检查表列是否存在
	err = dbpkg.CheckAndCreateColumns(sqliteDB, "globalWhiteList", []string{"username TEXT", "whitelist_enabled INTEGER", "whitelist_ips TEXT", "whitelist_action TEXT"})
	if err != nil {
		tx.Rollback()
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "检查或创建数据库列失败: "+err.Error())
		return
	}

	// 检查username对应的记录是否存在
	var exists bool
	err = sqliteDB.QueryRow("SELECT EXISTS(SELECT 1 FROM globalWhiteList WHERE username = ?)", username).Scan(&exists)
	if err != nil {
		tx.Rollback()
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "检查用户名是否存在失败: "+err.Error())
		return
	}
	if !exists {
		// 如果不存在，插入新记录
		insertRes, insertErr := sqliteDB.Exec(`
			INSERT INTO globalWhiteList (username, whitelist_enabled, whitelist_ips, whitelist_action)
			VALUES (?, ?, ?, ?)
		`, username, whitelistConfig.Enabled, strings.Join(whitelistConfig.Whitelist, ","), whitelistConfig.ActionOutsideWhitelist)
		if insertErr != nil {
			tx.Rollback()
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "插入白名单配置失败: "+insertErr.Error())
			return
		}
		rowsAffected, err := insertRes.RowsAffected()
		if err != nil {
			tx.Rollback()
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据保存失败: "+err.Error())
			return
		}
		if rowsAffected == 0 {
			tx.Rollback()
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "未更新任何行，可能用户名不存在")
			return
		}
		commitErr := tx.Commit()
		if commitErr != nil {
			tx.Rollback()
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "提交事务失败: "+commitErr.Error())
			return
		}
		respMessage.SendSuccessResponse(w, "Success", nil)
		return
	} else {
		// 如果存在，更新记录
		saveRes, saveErr := sqliteDB.Exec(`
			UPDATE globalWhiteList
			SET whitelist_enabled = ?, whitelist_ips = ?, whitelist_action = ?
			WHERE username = ?
		`, whitelistConfig.Enabled, strings.Join(whitelistConfig.Whitelist, ","), whitelistConfig.ActionOutsideWhitelist, username)
		if saveErr != nil {
			tx.Rollback()
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "更新白名单配置失败: "+saveErr.Error())
			return
		}
		// 检查是否有影响的行
		rowsAffected, err := saveRes.RowsAffected()
		if err != nil {
			tx.Rollback()
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "数据保存失败: "+err.Error())
			return
		}
		if rowsAffected == 0 {
			tx.Rollback()
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "未更新任何行，可能用户名不存在")
			return
		}
		// 提交事务
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "提交事务失败: "+err.Error())
			return
		}
		if rowsAffected == 0 {
			tx.Rollback()
			respMessage.SendErrorResponse(w, http.StatusInternalServerError, "未更新任何行，可能用户名不存在")
			return
		}
		respMessage.SendSuccessResponse(w, "Success", nil)
		return
	}
}

// 限流状态数据查询
func stats(w http.ResponseWriter, r *http.Request) {
	stats := rateLimiter.Stats()
	respMessage.SendSuccessResponse(w, "Success", stats)
}

func importByFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respMessage.SendErrorResponse(w, http.StatusMethodNotAllowed, "仅支持POST方法")
		return
	}
	var encryptedRequestData dataStructs.EncryptedRequestData
	err := json.NewDecoder(r.Body).Decode(&encryptedRequestData)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "请求体格式错误: "+err.Error())
		return
	}
	// 解密
	decryptedData, err := encryption.AesDecryptData(encryptedRequestData.IV, encryptedRequestData.EncryptedData, tmpSessionAESKey)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "解密数据失败: "+err.Error())
		return
	}
	// 解析 JSON 数据
	var jsonDatas []dataHandler.InputData
	err = json.Unmarshal(decryptedData, &jsonDatas)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "解析 JSON 数据失败: "+err.Error())
		return
	}
	//decryptedDataStr := string(decryptedData)
	//fmt.Println(jsonDatas)
	// 数据批量保存
	if err = dbpkg.SaveToDatabaseBatch(jsonDatas, cfg.Database.SQLite.SqliteDbPath); err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, "保存数据出错: "+err.Error())
		return
	}
	// 保存成功后返回响应
	respMessage.SendSuccessResponse(w, "Success", nil)
}

func exportAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respMessage.SendErrorResponse(w, http.StatusMethodNotAllowed, "仅支持GET方法")
		return
	}

	// 检查key是否为空
	if userAesKeyHex == "" {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "用户AES密钥不能为空")
		return
	}
	// 密钥转byte
	userAesKey, err := hex.DecodeString(userAesKeyHex)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusBadRequest, "用户AES密钥格式错误: "+err.Error())
		return
	}
	dbpkg.ExportAllHandler(w,r,cfg.Database.SQLite.SqliteDbPath,userAesKey)
}
