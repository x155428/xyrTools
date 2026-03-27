/**
 * @Author: 小鱼
 * @Date: 2024-04-18 13:22:35
 * @LastEditors: 小鱼
 * @LastEditTime: 2024-04-18 13:22:35
 * @FilePath: \passwordManageServer\pkg\db\dataSaveToDb.go
 * @Description: 数据存储模块，负责将用户数据保存到数据库
 *
 * Copyright (c) 2024 by 小鱼, All Rights Reserved.
 */
package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"xyrTools/passwordManage/passwordManageServer/pkg/dataHandler"

	_ "github.com/mattn/go-sqlite3"
)

// SaveToDatabase 保存数据到数据库
// - 功能说明：将用户输入的数据保存到SQLite数据库中，包括创建表和插入数据
// - 参数：
//   - input: 输入数据对象，包含应用名称、用户名、密码等信息
//   - dbPath: 数据库文件路径
//
// - 返回值：
//   - 错误信息（如果保存失败）
func SaveToDatabase(input dataHandler.InputData, dbPath string) error {
	// 连接数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// 创建表
	createTable := `
	CREATE TABLE IF NOT EXISTS input_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		app_name TEXT,
		is_app_name_encrypted BOOLEAN,
		username TEXT,
		is_username_encrypted BOOLEAN,
		input_type TEXT,
		password TEXT,
		key_file TEXT,
		url TEXT,
		is_url_encrypted BOOLEAN,
		notes TEXT,
		is_notes_encrypted BOOLEAN,
		tags TEXT,
		is_tags_encrypted BOOLEAN,
		chose_encrypt TEXT,
		key TEXT
	);`
	var execErr error
	if _, execErr = db.Exec(createTable); execErr != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	// 将复杂字段序列化为 JSON，空值处理为 "null"
	appNameJSON := jsonOrNull(input.AppName)
	usernameJSON := jsonOrNull(input.Username)
	passwordJSON := jsonOrNull(input.Password)
	keyFileJSON := jsonOrNull(input.KeyFile)
	urlJSON := jsonOrNull(input.URL)
	notesJSON := jsonOrNull(input.Notes)
	tagsJSON := jsonOrNull(input.Tags)
	choseEncryptJSON := jsonOrNull(input.ChoseEncrypt)
	keyJSON := jsonOrNull(input.Key)

	// 插入数据
	insertSQL := `
	INSERT INTO input_data (
		app_name, is_app_name_encrypted, username, is_username_encrypted,
		input_type, password, key_file, url, is_url_encrypted,
		notes, is_notes_encrypted, tags, is_tags_encrypted,chose_encrypt,key
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	_, err = db.Exec(insertSQL,
		appNameJSON, input.IsAppNameEncrypted,
		usernameJSON, input.IsUsernameEncrypted,
		input.InputType, passwordJSON,
		keyFileJSON, urlJSON, input.IsUrlEncrypted,
		notesJSON, input.IsNotesEncrypted,
		tagsJSON, input.IsTagsEncrypted,
		choseEncryptJSON, keyJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to insert data: %v", err)
	}

	fmt.Println("Data inserted successfully!")
	return nil
}

// UpdateData 更新数据库中的数据
// - 功能说明：根据ID更新数据库中已有的记录
// - 参数：
//   - input: 包含更新数据的输入数据对象
//   - dbPath: 数据库文件路径
//   - id: 要更新的记录ID
//
// - 返回值：
//   - 错误信息（如果更新失败）
func UpdateData(input dataHandler.InputData, dbPath string, id int) error {
	// 连接数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// 将复杂字段序列化为 JSON，空值处理为 "null"
	appNameJSON := jsonOrNull(input.AppName)
	usernameJSON := jsonOrNull(input.Username)
	passwordJSON := jsonOrNull(input.Password)
	keyFileJSON := jsonOrNull(input.KeyFile)
	urlJSON := jsonOrNull(input.URL)
	notesJSON := jsonOrNull(input.Notes)
	tagsJSON := jsonOrNull(input.Tags)
	choseEncryptJSON := jsonOrNull(input.ChoseEncrypt)
	keyJSON := jsonOrNull(input.Key)

	// 根据id更新数据
	updateSQL := `
	UPDATE input_data SET
		app_name = ?, is_app_name_encrypted = ?,
		username = ?, is_username_encrypted = ?,
		input_type = ?, password = ?,
		key_file = ?, url = ?, is_url_encrypted = ?,
		notes = ?, is_notes_encrypted = ?,
		tags = ?, is_tags_encrypted = ?,
		chose_encrypt = ?, key = ?
	WHERE id = ?;`

	_, err = db.Exec(updateSQL,
		appNameJSON, input.IsAppNameEncrypted,
		usernameJSON, input.IsUsernameEncrypted,
		input.InputType, passwordJSON,
		keyFileJSON, urlJSON, input.IsUrlEncrypted,
		notesJSON, input.IsNotesEncrypted,
		tagsJSON, input.IsTagsEncrypted,
		choseEncryptJSON, keyJSON, id,
	)
	if err != nil {
		return fmt.Errorf("更新数据出错: %v", err)
	}
	return nil
}

// jsonOrNull 将数据序列化为JSON字符串或返回"null"
// - 功能说明：处理各种类型的数据，将其转换为JSON字符串表示，空值转换为"null"
// - 参数：
//   - v: 要序列化的任意类型数据
//
// - 返回值：
//   - 序列化后的JSON字符串或"null"
func jsonOrNull(v interface{}) string {
	// 如果传入的值是 nil，直接返回 "null"
	if v == nil {
		return "null"
	}

	// 检查是否是 EncryptedString 类型
	if encStr, ok := v.(dataHandler.EncryptedString); ok {
		// 如果encStr.Raw.Valid为true，则返回encStr.Raw.String
		if encStr.Raw.Valid {
			return encStr.Raw.String
		} else {
			// 如果encStr.Raw.Valid为false，则返回encStr.IV和encStr.Data的JSON字符串
			jsonData, err := encStr.MarshalJSON()
			if err != nil {
				return "null"
			}
			return string(jsonData)
		}
	}

	// 对其他类型，使用默认的 json.Marshal
	jsonData, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(jsonData)
}

// nullOrString 处理字符串，为空则返回nil
// - 功能说明：检查字符串是否为空，为空则返回nil，否则返回原字符串
// - 参数：
//   - s: 要检查的字符串
//
// - 返回值：
//   - 原字符串或nil
func nullOrString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// nullOrBool 处理布尔值
// - 功能说明：根据布尔值返回true或nil
// - 参数：
//   - b: 要处理的布尔值
//
// - 返回值：
//   - true或nil
func nullOrBool(b bool) interface{} {
	if b {
		return true
	}
	return nil
}

// SaveToDatabaseBatch 批量保存数据到数据库
// - 功能说明：将多个输入数据对象批量保存到数据库中，使用事务确保数据的原子性
// - 参数：
//   - inputs: 包含多个输入数据对象的切片
//   - dbPath: 数据库文件路径
//
// - 返回值：
//   - 错误信息（如果保存失败）
func SaveToDatabaseBatch(inputs []dataHandler.InputData, dbPath string) error {
	// 检查输入切片是否为空
	if len(inputs) == 0 {
		return nil
	}

	// 连接数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// 创建表（如果不存在）
	createTable := `
	CREATE TABLE IF NOT EXISTS input_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		app_name TEXT,
		is_app_name_encrypted BOOLEAN,
		username TEXT,
		is_username_encrypted BOOLEAN,
		input_type TEXT,
		password TEXT,
		key_file TEXT,
		url TEXT,
		is_url_encrypted BOOLEAN,
		notes TEXT,
		is_notes_encrypted BOOLEAN,
		tags TEXT,
		is_tags_encrypted BOOLEAN,
		chose_encrypt TEXT,
		key TEXT
	);`
	if _, err = db.Exec(createTable); err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// 准备插入语句
	insertSQL := `
	INSERT INTO input_data (
		app_name, is_app_name_encrypted, username, is_username_encrypted,
		input_type, password, key_file, url, is_url_encrypted,
		notes, is_notes_encrypted, tags, is_tags_encrypted,chose_encrypt,key
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	// 循环插入数据
	for _, input := range inputs {
		// 将复杂字段序列化为 JSON，空值处理为 "null"
		appNameJSON := jsonOrNull(input.AppName)
		usernameJSON := jsonOrNull(input.Username)
		passwordJSON := jsonOrNull(input.Password)
		keyFileJSON := jsonOrNull(input.KeyFile)
		urlJSON := jsonOrNull(input.URL)
		notesJSON := jsonOrNull(input.Notes)
		tagsJSON := jsonOrNull(input.Tags)
		choseEncryptJSON := jsonOrNull(input.ChoseEncrypt)
		keyJSON := jsonOrNull(input.Key)

		// 执行插入
		_, err = stmt.Exec(
			appNameJSON, input.IsAppNameEncrypted,
			usernameJSON, input.IsUsernameEncrypted,
			input.InputType, passwordJSON,
			keyFileJSON, urlJSON, input.IsUrlEncrypted,
			notesJSON, input.IsNotesEncrypted,
			tagsJSON, input.IsTagsEncrypted,
			choseEncryptJSON, keyJSON,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert data: %v", err)
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("事务提交失败: %v", err)
	}

	fmt.Printf("Successfully inserted %d records!\n", len(inputs))
	return nil
}
