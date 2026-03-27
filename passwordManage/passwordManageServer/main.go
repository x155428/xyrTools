/**
 * @Author: 小鱼
 * @Date: 2024-04-18 13:22:35
 * @LastEditors: 小鱼
 * @LastEditTime: 2024-04-18 13:22:35
 * @FilePath: \passwordManageServer\main.go
 * @Description: 主程序入口，负责初始化配置、路由注册和启动服务器
 *
 * Copyright (c) 2024 by 小鱼, All Rights Reserved.
 */

package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	proxyproto "github.com/pires/go-proxyproto"
	"github.com/rs/cors"

	"xyrTools/passwordManage/passwordManageServer/pkg/db"
	config "xyrTools/passwordManage/passwordManageServer/pkg/initConf"
	"xyrTools/passwordManage/passwordManageServer/pkg/mail"
	"xyrTools/passwordManage/passwordManageServer/pkg/middleware"
	"xyrTools/passwordManage/passwordManageServer/pkg/otherFunc"
	"xyrTools/passwordManage/passwordManageServer/pkg/sessionManage"
)

var (
	address          string                                              // 监听地址
	tmpSessionAESKey []byte                                              // 临时会话密钥
	userAesKeyHex    string                                              // 用户AES密钥（16进制）
	configPath       string                      = "./conf/sysConf.toml" // 配置文件路径
	sessionStore     *sessionManage.SessionStore                         // 会话存储对象
	cfg              config.Config                                       // 配置对象
	mu               sync.RWMutex                = sync.RWMutex{}        // 线程锁
	globalMail       mail.Mail                                           // 全局邮件对象初始化
	whitelistCfg     WhitelistConfig                                     // 全局白名单配置
	rateLimiter      *middleware.RateLimiter                             // 全局请求限流对象
)

type loginStatus struct {
	isPwdAuthed bool          // 密码是否已认证
	StopClearCh chan struct{} // 停止清除登录状态通道
}

var (
	loginStatusList map[string]*loginStatus = make(map[string]*loginStatus) // 登录状态记录
	statusMux       sync.Mutex
)

// 前端少量运行时配置，仅运行时有效，服务端重启失效
type tmpCfg struct {
	SystemName    string `json:"systemName"`    // 系统名称
	Timeout       int    `json:"timeout"`       // 会话超时时间（秒）
	WeatherApiKey string `json:"weatherApiKey"` // 天气API密钥
}

var tmpRunCfg = tmpCfg{
	SystemName:    "密码管理系统",
	WeatherApiKey: "",
}

func main() {
	// 命令行参数解析
	configFilePath := flag.String("config", "", "Path to the configuration file")
	ip := flag.String("ip", "", "IP address to listen on")
	port := flag.String("port", "", "Port to listen on")
	uuid := flag.String("uuid", "", "UUID to listen on")
	name := flag.String("name", "", "Name to listen on")
	heartbeatAddr := flag.String("heartbeatAddress", "", "HeartbeatAddr to listen on")
	isHeartBeatStr := flag.String("isHeartbeat", "", "IsHeartBeat to listen on")
	flag.Parse()

	fmt.Println(time.Now())

	// 心跳检测
	isHeartBeat, err := strconv.ParseBool(*isHeartBeatStr)
	if err != nil {
		fmt.Println("[-]心跳检测配置异常，关闭心跳检测！")
		isHeartBeat = false
	}
	if isHeartBeat {
		fmt.Print("[+]开启心跳检测！")
		go otherFunc.HeartBeat(*heartbeatAddr, *uuid, *name)
	}

	// 加载配置文件
	if *configFilePath != "" {
		configPath = *configFilePath
	}

	var loadConfErr error
	cfg, loadConfErr = config.LoadConfig(configPath)
	if loadConfErr != nil {
		log.Fatalf("[-]加载配置异常: %v\n", loadConfErr)
	}
	tmpRunCfg.Timeout = cfg.Session.MaxAge

	// 从 base64 解出 user key
	if cfg.UserKeyBase64.AesKeyBase64 != "" {
		userAesKeyBytes, _ := base64.StdEncoding.DecodeString(cfg.UserKeyBase64.AesKeyBase64)
		userAesKeyHex = hex.EncodeToString(userAesKeyBytes)
	} else {
		userAesKeyHex = ""
	}

	// 命令行覆盖 ip/port
	if *ip != "" {
		cfg.Server.Host = *ip
	}
	if *port != "" {
		portInt, err := strconv.Atoi(*port)
		if err != nil {
			log.Fatalf("[-]端口配置异常: %v\n", err)
		}
		cfg.Server.Port = portInt
	}
	address = fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	// 数据库初始化
	db.InitDB(cfg.Database.SQLite.SqliteDbPath, "./db/initDb.sql")

	// CORS跨域配置
	sysCors := cors.New(cors.Options{
		AllowedOrigins:   cfg.Cors.AllowedOrigins,
		AllowedMethods:   cfg.Cors.AllowedMethods,
		AllowedHeaders:   cfg.Cors.AllowedHeaders,
		AllowCredentials: true,
		ExposedHeaders:   []string{"Link"},
	})

	// Session 初始化
	sessionConf := sessionManage.SessionConfig{
		Domain:       cfg.Session.Domain,
		Path:         cfg.Session.Path,
		MaxAge:       cfg.Session.MaxAge,
		Secure:       cfg.Session.Secure,
		HttpOnly:     cfg.Session.HttpOnly,
		SameSiteMode: cfg.Session.SameSiteMode,
	}
	// 创建会话存储实例
	var errSession error
	sessionStore, errSession = sessionManage.NewSessionStore(cfg.Database.SQLite.SqliteDbPath, sessionConf)
	if errSession != nil {
		fmt.Printf("[-]创建session实例出错: %s\n", errSession)
		return
	}
	defer sessionStore.Close()
	// 定期清理过期会话
	go sessionStore.CleanupExpiredSessions(cfg.Session)

	rateLimiter = middleware.NewRateLimiter(2, 10, time.Minute)

	// 路由注册
	mux := http.NewServeMux()
	// 全局提取ip拦截器、请求限流拦截器
	filterGlobal := middleware.NewRouterManager(mux, middleware.IPInterceptor, middleware.ReqRateIpLimitInterceptor(rateLimiter, 4, 20))
	// 附加认证拦截器
	authFilter := filterGlobal.NewGroup(middleware.AuthInterceptor(sessionStore))
	// 公开路由
	filterGlobal.Handle("/stats", stats, middleware.ReqRateApiLimitInterceptor(rateLimiter, 0.5, 5))
	filterGlobal.Handle("/getAesKey", getAesKey)
	filterGlobal.Handle("/register", register, middleware.ReqRateApiLimitInterceptor(rateLimiter, 0.5, 5))
	filterGlobal.Handle("/login", login, middleware.ReqRateApiLimitInterceptor(rateLimiter, 0.01, 5))
	filterGlobal.Handle("/loginAlert", loginAlert, middleware.ReqRateApiLimitInterceptor(rateLimiter, 0.5, 5))
	filterGlobal.Handle("/verifyTotp", verifyTotp, middleware.ReqRateApiLimitInterceptor(rateLimiter, 0.5, 5))

	// 需登录认证的路由
	authFilter.Handle("/logout", logout)
	authFilter.Handle("/saveSecret", saveSecret)
	authFilter.Handle("/setting", setting)
	authFilter.Handle("/checkSession", checkSession)
	authFilter.Handle("/queryData", queryData)
	authFilter.Handle("/saveKeyToCS", saveKeyToCS)
	authFilter.Handle("/deleteRecords", deleteRecords, middleware.ReqRateApiLimitInterceptor(rateLimiter, 0.5, 5))
	authFilter.Handle("/clearKey", clearKey)
	authFilter.Handle("/updateStats", updateStats)
	authFilter.Handle("/queryWinApi", queryWinApi)
	authFilter.Handle("/insertWinApi", insertWinApi)
	authFilter.Handle("/deleteWinApi", deleteWinApi, middleware.ReqRateApiLimitInterceptor(rateLimiter, 0.5, 5))
	authFilter.Handle("/downloadKeyFile", downloadKeyFile, middleware.ReqRateApiLimitInterceptor(rateLimiter, 0.5, 5))
	authFilter.Handle("/updatePassword", updatePassword)
	authFilter.Handle("/getTmpCfg", getTmpCfg)
	authFilter.Handle("/changeMasterKey", changeMasterKey)
	authFilter.Handle("/getAvatarAndUser", getAvatarAndUser)
	authFilter.Handle("/updateAvatar", updateAvatar)
	authFilter.Handle("/setTotp", setTotp, middleware.ReqRateApiLimitInterceptor(rateLimiter, 0.5, 5))
	authFilter.Handle("/setLoginAlertSettings", setLoginAlertSettings, middleware.ReqRateApiLimitInterceptor(rateLimiter, 0.5, 5))
	authFilter.Handle("/getSecuritySettings", getSecuritySettings)
	authFilter.Handle("/enableTotp", enableTotp)
	authFilter.Handle("/setWhitelistSettings", setWhitelistSettings, middleware.ReqRateApiLimitInterceptor(rateLimiter, 0.5, 5))
	authFilter.Handle("/importByFile", importByFile, middleware.ReqRateApiLimitInterceptor(rateLimiter, 0.5, 5))
	authFilter.Handle("/exportAll", exportAll)

	handler := sysCors.Handler(mux)

	// 证书加载
	certPath, err := filepath.Abs(cfg.SSL.CertificateFile)
	if err != nil {
		log.Fatalf("[-]证书路径错误: %v", err)
	}
	keyPath, err := filepath.Abs(cfg.SSL.PrivateKeyFile)
	if err != nil {
		log.Fatalf("[-]私钥路径错误: %v", err)
	}
	tlsCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Fatalf("[-]加载证书失败: %v", err)
	}

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	// 网络net监听tcp端口
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("[-]监听失败: %v", err)
	}

	// 包装监听支持 PROXY protocol并设置 header 读取超时（解析proxy protocol header）
	proxyListener := &proxyproto.Listener{
		Listener:          ln,
		ReadHeaderTimeout: 3 * time.Second,
	}
	defer proxyListener.Close()

	tlsListener := tls.NewListener(proxyListener, tlsConf)

	server := &http.Server{
		Addr:         address,
		Handler:      handler,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)), // 禁用 HTTP/2（保持你原来的调试策略）
	}

	fmt.Printf("[+]HTTPS 服务器启动：%s\n", address)
	if err := server.Serve(tlsListener); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[-]服务器启动失败: %v", err)
	}
}
