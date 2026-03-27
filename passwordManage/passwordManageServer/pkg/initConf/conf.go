/*
 * @Author: 小鱼
 * @Date: 2024-10-18 15:52:16
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-09-29 11:15:00
 * @FilePath: \passwordManageServer\pkg\initConf\conf.go
 * @Description: 配置文件加载和管理，负责读取系统配置并初始化相关组件
 *
 * Copyright (c) 2024 by 小鱼, All Rights Reserved.
 */
package config

import (
	"net/http"
	"path/filepath"
	"xyrTools/passwordManage/passwordManageServer/pkg/sessionManage"

	"github.com/BurntSushi/toml"
)

type ServerConfig struct {
	Host         string `toml:"host"`
	Port         int    `toml:"port"`
	ReadTimeout  int    `toml:"read_timeout"`  // in seconds
	WriteTimeout int    `toml:"write_timeout"` // in seconds
}

type CorsConfig struct {
	AllowedOrigins []string `toml:"allowed_origins"`
	AllowedMethods []string `toml:"allowed_methods"`
	AllowedHeaders []string `toml:"allowed_headers"`
}

type DatabaseConfig struct {
	Type   string `toml:"type"` // "sqlite" or "mysql"
	SQLite SQLiteConfig
	MySQL  MySQLConfig
}

type SQLiteConfig struct {
	SqliteDbPath string `toml:"sqliteDbPath"`
}

type MySQLConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DBName   string `toml:"dbname"`
}

type LoggingConfig struct {
	Level   string `toml:"level"`
	LogFile string `toml:"log_file"`
}

type SecurityConfig struct {
	EncryptionMethod string `toml:"encryption_method"`
	KeyFilePath      string `toml:"key_file_path"`
}

type SSLConfig struct {
	CertificateFile     string `toml:"certificate_file"`
	PrivateKeyFile      string `toml:"private_key_file"`
	CACertificateFile   string `toml:"ca_certificate_file"`
	CertificatePassword string `toml:"certificate_password"`
	EnableSSL           bool   `toml:"enable_ssl"`
}

type UserKeyConfig struct {
	AesKeyBase64 string `toml:"aesKeyBase64"`
}

// EmailConfig 邮件配置结构体
type EmailConfig struct {
	SmtpServer   string `toml:"smtp_server"`
	SmtpPort     int    `toml:"smtp_port"`
	SmtpUser     string `toml:"smtp_user"`
	SmtpPassword string `toml:"smtp_password"`
	FromEmail    string `toml:"from_email"`
	ToEmail      string `toml:"to_email"`
}

type Config struct {
	Server        ServerConfig
	Cors          CorsConfig
	Database      DatabaseConfig
	Logging       LoggingConfig
	Security      SecurityConfig
	SSL           SSLConfig
	Session       sessionManage.SessionConfig
	UserKeyBase64 UserKeyConfig `toml:"userKey"`
}

// LoadConfig 读取配置文件并返回配置结构体
func LoadConfig(cfgPath string) (Config, error) {
	// 统一绝对路径
	configPath, err := filepath.Abs(cfgPath)
	if err != nil {
		return Config{}, err
	}
	var config Config
	_, err = toml.DecodeFile(configPath, &config)
	if err != nil {
		return config, err
	}

	// 解析 SameSite 字段
	switch config.Session.SameSiteMode {
	case "None":
		config.Session.SameSite = http.SameSiteNoneMode
	case "Lax":
		config.Session.SameSite = http.SameSiteLaxMode
	case "Strict":
		config.Session.SameSite = http.SameSiteStrictMode
	default:
		config.Session.SameSite = http.SameSiteDefaultMode
	}

	// 根据数据库类型加载数据库配置
	// switch config.Database.Type {
	// case "sqlite":
	// 	// SQLite 配置加载
	// 	if config.Database.SQLite.ConnectionString == "" {
	// 		return config, errors.New("missing SQLite connection string")
	// 	}
	// case "mysql":
	// 	// MySQL 配置加载
	// 	if config.Database.MySQL.Host == "" || config.Database.MySQL.User == "" || config.Database.MySQL.Password == "" || config.Database.MySQL.DBName == "" {
	// 		return config, errors.New("missing MySQL configuration")
	// 	}
	// default:
	// 	return config, errors.New("unsupported database type")
	// }

	// 返回解析的配置
	return config, nil
}
