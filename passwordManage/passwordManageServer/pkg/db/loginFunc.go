/**
 * @Author: 小鱼
 * @Date: 2024-04-18 13:22:35
 * @LastEditors: 小鱼
 * @LastEditTime: 2024-04-18 13:22:35
 * @FilePath: \passwordManageServer\pkg\db\loginFunc.go
 * @Description: 用户登录相关数据库操作，包含验证用户密码等功能
 *
 * Copyright (c) 2024 by 小鱼, All Rights Reserved.
 */
package db

import (
	"database/sql"
	"fmt"
)

// GetPasswordHash 根据用户名查找密码哈希
// 参数：
// - db: 数据库连接对象
// - username: 用户名
// 返回值：
// - 密码哈希值
// - 错误信息
func GetPasswordHash(db *sql.DB, username string) (string, error) {
	// SQL 查询，查找指定用户名的密码哈希
	query := `SELECT password_hash FROM userMeta_data WHERE username = ?`

	var passwordHash string

	// 执行查询并扫描结果
	err := db.QueryRow(query, username).Scan(&passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果没有找到该用户名
			return "", fmt.Errorf("用户 %s 不存在", username)
		}
		return "", fmt.Errorf("查询失败: %v", err)
	}

	return passwordHash, nil
}

// GetUserMeta 获取用户的元数据
// - 功能说明：根据用户名查询用户的完整元数据信息
// - 参数：
//   - db: 数据库连接对象
//   - username: 用户名
// - 返回值：
//   - UserMeta: 包含用户名、密码哈希、公钥等信息的用户元数据结构体
//   - 错误信息（如果查询失败）
func GetUserMeta(db *sql.DB, username string) (UserMeta, error) {
	// SQL 查询，查找指定用户名的密码哈希
	query := `SELECT username, password_hash, public_key, nonce, ciphertext FROM userMeta_data WHERE username = ?`

	var userMeta UserMeta

	// 执行查询并扫描结果

	err := db.QueryRow(query, username).Scan(&userMeta.Username, &userMeta.PasswordHash, &userMeta.PublicKey, &userMeta.Nonce, &userMeta.Ciphertext)
	if err != nil {
		return userMeta, fmt.Errorf("查询失败: %v", err)
	}

	return userMeta, nil
}
