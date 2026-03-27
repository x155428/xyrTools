/*
 * @Author: 小鱼
 * @Date: 2025-03-04 15:46:53
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-03-05 17:52:34
 * @FilePath: \passwordManage\passwordManageServer\pkg\tools\passwordOverview.go
 * @Description: 密码概览工具，提供密码统计数据、强度计算和数据库更新功能
 *
 * Copyright (c) 2025 by 小鱼, All Rights Reserved.
 */
package tools

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"xyrTools/passwordManage/passwordManageServer/pkg/dataStructs"
	"xyrTools/passwordManage/passwordManageServer/pkg/encryption"

	_ "github.com/mattn/go-sqlite3"
)

// AnalyzePasswordStrength 传入数据库路径、表名称和密钥，对密码进行强度评估并更新到数据库
func AnalyzePasswordStrength(dbPath, tableName string, mainKeyHexStr string) (map[string]interface{}, error) {
	// 检查mainKeyHexStr是否为空
	if mainKeyHexStr == "" {
		return nil, fmt.Errorf("主密钥未配置！")
	}
	// mainKeyHexStr转成[]byte
	mainKeyByte, err := hex.DecodeString(mainKeyHexStr)
	if err != nil {
		return nil, fmt.Errorf("主密钥转换失败: %w", err)
	}

	// 打开数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("无法打开数据库: %w", err)
	}
	defer db.Close()

	// 测试数据库连接
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("无法连接到数据库: %w", err)
	}

	// 开启事务,事务管理数据库操作
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("启动事务失败: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback() // 回滚失败
			panic(p)          // 重新抛出 panic
		}
		if err != nil {
			_ = tx.Rollback() // 出现错误时回滚
			return
		}
		// 只有在没有错误时才提交事务
		if commitErr := tx.Commit(); commitErr != nil {
			log.Printf("事务提交失败: %v", commitErr) // 提交失败时记录日志
		}
	}()

	// 查询密码总条数
	var totalCount int
	err = tx.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("查询总条数时出错: %w", err)
	}
	log.Printf("数据库 %s 中表 %s 的总记录数为: %d", dbPath, tableName, totalCount)

	// 查询表结构，检查是否存在 strength 列
	rows, err := tx.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return nil, fmt.Errorf("查询表结构时出错: %w", err)
	}
	defer rows.Close()

	columnsMap := make(map[string]struct{})
	for rows.Next() {
		var cid int
		var name, datatype string
		var notnull, primaryKey int
		var defaultVal sql.NullString

		if err = rows.Scan(&cid, &name, &datatype, &notnull, &defaultVal, &primaryKey); err != nil {
			return nil, fmt.Errorf("解析列信息时出错: %w", err)
		}
		columnsMap[name] = struct{}{}
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历表结构时出错: %w", err)
	}
	rows.Close()

	// 如果 strength 列不存在，则创建
	if _, exists := columnsMap["strength"]; !exists {
		_, err = tx.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN strength TEXT", tableName))
		if err != nil {
			return nil, fmt.Errorf("创建 strength 列时出错: %w", err)
		}
	}

	// 查询所有密码记录
	rows, err = tx.Query(fmt.Sprintf("SELECT id, password, input_type, chose_encrypt, key FROM %s", tableName))
	if err != nil {
		return nil, fmt.Errorf("查询数据时出错: %w", err)
	}
	defer rows.Close()

	// 遍历所有记录并更新密码强度
	for rows.Next() {
		var id int
		var password, inputType, choseEncrypt, key string

		err = rows.Scan(&id, &password, &inputType, &choseEncrypt, &key)
		if err != nil {
			return nil, fmt.Errorf("扫描行数据时出错: %w", err)
		}

		// 仅对密码类型的数据进行解密
		if inputType == "password" && password != "" {
			// 解密 key
			var keyData dataStructs.EncryptedUserDataDB
			if err = json.Unmarshal([]byte(key), &keyData); err != nil {
				log.Printf("解析 key 失败: %v", err)
				continue
				//return nil, fmt.Errorf("解析 key 失败: %w", err)
			}

			decryptKeyTmp, err := encryption.AesDecryptData(keyData.IV, keyData.Data, mainKeyByte)
			if err != nil {
				log.Printf("解密 key 失败: %v", err)
				continue
				//return nil, fmt.Errorf("解密 key 失败: %w", err)
			}
			decryptKey, err := hex.DecodeString(string(decryptKeyTmp))
			if err != nil {
				log.Printf("key 转换失败: %v", err)
				continue
				//return nil, fmt.Errorf("key 转换失败: %w", err)
			}

			// 解密 choseEncrypt
			var choseEncryptData dataStructs.EncryptedUserDataDB
			if err = json.Unmarshal([]byte(choseEncrypt), &choseEncryptData); err != nil {
				log.Printf("解析 choseEncrypt 失败: %v", err)
				continue
				//return nil, fmt.Errorf("解析 choseEncrypt 失败: %w", err)
			}

			choseEncryptBytes, err := encryption.AesDecryptData(choseEncryptData.IV, choseEncryptData.Data, mainKeyByte)
			if err != nil {
				log.Printf("解密 choseEncrypt 失败: %v", err)
				continue
				//return nil, fmt.Errorf("解密 choseEncrypt 失败: %w", err)
			}
			choseEncryptStr := string(choseEncryptBytes)

			// 选择加密算法
			if choseEncryptStr == "" {
				choseEncryptStr = "AES-GCM"
				decryptKey = mainKeyByte
			}

			//log.Printf("本条记录加密算法: %s", choseEncryptStr)

			// 解密 password
			var passwordData dataStructs.EncryptedUserDataDB
			if err = json.Unmarshal([]byte(password), &passwordData); err != nil {
				log.Printf("解析 password 失败: %v", err)
				continue
				//return nil, fmt.Errorf("解析 password 失败: %w", err)
			}

			var decryptedPassword string
			// TODO: 完善解密算法
			switch choseEncryptStr {
			case "AES-GCM":
				decryptedPasswordBytes, err := encryption.AesDecryptData(passwordData.IV, passwordData.Data, decryptKey)
				if err != nil {
					log.Printf("解密密码失败: %v", err)
					continue
					//return nil, fmt.Errorf("解密密码失败: %w", err)
				}
				decryptedPassword = string(decryptedPasswordBytes)
				//fmt.Printf("解密后的密码: %s\n", decryptedPassword)

			case "ChaCha20-Poly1305":
				{
				}
			default:
				log.Printf("不支持的加密算法: %s", choseEncryptStr)
				continue
				//return nil, fmt.Errorf("不支持的加密算法: %s", choseEncryptStr)
			}

			// 计算密码强度
			strength := EvaluateStrength(decryptedPassword)

			// 拼接强度信息和评分
			strengthMessageWithScore := fmt.Sprintf("%s (%d)", strength.StrengthMessage, strength.Score)

			// 更新数据库
			_, err = tx.Exec(fmt.Sprintf("UPDATE %s SET strength = ? WHERE id = ?", tableName), strengthMessageWithScore, id)
			if err != nil {
				log.Printf("更新 strength 列失败: %v", err)
				continue
				//return nil, fmt.Errorf("更新 strength 列失败: %w", err)
			}
		}
	}

	// 确保 `rows` 遍历完成
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历数据库行时出错: %w", err)
	}
	// 提交完事务后进行查询，查询 strength 列，强度为非常弱的记录个数
	var veryWeakCount int
	err = tx.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE strength LIKE '非常弱%%'", tableName)).Scan(&veryWeakCount)
	if err != nil {
		return nil, fmt.Errorf("查询弱口令数量记录时出错: %w", err)
	}
	//log.Printf("数据库 %s 中表 %s 的非常弱口令数量为: %d", dbPath, tableName, veryWeakCount)

	// 查询各个强度的数量
	strengths := []string{"非常弱", "弱", "一般", "强", "非常强"}
	strengthCounts := make(map[string]int)

	for _, strength := range strengths {
		var count int
		err = tx.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE strength LIKE '%s%%'", tableName, strength)).Scan(&count)
		if err != nil {
			if err == sql.ErrNoRows {
				strengthCounts[strength] = 0 // 如果没有记录，则将计数设置为0
				continue
			} else {
				log.Printf("查询 %s 强度记录数量时出错: %v", strength, err)
				continue
			}
			//return nil, fmt.Errorf("查询 %s 强度记录数量时出错: %w", strength, err)
		}
		strengthCounts[strength] = count
		//log.Printf("数据库 %s 中表 %s 的 %s 强度口令数量为: %d", dbPath, tableName, strength, count)
	}

	// 计算有效记录的总数
	totalValidCount := 0
	for _, count := range strengthCounts {
		totalValidCount += count
	}

	// 计算各个强度所占总数的百分比
	strengthPercentages := make(map[string]float64)
	for strength, count := range strengthCounts {
		if totalValidCount > 0 {
			percentage := float64(count) / float64(totalValidCount) * 100
			strengthPercentages[strength] = percentage
			//log.Printf("数据库 %s 中表 %s 的 %s 强度口令所占百分比为: %.2f%%", dbPath, tableName, strength, percentage)
		}
	}

	// 计算网站记录总数
	var validUrlAndPasswordCount int
	err = tx.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE url IS NOT NULL AND url != '' AND password IS NOT NULL AND password != ''", tableName)).Scan(&validUrlAndPasswordCount)
	if err != nil {
		return nil, fmt.Errorf("查询 url 字段不为空且密码记录不为空的数量时出错: %w", err)
	}
	//log.Printf("数据库 %s 中表 %s 的 url 字段不为空且密码记录不为空的数量为: %d", dbPath, tableName, validUrlAndPasswordCount)

	// 组装数据
	responseData := map[string]interface{}{
		"stats": []map[string]string{
			{
				"title": "记录总数",
				// 将 totalCount 转换为字符串
				"value": fmt.Sprintf("%d", totalCount),
			},
			{
				"title": "网站记录数量",
				// 这里也可以用实际的 validUrlAndPasswordCount 来替代硬编码的 "30"
				"value": fmt.Sprintf("%d", validUrlAndPasswordCount),
			},
			{
				"title": "弱口令",
				// 这里可以用实际的 veryWeakCount 来替代硬编码的 "10"
				"value": fmt.Sprintf("%d", veryWeakCount),
			},
			{
				// TODO: 查询count计算最高
				"title": "使用率最高",
				"value": "example.com",
			},
		},
		"strength": []map[string]interface{}{
			{
				"label":      "非常弱",
				"percentage": strengthPercentages["非常弱"],
			},
			{
				"label":      "弱",
				"percentage": strengthPercentages["弱"],
			},
			{
				"label":      "一般",
				"percentage": strengthPercentages["一般"],
			},
			{
				"label":      "强",
				"percentage": strengthPercentages["强"],
			},
			{
				"label":      "非常强",
				"percentage": strengthPercentages["非常强"],
			},
		},
	}

	return responseData, nil
}
