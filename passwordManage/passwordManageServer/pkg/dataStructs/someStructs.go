/*
 * @Author: 小鱼
 * @Date: 2025-09-29 14:00:00
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-09-29 14:00:00
 * @FilePath: \passwordManageServer\pkg\dataStructs\someStructs.go
 * @Description: 数据结构定义模块，包含系统中使用的各种数据结构和请求/响应模型
 *
 * Copyright (c) 2025 by 小鱼, All Rights Reserved.
 */
package dataStructs

import "encoding/json"

// UserSettings 用户设置结构体
type SettingRequest struct {
	SetType string          `json:"setType"` // 配置类型
	Data    json.RawMessage `json:"data"`    // 用于接收配置的具体数据
}

// 注册返回数据结构体
type RegisterResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type EncryptedUserData struct {
	UserInfo   []string `json:"userInfo"`
	UserPasswd string   `json:"userPasswd"`
	AESKey     []byte   `json:"aesKey"`
}

// 用于接收注册/增加记录
type EncryptedRequestData struct {
	IV            []byte `json:"iv"`
	EncryptedData []byte `json:"encryptedData"`
}

// 用户接收登录请求体
type LoginPwdRequest struct {
	IV            []byte `json:"iv"`
	EncryptedData []byte `json:"encryptedData"`
}

// 数据库单项密文存储结构byte
type EncryptedUserDataDB struct {
	IV   []byte `json:"iv"`
	Data []byte `json:"data"`
}

// 数据库单项密文存储结构str
type EncryptedUserDataDBStr struct {
	IV   string `json:"iv"`
	Data string `json:"data"`
}

// 注册数据解密结构体
type DecryptedRegisterData struct {
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	CustomInfo []string `json:"customInfo"`
}

// 登录数据解密结构体
type DecryptedLoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SystemNameSettings struct {
	SysName string `json:"sysName"`
}

type SecuritySettings struct {
	Timeout int `json:"timeout"`
}

type WeatherAPISettings struct {
	APIKey string `json:"apiKey"`
}

// 导出数据结构
type OutputData struct {
	AppName             string `json:"appName"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	KeyFile             string `json:"keyFile"`
	URL                 string `json:"url"`
	Notes               string `json:"notes"`
	Tags                string `json:"tags"`
	ChoseEncrypt        string `json:"choseEncrypt"`
	Key                 string `json:"key"`
	InputType           string `json:"inputType"`
	IsAppNameEncrypted  bool   `json:"isAppNameEncrypted"`
	IsUsernameEncrypted bool   `json:"isUsernameEncrypted"`
	IsUrlEncrypted      bool   `json:"isUrlEncrypted"`
	IsNotesEncrypted    bool   `json:"isNotesEncrypted"`
	IsTagsEncrypted     bool   `json:"isTagsEncrypted"`
}
