/*
 * @Author: 小鱼
 * @Date: 2025-09-29 14:00:00
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-09-29 14:00:00
 * @FilePath: \passwordManageServer\pkg\otherFunc\otherFunc.go
 * @Description: 其他辅助功能函数，包含密码哈希生成和校验等功能
 *
 * Copyright (c) 2025 by 小鱼, All Rights Reserved.
 */
package otherFunc

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// 生成密码哈希值
func GeneratePasswordHash(password string) (string, error) {
	// 使用 bcrypt 生成密码哈希
	// bcrypt.DefaultCost 是推荐的默认成本，表示算法复杂度，增加复杂度可增加破解难度
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码哈希生成失败: %v", err)
	}
	return string(hash), nil
}

// 校验输入的密码是否与存储的哈希匹配
func VerifyPassword(storedHash, password string) bool {
	// 使用 bcrypt 比较存储的哈希和用户输入的密码
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		// 如果密码不匹配，返回 false
		return false
	}
	// 如果密码匹配，返回 true
	return true
}
