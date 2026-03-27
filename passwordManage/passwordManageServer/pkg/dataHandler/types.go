/*
 * @Author: 小鱼
 * @Date: 2025-09-29 14:00:00
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-09-29 14:00:00
 * @FilePath: \passwordManageServer\pkg\dataHandler\types.go
 * @Description: 数据处理相关的类型定义，包含加密字符串和密钥文件等数据结构
 *
 * Copyright (c) 2025 by 小鱼, All Rights Reserved.
 */
package dataHandler

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// EncryptedString 定义加密字符串结构，同时支持普通字符串
type EncryptedString struct {
	IV   string         `json:"iv,omitempty"`
	Data string         `json:"data,omitempty"`
	Raw  sql.NullString `json:"-"` // 普通字符串
}

// KeyFile 定义 keyFile 的结构
type KeyFile struct {
	Name     string         `json:"name"`
	Size     int            `json:"size"`
	Type     string         `json:"type"`
	Content  KeyFileContent `json:"content"`
	FileHash string         `json:"fileHash"`
}

// InputData 定义完整输入数据结构
type InputData struct {
	AppName             EncryptedString `json:"appName"`
	IsAppNameEncrypted  bool            `json:"isAppNameEncrypted"`
	Username            EncryptedString `json:"username"`
	IsUsernameEncrypted bool            `json:"isUsernameEncrypted"`
	InputType           string          `json:"inputType"`
	Password            EncryptedString `json:"password"`
	KeyFile             NullableKeyFile `json:"keyFile"`
	URL                 EncryptedString `json:"url"`
	IsUrlEncrypted      bool            `json:"isUrlEncrypted"`
	Notes               EncryptedString `json:"notes"`
	IsNotesEncrypted    bool            `json:"isNotesEncrypted"`
	Tags                EncryptedString `json:"tags"`
	IsTagsEncrypted     bool            `json:"isTagsEncrypted"`
	ChoseEncrypt        EncryptedString `json:"choseCrypto"`
	Key                 EncryptedString `json:"key"`
}

// QueryData 定义查询数据结构，密码管理中查询数据
type QueryData struct {
	Id                  int             `json:"id"`
	AppName             EncryptedString `json:"appName"`
	IsAppNameEncrypted  bool            `json:"isAppNameEncrypted"`
	Username            EncryptedString `json:"username"`
	IsUsernameEncrypted bool            `json:"isUsernameEncrypted"`
	InputType           string          `json:"inputType"`
	Password            EncryptedString `json:"password"`
	KeyFile             NullableKeyFile `json:"keyFile"`
	URL                 EncryptedString `json:"url"`
	IsUrlEncrypted      bool            `json:"isUrlEncrypted"`
	Notes               EncryptedString `json:"notes"`
	IsNotesEncrypted    bool            `json:"isNotesEncrypted"`
	Tags                EncryptedString `json:"tags"`
	IsTagsEncrypted     bool            `json:"isTagsEncrypted"`
	Strength            sql.NullString  `json:"strength"`
	ChoseEncrypt        EncryptedString `json:"choseCrypto"`
	Key                 EncryptedString `json:"key"`
	Count               sql.NullInt64   `json:"count"`
}

// UnmarshalJSON 自定义解析方法
func (es *EncryptedString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		// 如果不是普通字符串，尝试解析为加密对象
		var encrypted struct {
			IV   string `json:"iv"`
			Data string `json:"data"`
		}
		if err := json.Unmarshal(data, &encrypted); err != nil {
			return err
		}
		es.IV = encrypted.IV
		es.Data = encrypted.Data
		es.Raw = sql.NullString{Valid: false}
		return nil
	}
	es.Raw = sql.NullString{String: s, Valid: true}
	return nil
}

// MarshalJSON 自定义序列化方法
func (es EncryptedString) MarshalJSON() ([]byte, error) {
	if es.Raw.Valid {
		return json.Marshal(es.Raw.String)
	}
	if es.IV != "" && es.Data != "" {
		return json.Marshal(struct {
			IV   string `json:"iv"`
			Data string `json:"data"`
		}{
			IV:   es.IV,
			Data: es.Data,
		})
	}
	return json.Marshal(nil)
}

// KeyFileContent 定义 keyFile 的加密内容结构
type KeyFileContent struct {
	IV   string `json:"iv"`
	Data string `json:"data"`
}

// NullableKeyFile 支持 keyFile 为对象或 null
type NullableKeyFile struct {
	File *KeyFile
}

// UnmarshalJSON 自定义反序列化，处理 null 和对象
func (nkf *NullableKeyFile) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nkf.File = nil
		return nil
	}

	var file KeyFile
	if err := json.Unmarshal(data, &file); err != nil {
		return err
	}
	nkf.File = &file
	return nil
}

// MarshalJSON 自定义序列化
func (nkf NullableKeyFile) MarshalJSON() ([]byte, error) {
	if nkf.File == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(nkf.File)
}

// 打印数据，调试用
func AnalyzeDecryptedDataPrint(input InputData) {
	if input.IsAppNameEncrypted {
		fmt.Println("AppName (encrypted):", input.AppName.IV, input.AppName.Data)
	} else {
		fmt.Println("AppName (plain):", input.AppName.Raw)
	}

	// 处理 keyFile
	if input.KeyFile.File != nil {
		fmt.Println("KeyFile Name:", input.KeyFile.File.Name)
		fmt.Println("KeyFile Size:", input.KeyFile.File.Size)
		fmt.Println("KeyFile Content IV:", input.KeyFile.File.Content.IV)
		fmt.Println("KeyFile Hash:", input.KeyFile.File.FileHash)
	} else {
		fmt.Println("KeyFile: null")
	}
}
