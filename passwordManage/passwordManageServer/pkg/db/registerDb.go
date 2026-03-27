/**
 * @Author: 小鱼
 * @Date: 2024-04-18 13:22:35
 * @LastEditors: 小鱼
 * @LastEditTime: 2024-04-18 13:22:35
 * @FilePath: \passwordManageServer\pkg\db\registerDb.go
 * @Description: 用户注册相关数据库操作，包含保存用户加密数据等功能
 *
 * Copyright (c) 2024 by 小鱼, All Rights Reserved.
 */
package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// SaveEncryptedUserDataToDB 存储用户加密数据到数据库
// - 功能说明：创建用户元数据表（如果不存在），检查用户名是否已存在，并保存用户加密数据
// - 参数：
//   - db: 数据库连接对象
//   - username: 用户名
//   - passwordHash: 密码哈希值
//   - publicKey: 公钥字符串
//   - nonce: 加密使用的随机数
//   - ciphertext: 加密后的密文
// - 返回值：
//   - 错误信息（如果保存失败或用户名已存在）
func SaveEncryptedUserDataToDB(db *sql.DB, username string, passwordHash string, publicKey string, nonce, ciphertext []byte) error {

	var err error
	// 创建表格（如果表格不存在）
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS userMeta_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE,
			password_hash TEXT,
			public_key TEXT,
			nonce BLOB,
			ciphertext BLOB	
		)
	`)
	if err != nil {
		return fmt.Errorf("创建表失败: %v", err)
	}

	// 查询数据库以检查用户名是否已存在
	var exists bool
	err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM userMeta_data WHERE username = ?)`, username).Scan(&exists)
	if err != nil {
		return fmt.Errorf("查询用户名失败: %v", err)
	}

	// 如果用户名已存在，则返回错误
	if exists {
		return fmt.Errorf("用户名已存在")
	}

	// 将密文、公钥、用户名和密码哈希插入到数据库
	_, err = db.Exec(`
		INSERT INTO userMeta_data (username, password_hash, public_key, nonce, ciphertext)
		VALUES (?, ?, ?, ?, ?)
	`, username, passwordHash, publicKey, nonce, ciphertext)
	if err != nil {
		return fmt.Errorf("插入数据失败: %v", err)
	}

	return nil
}

// UpdateRegisterInfo 更新用户注册信息
// - 功能说明：检查用户名是否存在，并更新用户的密码哈希、公钥和加密数据
// - 参数：
//   - db: 数据库连接对象
//   - username: 用户名
//   - passwordHash: 密码哈希值
//   - publicKey: 公钥字符串
//   - nonce: 加密使用的随机数
//   - ciphertext: 加密后的密文
// - 返回值：
//   - 错误信息（如果更新失败或用户名不存在）
func UpdateRegisterInfo(db *sql.DB, username string, passwordHash string, publicKey string, nonce, ciphertext []byte) error {
	// 查询数据库以检查用户名是否存在
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM userMeta_data WHERE username = ?)`, username).Scan(&exists)
	if err != nil {
		return fmt.Errorf("数据库错误: %v", err)
	}

	// 如果用户名不存在，则返回错误
	if !exists {
		return fmt.Errorf("用户名不存在")
	}

	// 更新数据
	_, err = db.Exec(`
		UPDATE userMeta_data
		SET password_hash = ?, public_key = ?, nonce = ?, ciphertext = ?
		WHERE username = ?
	`, passwordHash, publicKey, nonce, ciphertext, username)
	if err != nil {
		return fmt.Errorf("修改失败: %v", err)
	}

	return nil
}
