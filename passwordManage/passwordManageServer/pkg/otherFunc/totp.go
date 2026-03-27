/*
 * @Author: 小鱼
 * @Date: 2025-09-29 14:00:00
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-10-13 17:27:24
 * @FilePath: \passwordManageServer\pkg\otherFunc\totp.go
 * @Description: TOTP (基于时间的一次性密码) 相关功能实现，包含密钥生成和验证等
 *
 * Copyright (c) 2025 by 小鱼, All Rights Reserved.
 */
package otherFunc

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"strings"
	"time"

	"rsc.io/qr"
)

/**
 * @description: 生成 TOTP码
 * @param {string} secretKey 密钥
 * @param {int64} interval 周期
 * @param {int} digits 位数
 * @return {*}
 */
func GenerateTOTP(secretKey string, interval int64, digits int) (string, error) {
	// 解析并解码密钥
	// TOTP标准要求Base32密钥大写
	secretKey = strings.ToUpper(secretKey)
	key, err := base32.StdEncoding.DecodeString(secretKey)
	if err != nil {
		return "", err
	}

	// 计算当前时间片
	timeCounter := time.Now().Unix() / interval

	// HOTP规范要求时间片转换为8字节大端格式
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, uint64(timeCounter))

	// 使用 HMAC-SHA1 生成哈希值
	hash := hmac.New(sha1.New, key)
	hash.Write(counterBytes)
	hmacHash := hash.Sum(nil)

	// 动态截取，获取哈希的特定部分
	// 提取低四位，偏移量
	offset := hmacHash[len(hmacHash)-1] & 0x0F
	// 从偏移量开始提取4字节，确保正整数
	code := (int(hmacHash[offset]&0x7F) << 24) |
		(int(hmacHash[offset+1]&0xFF) << 16) |
		(int(hmacHash[offset+2]&0xFF) << 8) |
		(int(hmacHash[offset+3] & 0xFF))

	// 取最后x位数字
	mod := int32(math.Pow10(digits))
	otp := int32(code) % mod
	return fmt.Sprintf("%0*d", digits, otp), nil
}

/**
 * @description: 生成 TOTP URI
 * @param {*} secretKey 密钥
 * @param {*} issuer 应用名
 * @param {string} accountName  用户名
 * @param {*} digits 位数
 * @param {int} period 刷新周期
 * @return {*}
 */
func GenerateTotpURI(secretKey, issuer, accountName string, digits, period int) string {
	//密钥，应用名，用户名，密钥，发行者，位数，刷新周期s
	// Base32 编码的密钥，去掉填充的 "=" 号
	secretKey = strings.ToUpper(strings.TrimRight(secretKey, "="))

	// TOTP URI 格式
	uri := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&digits=%d&period=%d",
		issuer, accountName, secretKey, issuer, digits, period)
	return uri
}

/**
 * @description: 生成二维码图像并返回其PNG格式的二进制数据
 * @param {*} uri  totp的url
 * @return {[]byte} PNG格式的二进制数据
 * @return {error} 错误信息
 */
func GenerateQRCode(uri string) ([]byte, error) {
	//二维码图像放大倍数
	scale := 5
	// 使用 rsc.io/qr 库生成二维码
	code, err := qr.Encode(uri, qr.Q)
	if err != nil {
		return nil, err
	}

	// 获取二维码的原始尺寸
	originalSize := code.Size

	// 创建新图像，尺寸放大
	img := image.NewRGBA(image.Rect(0, 0, originalSize*scale, originalSize*scale))
	for y := 0; y < originalSize; y++ {
		for x := 0; x < originalSize; x++ {
			// 获取二维码像素颜色
			color := code.Image().At(x, y)
			// 将每个像素按比例填充到新图像
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					img.Set(x*scale+dx, y*scale+dy, color)
				}
			}
		}
	}
	// 创建内存缓冲区存储PNG数据
	var buf bytes.Buffer
	// 将二维码图像编码为PNG格式并写入缓冲区
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	// 返回PNG格式的二进制数据
	return buf.Bytes(), nil
}

/**
 * @description: 生成一个随机的 TOTP 密钥，返回 Base32 编码的密钥字符串
 * @return {*}
 */
func GenerateTOTPKey() (string, error) {
	// 生成 20 字节的随机数
	key := make([]byte, 20)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	// 使用 'A-Z2-7' 的 Base32 编码
	encodedKey := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(key)
	return encodedKey, nil
}

// 验证totp设置状态
func IsTotpSetting(sqliteDB *sql.DB, userName string) bool {
	// 使用 sql.NullInt64 处理可能为 NULL 的列
	var totpEnabled, totpSetupStatus sql.NullInt64

	err := sqliteDB.QueryRow(
		"SELECT totp_enabled, totp_setup_status FROM userMeta_data WHERE username = ?",
		userName,
	).Scan(&totpEnabled, &totpSetupStatus)

	if err != nil {
		// 查询失败或用户不存在，认为未设置
		return false
	}

	// NULL 当作 0 处理
	enabled := totpEnabled.Valid && totpEnabled.Int64 == 1
	setup := totpSetupStatus.Valid && totpSetupStatus.Int64 == 1

	return enabled && setup
}

// 验证totp
func VerifyTotp(totpCode, userName string, dbPath string) bool {
	// 连接数据库
	sqliteDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("[-]数据库连接失败: %v", err)
	}
	defer sqliteDB.Close()

	// 从数据库获取用户的 TOTP配置secretKey string, interval int64, digits int
	var secretKey string
	var interval int64
	var digits int
	// 连接数据库查询用户的 TOTP配置
	err = sqliteDB.QueryRow(
		"SELECT totp_secret, totp_refresh_period, totp_digits FROM userMeta_data WHERE username = ?",
		userName,
	).Scan(&secretKey, &interval, &digits)
	if err != nil {
		return false
	}
	// 生成totp码
	totp, err := GenerateTOTP(secretKey, interval, digits)
	if err != nil {
		return false
	}
	return totp == totpCode

}
