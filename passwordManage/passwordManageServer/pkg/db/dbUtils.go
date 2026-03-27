package db

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var safeIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// TableExists 检查数据库中是否存在指定表
// 参数：
// - db: 数据库连接对象
// - tableName: 表名
// 返回值：
// - 布尔值，表示表是否存在
// - 错误信息
func TableExists(db *sql.DB, tableName string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name =?", tableName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ColumnExists 检查数据库表中是否存在指定列
// - 参数：
//   - db: 数据库连接对象
//   - tableName: 表名
//   - columnNames: 要检查的列名
//
// - 返回值：
//   - 不存在的列名数组
//   - 错误信息
func ColumnsExist(db *sql.DB, tableName string, columnNames []string) (noColumns []string, err error) {

	if db == nil {
		return nil, errors.New("db is nil")
	}
	if tableName == "" || len(columnNames) == 0 {
		return nil, errors.New("invalid arguments")
	}

	// 强校验：只允许安全标识符
	if !safeIdentifier.MatchString(tableName) {
		return nil, fmt.Errorf("invalid table name: %q", tableName)
	}
	for _, columnName := range columnNames {
		// 提取列名，忽略空格和约束
		parts := strings.Fields(columnName)
		if len(parts) == 0 {
			continue // 空字符串或仅包含空格，跳过
		}
		colName := parts[0]

		// 校验列名是否安全
		if !safeIdentifier.MatchString(colName) {
			return nil, fmt.Errorf("invalid column name: %q", colName)
		}
	}

	query := fmt.Sprintf("PRAGMA table_info('%s')", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 收集数据库中实际存在的所有列名
	existingColumns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notnull int
		var dfltValue interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &dataType, &notnull, &dfltValue, &pk); err != nil {
			return nil, err
		}
		existingColumns[strings.ToLower(name)] = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 遍历传入的 columnNames，只保留那些在实际列名中不存在的列名
	for _, columnName := range columnNames {
		// 提取空格分隔的第一个单词作为列名
		parts := strings.Fields(columnName)
		if len(parts) == 0 {
			continue // 空字符串或仅包含空格，跳过
		}
		colName := parts[0]

		if !existingColumns[strings.ToLower(colName)] {
			noColumns = append(noColumns, columnName)
		}
	}

	return noColumns, nil
}

// CreateTable 创建指定表
// - 参数：
//   - db: 数据库连接对象
//   - tableName: 表名
//   - columns: 表列定义数组，每个元素为列名和数据类型，例如："id INTEGER PRIMARY KEY", "name TEXT"
//
// - 返回值：
//   - 错误信息（如果创建失败）
func CreateTable(db *sql.DB, tableName string, columns []string) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if tableName == "" || len(columns) == 0 {
		return errors.New("invalid arguments")
	}

	// 检查表名是否合法
	if !safeIdentifier.MatchString(tableName) {
		return fmt.Errorf("invalid table name: %q", tableName)
	}

	// 对每个列定义进行校验：只允许“列名 + 类型 [+约束]”
	validatedColumns := make([]string, 0, len(columns))
	for _, col := range columns {
		col = strings.TrimSpace(col)
		if col == "" {
			return errors.New("empty column definition")
		}

		// 拆出列名（第一个空格前）
		parts := strings.Fields(col)
		if len(parts) == 0 {
			return fmt.Errorf("invalid column definition: %q", col)
		}
		colName := parts[0]

		if !safeIdentifier.MatchString(colName) {
			return fmt.Errorf("invalid column name: %q", colName)
		}

		// 限制类型声明只允许特定关键字
		if len(parts) > 1 {
			typeDecl := strings.ToUpper(parts[1])
			if !strings.Contains("INTEGER REAL TEXT BLOB NUMERIC", typeDecl) {
				return fmt.Errorf("invalid type in column %q: %s", colName, typeDecl)
			}
		}

		validatedColumns = append(validatedColumns, col)
	}

	columnStr := strings.Join(validatedColumns, ", ")

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, columnStr)

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("create table failed: %w", err)
	}
	return nil
}

// AddColumn 向指定表中插入列
// - 参数：
//   - db: 数据库连接对象
//   - tableName: 表名
//   - columns: 列定义数组，每个元素为列名和数据类型，例如：["name TEXT", "age INTEGER"]
//
// - 返回值：
//   - 错误信息（如果插入失败）
func AddColumns(db *sql.DB, tableName string, columns []string) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if tableName == "" || len(columns) == 0 {
		return errors.New("invalid arguments")
	}
	if !safeIdentifier.MatchString(tableName) {
		return fmt.Errorf("invalid table name: %q", tableName)
	}
	// 检查列名是否合法
	for _, column := range columns {
		column = strings.TrimSpace(column)
		if column == "" {
			return errors.New("empty column definition")
		}
	}
	// 插入列
	for _, column := range columns {
		query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, column)
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to add column %q: %w", column, err)
		}
	}
	return nil
}

// 判断数据库中某些列是否存在，不存在尝试新建
// - 参数：
//   - db: 数据库连接对象
//   - tableName: 表名
//   - columns: 列名数组
//
// - 返回值：
//   - 错误信息（如果检查或创建失败）
func CheckAndCreateColumns(db *sql.DB, tableName string, columns []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	exists, err := TableExists(db, tableName)
	if !exists && err == nil {
		// 表不存在，尝试创建
		err = CreateTable(db, tableName, columns)
		if err != nil {
			tx.Rollback()
			return err
		}
	} else if err != nil { // 出错，回滚
		tx.Rollback()
		return err
	}

	noColumns, err := ColumnsExist(db, tableName, columns)
	if err != nil {
		tx.Rollback() // 出错，回滚
		return err
	}
	if len(noColumns) > 0 {
		// 有列不存在，尝试添加
		err := AddColumns(db, tableName, noColumns)
		if err != nil {
			tx.Rollback() // 添加列失败，回滚
			return err
		}
	}
	err = tx.Commit() // 成功，提交事务
	if err != nil {
		tx.Rollback() // 提交失败，回滚
		return err
	}
	return nil
}
