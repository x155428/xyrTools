/*
 * @Author: 小鱼
 * @Date: 2025-09-29 10:44:29
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-10-22 14:16:44
 * @FilePath: \passwordManageServer\pkg\mail\sendMail.go
 * @Description: 邮件发送模块，负责发送系统邮件通知、验证码等功能
 *
 * Copyright (c) 2025 by 小鱼, All Rights Reserved.
 */
package mail

import (
	"crypto/tls"
	"fmt"
	"strings"

	"gopkg.in/gomail.v2"
)

type Mail struct {
	MailConfig GlobalConfig
}

//	{
//	    From:     "your_email@example.com",
//	    Nickname: "Go Mailer",
//	    Secret:   "your_smtp_secret", // QQ/163/Gmail 需要使用授权码
//	    Host:     "smtp.example.com",
//	    Port:     465,
//	    IsSSL:    true,
//	    To:       "receiver@example.com",
//	}
type GlobalConfig struct {
	From     string `json:"from"`
	Nickname string `json:"nickname"`
	Secret   string `json:"secret"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	IsSSL    bool   `json:"isSSL"`
	To       string `json:"to"`
}

type MailAlertConfig struct {
	GlobalConfig
	EncryptSecret bool `json:"encryptSecret"`
	Enabled       bool `json:"enabled"`
}

func (mail *Mail) send(to []string, subject, body string) error {
	cfg := mail.MailConfig
	if cfg.From == "" || cfg.Secret == "" || cfg.Host == "" {
		return fmt.Errorf("邮件配置不完整（From/Secret/Host 不能为空）")
	}

	// 如果没有设置端口，根据 SSL 配置自动设置端口
	if cfg.Port == 0 {
		if cfg.IsSSL {
			cfg.Port = 465 // 默认 SSL 端口
		} else {
			cfg.Port = 587 // 默认非 SSL 端口
		}
	}

	m := gomail.NewMessage()
	if cfg.Nickname != "" {
		m.SetHeader("From", fmt.Sprintf("%s <%s>", cfg.Nickname, cfg.From))
	} else {
		m.SetHeader("From", cfg.From)
	}
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.From, cfg.Secret)
	switch cfg.IsSSL {
	case true:
		d.SSL = true
		d.TLSConfig = &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         cfg.Host,
		}
	case false:
		d.SSL = false
		d.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         cfg.Host,
		}
	}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("邮件发送失败: %w (服务器=%s, 用户=%s)", err, cfg.Host, cfg.From)
	}
	return nil
}

// 封装，外部调用邮件发送方法
func (mail *Mail) Email(To, subject, body string) error {
	to := strings.Split(To, ",")
	return mail.send(to, subject, body)
}

func (mail *Mail) SetGlobalConfig(config GlobalConfig) {
	mail.MailConfig = config
}

func (mail *Mail) ClearGlobalConfig() {
	mail.MailConfig = GlobalConfig{}
}
