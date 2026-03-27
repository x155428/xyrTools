/**
 * @Author: 小鱼
 * @Date: 2024-04-18 13:22:35
 * @LastEditors: 小鱼
 * @LastEditTime: 2024-04-18 13:22:35
 * @FilePath: \passwordManageServer\pkg\db\initDb.go
 * @Description: 系统启动时初始化检查数据库信息，创建必要的表结构
 *
 * Copyright (c) 2024 by 小鱼, All Rights Reserved.
 */
package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB 初始化数据库
// - 功能说明：检查数据库文件是否存在，不存在则创建并初始化，存在则检查连接和表结构
// - 参数：
//   - dbPath: 数据库文件路径
//   - sqlScriptPath: SQL初始化脚本文件路径
//
// - 注意：如果操作失败，程序会自动退出
func InitDB(dbPath, sqlScriptPath string) {
	//判断数据库文件是否存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("----------------------------------------")
		fmt.Printf("数据库文件 [%s] 不存在！", dbPath)
		fmt.Println("请选择操作：")
		fmt.Println("n - 结束程序，检查数据库路径配置")
		fmt.Println("y - 直接新建数据库")
		fmt.Println("----------------------------------------")
		fmt.Print("请输入选择 (n/y): ")

		var choice string
		fmt.Scanln(&choice)

		if choice == "n" || choice == "N" {
			fmt.Println("[-]请检查数据库路径，3秒后程序退出。")
			time.Sleep(3 * time.Second)
			os.Exit(1)
		}
		fmt.Println("[+]正在新建数据库...")
		// 创建数据库文件和表项
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			fmt.Printf("[-]打开数据库出错: %v\n", err)
			time.Sleep(3 * time.Second)
			os.Exit(1)
		}
		defer db.Close()
		// 读取 SQL 文件
		sqlScript, err := os.ReadFile(sqlScriptPath)
		//fmt.Println(string(sqlScript))
		if err != nil {
			fmt.Printf("[-]读取sql脚本文件出错: %v\n", err)
			time.Sleep(3 * time.Second)
			os.Exit(1)
		}
		_, err = db.Exec(string(sqlScript))
		if err != nil {
			fmt.Printf("[-]执行sql出错: %v\n", err)
			time.Sleep(3 * time.Second)
			os.Exit(1)
		}
		fmt.Println("[+]数据库初始化完成！")
	} else {
		fmt.Println("[+]数据库文件存在，开始初始化检查！")
		// 连接数据库
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			fmt.Printf("[-]打开数据库出错: %v\n", err)
			time.Sleep(3 * time.Second)
			os.Exit(1)
		}
		defer db.Close()
		// 检查数据库连接
		err = db.Ping()
		if err != nil {
			fmt.Printf("[-]连接数据库出错: %v\n", err)
			time.Sleep(3 * time.Second)
			os.Exit(1)
		}

		//检查数据库中是否存在数据表
		hasInputData, _ := TableExists(db, "input_data")
		//fmt.Println(hasInputData)
		// 表不存在，执行 SQL 脚本创建表（暂时这么搞，后续细化）
		if !hasInputData {
			// 读取 SQL 文件
			sqlScript, err := os.ReadFile(sqlScriptPath)
			//fmt.Println(string(sqlScript))
			if err != nil {
				fmt.Printf("[-]读取sql脚本文件出错: %v\n", err)
				time.Sleep(3 * time.Second)
				os.Exit(1)
			}
			_, err = db.Exec(string(sqlScript))
			if err != nil {
				fmt.Printf("[-]执行sql出错: %v\n", err)
				time.Sleep(3 * time.Second)
				os.Exit(1)
			}
			fmt.Println("[+]数据库检查完毕！")
		}
		fmt.Println("[+]数据库检查完毕！")
	}

}

// CheckTableIntegrity 检查数据库表结构是否完整
// - 功能说明：检查指定的所有表是否存在以及表中是否包含所有必需的列
// - 参数：
//   - db: 数据库连接对象
//   - tableInfo: 表结构信息映射，键为表名，值为该表必需的列名数组
//
// - 返回值：
//   - 布尔值，表示表结构是否完整
//   - 错误信息（如果检查失败或表结构不完整）
// func CheckTableIntegrity(db *sql.DB, tableInfo map[string][]string) (bool, error) {
// 	for tableName, columns := range tableInfo {
// 		// 检查表是否存在
// 		tableExists, err := TableExists(db, tableName)
// 		if err != nil {
// 			return false, err
// 		}
// 		if !tableExists {
// 			return false, fmt.Errorf(" %s 表不存在", tableName)
// 		}
// 		columnExists, err := ColumnExists(db, tableName, columns...)
// 	}
// 	return true, nil
// }
