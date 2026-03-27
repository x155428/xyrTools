/**
 * @Author: 小鱼
 * @Date: 2024-04-18 13:22:35
 * @LastEditors: 小鱼
 * @LastEditTime: 2024-04-18 13:22:35
 * @FilePath: \passwordManageServer\pkg\encryption\aesCryption.go
 * @Description: AES加密模块，用于提供AES加解密功能以及密钥生成
 *
 * Copyright (c) 2024 by 小鱼, All Rights Reserved.
 */
package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// GenerateAESKey_BaseInfo 生成AES加密密钥（基于用户信息和公钥信息）
// 参数：
// - publicKey: RSA公钥
// - customInfo: 自定义信息数组
// 返回值：
// - 32字节的AES密钥
// - 错误信息
func GenerateAESKey_BaseInfo(publicKey *rsa.PublicKey, customInfo []string) ([]byte, error) {
	// 将公钥转换为字节数据，使用较短的公钥部分（比如前 64 字节）
	publicKeyBytes := x509.MarshalPKCS1PublicKey(publicKey)
	publicKeyHash := sha256.Sum256(publicKeyBytes) // 对公钥哈希
	// 排序 customInfo 数据，确保顺序一致
	sort.Strings(customInfo)
	// 序列化 customInfo 数组
	customInfoBytes, err := json.Marshal(customInfo)
	if err != nil {
		return nil, fmt.Errorf("序列化 CustomInfo 失败: %v", err)
	}

	// 对 customInfo 进行哈希处理
	customInfoHash := sha256.Sum256(customInfoBytes)

	// 控制比重，公钥哈希，customInfo 哈希
	combinedData := append(publicKeyHash[:20], customInfoHash[:32]...) // 合并选定字节数

	// 使用 SHA-256 对合并后的数据进行哈希处理，生成 AES 密钥
	aesKey := sha256.Sum256(combinedData)

	// 返回 32 字节 AES 密钥
	return aesKey[:], nil
}

/**
 * @description: 生成 AES 密钥（基于随机数生成指定字节数据作为密钥）
 * @param {int} keySize
 * @return {字节转十六进制字符串}
 */
func GenerateAESKey_BaseRand(keySize int) (string, error) {
	key := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("生成 AES 密钥失败: %w", err)
	}
	// 将密钥转换为十六进制字符串，以便存储或传输
	return hex.EncodeToString(key), nil
}

// AesDecryptData AES解密数据
// 参数：
// - nonce: 随机数
// - encryptedData: 加密数据
// - key: AES密钥
// 返回值：
// - 解密后的明文
// - 错误信息
func AesDecryptData(nonce, encryptedData []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(encryptedData) < aes.BlockSize {
		return nil, fmt.Errorf("密文过短")
	}

	mode, err := cipher.NewGCMWithNonceSize(block, 12)
	if err != nil {
		return nil, err
	}

	if len(nonce) != mode.NonceSize() {
		return nil, fmt.Errorf("Nonce 长度不正确")
	}

	plaintext, err := mode.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// AesEncryptData AES加密数据
// 参数：
// - plaintext: 明文数据
// - key: AES密钥
// 返回值：
// - 随机数(nonce)
// - 加密后的密文
// - 错误信息
func AesEncryptData(plaintext, key []byte) ([]byte, []byte, error) {
	// 创建AES加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	// 使用GCM模式
	mode, err := cipher.NewGCMWithNonceSize(block, 12) // 12字节的Nonce
	if err != nil {
		return nil, nil, err
	}

	// 生成12字节的随机Nonce
	nonce := make([]byte, mode.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	zeroNonce := make([]byte, mode.NonceSize())
	if bytes.Equal(nonce, zeroNonce) {
		nonce = make([]byte, mode.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, nil, err
		}
	}

	// 使用AES-GCM加密明文
	ciphertext := mode.Seal(nil, nonce, plaintext, nil)

	// 分别返回Nonce和密文
	return nonce, ciphertext, nil
}
