package db

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"sync"
	"time"

	"xyrTools/passwordManage/passwordManageServer/pkg/dataStructs"
	"xyrTools/passwordManage/passwordManageServer/pkg/encryption"
	respMessage "xyrTools/passwordManage/passwordManageServer/pkg/responseStd"

	_ "github.com/mattn/go-sqlite3"
	"github.com/xuri/excelize/v2"
)

// 后端字段配置
var exportFields = []struct {
	FieldName string
	ColTitle  string
}{
	{"AppName", "name"},
	{"Username", "username"},
	{"Password", "password"},
	{"URL", "url"},
	{"Notes", "notes"},
	{"Tags", "tags"},
	{"KeyFile", "key_file"},
}

// -------------------
// worker task
type workItem struct {
	ID          int64
	Data        dataStructs.OutputData
	ChoseEncrypt string
	KeyBytes    []byte
}

type outputRow struct {
	Data dataStructs.OutputData
	Err  error
}

// -------------------
// 查询 worker: 分页查询 + 提取加密信息
func queryWorker(db *sql.DB, userAesKey []byte, batchSize int, workChan chan<- workItem) {
	defer close(workChan)

	var lastID int64 = 0
	for {
		rows, err := db.Query(`SELECT 
			rowid, app_name, is_app_name_encrypted, username, is_username_encrypted,
			input_type, password, key_file, url, is_url_encrypted,
			notes, is_notes_encrypted, tags, is_tags_encrypted, chose_encrypt, key
		FROM input_data
		WHERE rowid > ? ORDER BY rowid ASC LIMIT ?`, lastID, batchSize)
		if err != nil {
			log.Println("查询数据库失败:", err)
			return
		}

		count := 0
		for rows.Next() {
			var rowid int64
			var o dataStructs.OutputData
			if err := rows.Scan(
				&rowid, &o.AppName, &o.IsAppNameEncrypted, &o.Username, &o.IsUsernameEncrypted,
				&o.InputType, &o.Password, &o.KeyFile, &o.URL, &o.IsUrlEncrypted,
				&o.Notes, &o.IsNotesEncrypted, &o.Tags, &o.IsTagsEncrypted, &o.ChoseEncrypt, &o.Key,
			); err != nil {
				log.Println("扫描数据库行失败:", err)
				continue
			}
			lastID = rowid

			// 解析密钥
			var tmpEncryptedKey dataStructs.EncryptedUserDataDB
			if err := json.Unmarshal([]byte(o.Key), &tmpEncryptedKey); err != nil {
				log.Println("解析 JSON 密钥失败:", err)
				continue
			}
			keyBytes, err := encryption.AesDecryptData(tmpEncryptedKey.IV, tmpEncryptedKey.Data, userAesKey)
			if err != nil {
				log.Println("解密密钥失败:", err)
				continue
			}
			keyBytesHex, err := hex.DecodeString(string(keyBytes))
			if err != nil {
				log.Println("解密密钥十六进制编码失败:", err)
				continue
			}

			choseEncrypt := parseChoseEncrypt(o.ChoseEncrypt, userAesKey)

			// 写入 workChan，如果缓冲满就等待
			for {
				select {
				case workChan <- workItem{ID: rowid, Data: o, ChoseEncrypt: choseEncrypt, KeyBytes: keyBytesHex}:
					break
				default:
					time.Sleep(1 * time.Millisecond) // 缓冲满时等待
					continue
				}
				break
			}

			count++
		}
		rows.Close()
		if count < batchSize {
			break
		}
	}
}

// 解析 choseEncrypt
func parseChoseEncrypt(s string, userAesKey []byte) string {
	var tmp dataStructs.EncryptedUserDataDB
	if err := json.Unmarshal([]byte(s), &tmp); err != nil {
		return ""
	}
	plain, err := encryption.AesDecryptData(tmp.IV, tmp.Data, userAesKey)
	if err != nil {
		return ""
	}
	return string(plain)
}

// -------------------
// 解密 worker
func decryptWorker(workChan <-chan workItem, resultChan chan<- outputRow, wg *sync.WaitGroup) {
	defer wg.Done()
	for w := range workChan {
		if w.ChoseEncrypt == "AES-GCM" {
			if err := decryptOutputData(&w.Data, w.KeyBytes); err != nil {
				resultChan <- outputRow{Err: fmt.Errorf("解密数据失败: %w", err)}
				continue
			}
		} else {
			w.Data.Password = fmt.Sprintf("[不支持的加密方法: %s]", w.ChoseEncrypt)
		}
		// 写入 resultChan，如果缓冲满就等待
		for {
			select {
			case resultChan <- outputRow{Data: w.Data}:
				break
			default:
				time.Sleep(5 * time.Millisecond)
				continue
			}
			break
		}
	}
}

// -------------------
// 解密 OutputData
func decryptOutputData(o *dataStructs.OutputData, mainKey []byte) error {
	tasks := []struct {
		Encrypted bool
		Value     *string
	}{
		{o.IsAppNameEncrypted, &o.AppName},
		{o.IsUsernameEncrypted, &o.Username},
		{o.IsUrlEncrypted, &o.URL},
		{o.IsNotesEncrypted, &o.Notes},
		{o.IsTagsEncrypted, &o.Tags},
		{true, &o.Password}, // password 必定加密
		{false, &o.KeyFile}, // key_file 不参与解密，输出原始数据，过大处理太复杂
	}

	for _, t := range tasks {
		if !t.Encrypted || *t.Value == "" {
			continue
		}
		var tmp dataStructs.EncryptedUserDataDB
		if err := json.Unmarshal([]byte(*t.Value), &tmp); err != nil {
			return fmt.Errorf("解析 JSON 失败: %w", err)
		}
		plain, err := encryption.AesDecryptData(tmp.IV, tmp.Data, mainKey)
		if err != nil {
			return fmt.Errorf("解密失败: %w", err)
		}
		*t.Value = string(plain)
	}
	return nil
}

// -------------------
// OutputData 转 Excel 一行
func outputDataToSlice(o dataStructs.OutputData, fields []struct{ FieldName, ColTitle string }) []interface{} {
	val := reflect.ValueOf(o)
	row := make([]interface{}, len(fields))
	for i, f := range fields {
		field := val.FieldByName(f.FieldName)
		if field.IsValid() {
			row[i] = field.Interface()
		} else {
			row[i] = ""
		}
	}
	return row
}

// -------------------
// 主导出函数
func ExportAllHandler(w http.ResponseWriter, r *http.Request, dbPath string, userAesKey []byte) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("数据库连接失败: %v", err))
		return
	}
	defer db.Close()

	f := excelize.NewFile()
	stream, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("创建 Excel 流写入器失败: %v", err))
		return
	}

	// 写表头
	header := make([]interface{}, len(exportFields))
	for i, f := range exportFields {
		header[i] = f.ColTitle
	}
	stream.SetRow("A1", header)

	workChan := make(chan workItem, 1000)
	resultChan := make(chan outputRow, 1000)

	// 启动查询 worker
	go queryWorker(db, userAesKey, 1000, workChan)

	// 启动解密 worker pool
	numCPU := runtime.NumCPU()
	var wg sync.WaitGroup
	for i := 0; i < numCPU; i++ {
		wg.Add(1)
		go decryptWorker(workChan, resultChan, &wg)
	}

	// 关闭 resultChan
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 流式写 Excel，锁保护
	var writeMutex sync.Mutex
	rowIndex := 2
	for r := range resultChan {
		if r.Err != nil {
			log.Println("导出数据解密失败:", r.Err)
			continue
		}
		writeMutex.Lock()
		stream.SetRow(fmt.Sprintf("A%d", rowIndex), outputDataToSlice(r.Data, exportFields))
		rowIndex++
		writeMutex.Unlock()
	}

	if err := stream.Flush(); err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("文件刷新失败: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="密码导出记录_`+time.Now().Format("20060102")+`.xlsx"`)

	if err := f.Write(w); err != nil {
		respMessage.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("文件写入失败: %v", err))
		return
	}

	respMessage.SendSuccessResponse(w, "导出完成", nil)
}
