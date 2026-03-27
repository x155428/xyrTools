/**
 * @Author: 小鱼
 * @Date: 2024-04-18 13:22:35
 * @LastEditors: 小鱼
 * @LastEditTime: 2024-04-18 13:22:35
 * @FilePath: \passwordManageServer\pkg\encryption\specialKeyGen.go
 * @Description: 特殊密钥生成模块，用于生成AES密钥等特殊用途的密钥
 *
 * Copyright (c) 2024 by 小鱼, All Rights Reserved.
 */
package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
)

type AesKeyCfg struct {
	Length int `json:"length,omitempty"`
}

// GenerateAESKey 生成AES密钥
func GenerateAESKey(KeyCfg string) (string, error) {
	var aesKeyCfg AesKeyCfg
	if KeyCfg != "" {
		if err := json.Unmarshal([]byte(KeyCfg), &aesKeyCfg); err != nil {
			return "", fmt.Errorf("JSON解析失败: %w", err)
		}
	}
	if aesKeyCfg.Length == 0 {
		aesKeyCfg.Length = 256
	}

	validLengths := map[int]struct{}{128: {}, 192: {}, 256: {}}
	if _, ok := validLengths[aesKeyCfg.Length]; !ok {
		return "", errors.New("无效的AES密钥长度，必须是128、192或256位")
	}

	key := make([]byte, aesKeyCfg.Length/8)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("生成AES密钥失败: %w", err)
	}
	return hex.EncodeToString(key), nil
}

// GenerateRSAKeyPair 生成RSA密钥对
func GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	if bits != 2048 && bits != 4096 {
		return nil, nil, errors.New("RSA密钥长度必须是2048或4096位")
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("生成RSA密钥失败: %w", err)
	}
	return privateKey, &privateKey.PublicKey, nil
}

// ExportPrivateKeyToPEM 将私钥导出为PEM格式
func ExportPrivateKeyToPEM(privateKey *rsa.PrivateKey) ([]byte, error) {
	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})
	return privPEM, nil
}

// ExportPublicKeyToPEM 将公钥导出为PEM格式
func ExportPublicKeyToPEM(publicKey *rsa.PublicKey) ([]byte, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("公钥编码失败: %w", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})
	return pubPEM, nil
}

// PEMToPublicKey 将PEM格式的公钥字符串转换为*rsa.PublicKey对象
func PEMToPublicKey(pemStr string) (*rsa.PublicKey, error) {
	// 解码PEM格式字符串
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("无效的PEM格式")
	}
	
	// 解析公钥
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败: %w", err)
	}
	
	// 类型断言为RSA公钥
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("不是RSA公钥")
	}
	
	return rsaPub, nil
}
