/**
 * @Author: 小鱼
 * @Date: 2024-04-18 13:22:35
 * @LastEditors: 小鱼
 * @LastEditTime: 2024-04-18 13:22:35
 * @FilePath: \passwordManageServer\pkg\db\db.go
 * @Description: 数据模型定义，包含数据库相关的结构体和常量
 *
 * Copyright (c) 2024 by 小鱼, All Rights Reserved.
 */
package db

// PasswordType 表示密码的类型
const (
	CharPasswordType    = iota // 0: 字符密码类型
	KeyFilePasswordType        // 1: 密钥文件类型
)

type EncryptedData struct {
	IV   string `json:"iv"`   // 加密使用的初始向量
	Data string `json:"data"` // 加密后的数据
}

// KeyFileInfo 密钥文件信息结构体
type KeyFile struct {
	FileName    string        `json:"name"`    // 文件名
	FileSize    int           `json:"size"`    // 文件大小
	FileType    string        `json:"type"`    // 文件类型
	FileContent EncryptedData `json:"content"` // 文件内容加密数据
}

// PasswordEntry 密码条目数据模型
type PasswordEntry struct {
	AppName             EncryptedData `json:"appName"`             // 应用名加密数据
	IsAppNameEncrypted  bool          `json:"isAppNameEncrypted"`  // 应用名是否加密
	Username            string        `json:"username"`            // 用户名
	IsUsernameEncrypted bool          `json:"isUsernameEncrypted"` // 用户名是否加密
	InputType           string        `json:"inputType"`           // 输入类型
	Password            *string       `json:"password"`            // 密码（可能为 null）
	KeyFile             KeyFile       `json:"keyFile"`             // 密钥文件信息
	FileHash            string        `json:"fileHash"`            // 文件哈希值
	URL                 string        `json:"url"`                 // URL
	IsUrlEncrypted      bool          `json:"isUrlEncrypted"`      // URL 是否加密
	Notes               string        `json:"notes"`               // 备注
	IsNotesEncrypted    bool          `json:"isNotesEncrypted"`    // 备注是否加密
	Tags                []string      `json:"tags"`                // 标签
	IsTagsEncrypted     bool          `json:"isTagsEncrypted"`     // 标签是否加密
}

// UserMeta 用户元数据结构体
type UserMeta struct {
	Username     string
	PasswordHash string
	PublicKey    string
	Nonce        []byte
	Ciphertext   []byte
}
