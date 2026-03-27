/**
 * @Author: 小鱼
 * @Date: 2024-04-18 13:22:35
 * @LastEditors: 小鱼
 * @LastEditTime: 2024-04-18 13:22:35
 * @FilePath: \passwordManageServer\pkg\encryption\ecdsaFunc.go
 * @Description: ECDSA加密模块，用于提供ECDSA密钥对生成等功能
 *
 * Copyright (c) 2024 by 小鱼, All Rights Reserved.
 */
package encryption

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

/**
 * @description: 生成 ECC 密钥对
 * @param {*}
 * @return {*ecdsa.PrivateKey, *ecdsa.PublicKey, error}
 */
func GenerateEccKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	curve := elliptic.P521()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	publicKey := &privateKey.PublicKey
	return privateKey, publicKey, nil
}

/**
 * @description: 导出ECC私钥
 * @param {*ecdsa.PrivateKey} privateKey
 * @return {string, error}
 */
func ExportEccPublicKey(publicKey *ecdsa.PublicKey) (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(publicKeyBytes), nil
}

/**
 * @description: 导入ECC公钥
 * @param {string} publicKeyBase64
 * @return {*ecdsa.PublicKey, error}
 */

func ImportEccPublicKey(publicKeyBase64 string) (*ecdsa.PublicKey, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return nil, err
	}
	pub, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		return nil, err
	}
	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid public key type")
	}
	return ecdsaPub, nil
}
