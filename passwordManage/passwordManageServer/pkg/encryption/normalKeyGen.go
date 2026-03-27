/**
 * @Author: 小鱼
 * @Date: 2024-04-18 13:22:35
 * @LastEditors: 小鱼
 * @LastEditTime: 2024-04-18 13:22:35
 * @FilePath: \passwordManageServer\pkg\encryption\normalKeyGen.go
 * @Description: 密钥生成管理模块，用于生成各种类型的安全密钥和密码
 *
 * Copyright (c) 2024 by 小鱼, All Rights Reserved.
 */
package encryption
import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// 默认配置范围和约束
const (
	DefaultLengthMin = 12
	DefaultLengthMax = 24
	MinSpecialMin    = 2
	MinSpecialMax    = 6
	MinDigitMin      = 2
	MinDigitMax      = 6
)

// PasswordConfig 密码生成配置结构
type PasswordConfig struct {
	Length     int `json:"length,omitempty"`     // 密码总长度
	MinSpecial int `json:"minSpecial,omitempty"` // 最小特殊字符数
	MinDigit   int `json:"minDigit,omitempty"`   // 最小数字数
}

// GenerateComplexPassword 基于JSON配置生成高强度密码
func GenerateComplexPassword(configJSON string) (string, error) {
	// 解析JSON配置（如果有）
	var config PasswordConfig
	if configJSON != "" {
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			return "", fmt.Errorf("JSON解析失败: %w", err)
		}
	}

	// 生成智能随机默认值
	if config.Length == 0 {
		// 在12-24之间随机选择长度
		length, err := randInt(DefaultLengthMin, DefaultLengthMax)
		if err != nil {
			return "", err
		}
		config.Length = length
	}

	// 确保特殊字符数有效
	if config.MinSpecial == 0 {
		// 在2-6之间随机选择特殊字符数
		minSpecial, err := randInt(MinSpecialMin, MinSpecialMax)
		if err != nil {
			return "", err
		}
		config.MinSpecial = minSpecial
	}

	// 确保数字数有效
	if config.MinDigit == 0 {
		// 在2-6之间随机选择数字数
		minDigit, err := randInt(MinDigitMin, MinDigitMax)
		if err != nil {
			return "", err
		}
		config.MinDigit = minDigit
	}

	// 验证配置有效性
	if config.Length < 8 {
		return "", errors.New("密码长度必须至少8位")
	}

	if config.MinSpecial < 0 || config.MinDigit < 0 {
		return "", errors.New("特殊字符和数字数量不能为负数")
	}

	totalSpecialDigits := config.MinSpecial + config.MinDigit
	if totalSpecialDigits > config.Length {
		maxAllowed := config.Length
		return "", fmt.Errorf("特殊字符(%d)和数字(%d)要求超过可用长度(%d)",
			config.MinSpecial, config.MinDigit, maxAllowed)
	}

	// 字符集定义
	const (
		lowerChars   = "abcdefghjkmnpqrstuvwxyz"  // 排除易混淆字符(l, i, o)
		upperChars   = "ABCDEFGHJKLMNPQRSTUVWXYZ" // 排除易混淆字符(I, O)
		digits       = "23456789"                 // 排除易混淆数字(0, 1)
		specialChars = "!@#$%^&*-_=+?.,:;"
	)

	//allChars := lowerChars + upperChars
	var password strings.Builder

	// 生成特殊字符
	for i := 0; i < config.MinSpecial; i++ {
		char, err := randomChar(specialChars)
		if err != nil {
			return "", err
		}
		password.WriteByte(char)
	}

	// 生成数字
	for i := 0; i < config.MinDigit; i++ {
		char, err := randomChar(digits)
		if err != nil {
			return "", err
		}
		password.WriteByte(char)
	}

	// 填充剩余长度（字母）
	remaining := config.Length - config.MinSpecial - config.MinDigit
	for i := 0; i < remaining; i++ {
		pickLower := true
		if i%3 == 0 { // 随机选择大小写，但保持一定比例
			n, _ := rand.Int(rand.Reader, big.NewInt(2))
			pickLower = n.Int64() == 0
		}

		charSet := lowerChars
		if !pickLower {
			charSet = upperChars
		}

		char, err := randomChar(charSet)
		if err != nil {
			return "", err
		}
		password.WriteByte(char)
	}

	// 随机打乱结果
	result := []rune(password.String())
	shuffleRunes(result)

	return string(result), nil
}

// randInt 生成指定范围内的安全随机整数
func randInt(min, max int) (int, error) {
	if max <= min {
		return min, nil
	}

	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}
	return min + int(nBig.Int64()), nil
}

// randomChar 从字符集中随机选择一个字符
func randomChar(charSet string) (byte, error) {
	if len(charSet) == 0 {
		return 0, errors.New("字符集不能为空")
	}
	idx, err := randInt(0, len(charSet)-1)
	if err != nil {
		return 0, err
	}
	return charSet[idx], nil
}

// shuffleRunes 随机打乱字符顺序
func shuffleRunes(runes []rune) {
	n := len(runes)
	for i := n - 1; i > 0; i-- {
		j, _ := randInt(0, i)
		runes[i], runes[j] = runes[j], runes[i]
	}
}
