/**
 * @Author: 小鱼
 * @Date: 2024-04-18 13:22:35
 * @LastEditors: 小鱼
 * @LastEditTime: 2024-04-18 13:22:35
 * @FilePath: \passwordManageServer\pkg\encryption\rsaCryption.go
 * @Description: RSA加密模块，用于提供RSA加解密功能以及密钥对生成
 *
 * Copyright (c) 2024 by 小鱼, All Rights Reserved.
 */
package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

/**
 * @description: 用于加密给定的消息，返回加密后的密文
 * @param {*rsa.PublicKey} publicKey
 * @param {[]byte} message
 * @return {*}
 */
func EncryptMessageRsa(publicKey *rsa.PublicKey, message []byte) ([]byte, error) {
	label := []byte("")
	hash := sha256.New()

	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, publicKey, message, label)
	if err != nil {
		return nil, fmt.Errorf("error encrypting message: %v", err)
	}

	return ciphertext, nil
}

/**
 * @description: 用于使用 RSA 私钥解密密文
 * @param {*rsa.PrivateKey} privateKey
 * @param {[]byte} ciphertext
 * @return {*}
 */
func DecryptMessageRsa(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	hash := sha256.New()
	label := []byte("")

	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, ciphertext, label)
	if err != nil {
		return nil, fmt.Errorf("error decrypting message: %v", err)
	}

	return plaintext, nil
}

/**
 * @description: 生成 RSA 密钥对
 * @param {*}
 * @return {rsa.PrivateKey, rsa.PublicKey, error}
 */
func GenerateRsaKeyPair_BaseRand() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	publicKey := &privateKey.PublicKey
	return privateKey, publicKey, nil
}

/**
 * @description: 保存私钥到文件
 * @param {*rsa.PrivateKey} privateKey
 * @param {string} filename
 * @return {*}
 */
func SaveRsaPrivateKeyToFile(privateKey *rsa.PrivateKey, filename string) error {
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	return os.WriteFile(filename, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privKeyBytes}), 0600)
}

/**
 * @description: 从文件中加载私钥
 * @param {string} filename
 * @return {*}
 */
func LoadRsaPrivateKeyFromFile(filename string) (*rsa.PrivateKey, error) {
	privKeyPEM, err := os.ReadFile(filename) // 使用 os.ReadFile
	if err != nil {
		return nil, fmt.Errorf("error reading private key file: %v", err)
	}

	block, _ := pem.Decode(privKeyPEM)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("invalid private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %v", err)
	}

	return privateKey, nil
}

/**
 * @description: 私钥转pem字符串
 * @param {*rsa.PublicKey} publicKey
 * @param {string} filename
 * @return {*}
 */
func PrivateKeyToPEM_Str(privateKey *rsa.PrivateKey) string {
	// 使用 x509 标准将 RSA 私钥编码为 DER 格式
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)

	// 使用 pem 包装为 PEM 格式
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}

	// 返回 PEM 格式字符串
	return string(pem.EncodeToMemory(block))
}

/**
 * @description: 公钥转pem字符串
 * @param {*rsa.PublicKey} publicKey
 * @param {string} filename
 * @return {*}
 */

func PublicKeyToPEM_Str(publicKey *rsa.PublicKey) string {
	// 使用 x509 标准将 RSA 公钥编码为 DER 格式
	derStream, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		fmt.Println("Error converting public key to PEM:", err)
		return ""
	}

	// 使用 pem 包装为 PEM 格式
	block := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: derStream,
	}

	// 返回 PEM 格式字符串
	return string(pem.EncodeToMemory(block))
}
