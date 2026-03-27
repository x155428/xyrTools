/*
 * @Author: 小鱼
 * @Date: 2024-10-27 16:14:10
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-09-29 14:12:48
 * @FilePath: \passwordManageServer\pkg\tools\passwordStrength.go
 * @Description: 密码强度检测工具，提供密码强度评估和分析功能
 *
 * Copyright (c) 2024 by 小鱼, All Rights Reserved.
 */
package tools

import (
	"github.com/nbutton23/zxcvbn-go"
)

// PasswordStrengthResult 封装密码强度评分和描述
type PasswordStrengthResult struct {
	Score           int    // 密码强度评分（0到4）
	StrengthMessage string // 密码强度描述映射
}

/**
 * @description: 接收密码，返回评分和等级
 * @param {string} password 密码
 * @return {*}
 */
func EvaluateStrength(password string) PasswordStrengthResult {
	result := zxcvbn.PasswordStrength(password, []string{})
	score := int(result.Score)

	// 根据评分映射强度描述
	var strengthMessage string
	switch score {
	case 0:
		strengthMessage = "非常弱"
	case 1:
		strengthMessage = "弱"
	case 2:
		strengthMessage = "一般"
	case 3:
		strengthMessage = "强"
	case 4:
		strengthMessage = "非常强"
	default:
		strengthMessage = "未知"
	}

	return PasswordStrengthResult{
		Score:           score,
		StrengthMessage: strengthMessage,
	}
}
