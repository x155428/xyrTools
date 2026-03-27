/**
 * @Author: 小鱼
 * @Date: 2024-04-18 13:22:35
 * @LastEditors: 小鱼
 * @LastEditTime: 2024-04-18 13:22:35
 * @FilePath: \passwordManageServer\winApi.go
 * @Description: winApi管理工具相关处理函数
 *
 * Copyright (c) 2024 by 小鱼, All Rights Reserved.
 */
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type API struct {
	Module       string `json:"module"`
	FunctionName string `json:"functionName"`
	FuncDefine   string `json:"funcDefine"`
	Signature    string `json:"signature"`
	FuncDesc     string `json:"funcDesc"`
	Other        string `json:"other"`
}

// queryWinApi 查询WinAPI信息
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func queryWinApi(w http.ResponseWriter, r *http.Request) {
	// 只允许GET请求
	if r.Method != http.MethodGet {
		http.Error(w, "不允许的请求方法！", http.StatusMethodNotAllowed)
		return
	}

	// 检查数据库
	if _, err := os.Stat(cfg.Database.SQLite.SqliteDbPath); os.IsNotExist(err) {
		err := initWinApiDb()
		if err != nil {
			http.Error(w, "数据库初始化失败", http.StatusInternalServerError)
			return
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("无数据库，已新建数据库！"))
			return
		}
	}

	module := r.URL.Query().Get("module")
	functionName := r.URL.Query().Get("functionName")
	if module == "" && functionName == "" {
		http.Error(w, "空查询！", http.StatusBadRequest)
		return
	}

	// 连接数据库
	db, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	isTableExists := `SELECT count(*) FROM sqlite_master WHERE type='table' AND name='winapi'`
	var count int
	err = db.QueryRow(isTableExists).Scan(&count)
	if err != nil {
		http.Error(w, "检查表是否存在异常：", http.StatusInternalServerError)
		return
	}
	if count == 0 {
		err := initWinApiDb()
		if err != nil {
			http.Error(w, "数据库初始化失败", http.StatusInternalServerError)
			return
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("无数据表，已新建！"))
			return
		}
	}

	// 构造查询语句
	query := "SELECT module, functionName, funcDefine, signature, funcDesc, other FROM winapi WHERE 1=1"
	var args []interface{}
	if module != "" {
		query += " AND module = ? COLLATE NOCASE"
		args = append(args, module)
	}
	if functionName != "" {
		query += " AND functionName = ? COLLATE NOCASE"
		args = append(args, functionName)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "查询出错", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result []API
	for rows.Next() {
		var api API
		err := rows.Scan(&api.Module, &api.FunctionName, &api.FuncDefine, &api.Signature, &api.FuncDesc, &api.Other)
		if err == nil {
			result = append(result, api)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// initWinApiDb 初始化WinAPI数据库
// 返回值：
// - 错误信息
func initWinApiDb() error {
	// 连接sqlite3数据库
	db, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		fmt.Println("[-]连接数据库失败:", err)
		return err
	}
	defer db.Close()
	// 检查表是否存在，如果不存在则创建
	createTableSQL := `CREATE TABLE IF NOT EXISTS winapi (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
    	module TEXT NOT NULL,
    	functionName TEXT NOT NULL,
    	funcDefine TEXT,
    	signature TEXT,
    	funcDesc TEXT,
    	other TEXT,
    	UNIQUE(module, functionName)
    );`
	_, err = db.Exec(createTableSQL) //执行sql语句
	if err != nil {
		fmt.Println("[-]创建表失败:", err)
		return err
	} //如果执行失败，打印错误信息并返回
	fmt.Println("[+]winapi数据库初始化完成！")
	return nil
}

// insertWinApi 插入或更新WinAPI信息
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func insertWinApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "不允许的请求方法！", http.StatusMethodNotAllowed)
		return
	}

	var api API
	err := json.NewDecoder(r.Body).Decode(&api)
	if err != nil {
		http.Error(w, "解析JSON数据失败", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// 插入或更新
	upsertSQL := `
	INSERT INTO winapi (module, functionName, funcDefine, signature, funcDesc, other)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(module, functionName) DO UPDATE SET
		funcDefine = excluded.funcDefine,
		signature = excluded.signature,
		funcDesc = excluded.funcDesc,
		other = excluded.other;
	`

	_, err = db.Exec(upsertSQL, api.Module, api.FunctionName, api.FuncDefine, api.Signature, api.FuncDesc, api.Other)
	if err != nil {
		http.Error(w, "插入或更新数据失败", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("插入或更新成功"))
}

// deleteWinApi 删除WinAPI信息
// 参数：
// - w: HTTP响应写入器
// - r: HTTP请求对象
func deleteWinApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "不允许的请求方法！", http.StatusMethodNotAllowed)
		return
	}
	module := r.URL.Query().Get("module")
	functionName := r.URL.Query().Get("functionName")
	// 连接数据库
	db, err := sql.Open("sqlite3", cfg.Database.SQLite.SqliteDbPath)
	if err != nil {
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	// 删除数据
	deleteSQL := "DELETE FROM winapi WHERE module = ? AND functionName = ?"
	_, err = db.Exec(deleteSQL, module, functionName)
	if err != nil {
		http.Error(w, "删除数据失败", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("删除成功"))

}
