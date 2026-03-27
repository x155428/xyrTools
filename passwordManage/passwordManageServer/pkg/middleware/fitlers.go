/*
 * @Author: 小鱼
 * @Date: 2025-09-26 16:05:48
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-10-21 16:49:15
 * @FilePath: \passwordManageServer\pkg\middleware\fitlers.go
 * @Description:
 *
 * Copyright (c) 2025 by 小鱼, All Rights Reserved.
 */
package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
	respMessage "xyrTools/passwordManage/passwordManageServer/pkg/responseStd"
	"xyrTools/passwordManage/passwordManageServer/pkg/sessionManage"

	"golang.org/x/time/rate"
)

// AuthMiddleware 认证拦截器，用于验证请求是否是登录认证的合法请求
// 参数：
// - next: 下一个处理函数
// - sessionStore: 会话存储对象
// 返回：
// - 包装后的处理函数
func AuthMiddleware(next http.HandlerFunc, sessionStore *sessionManage.SessionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		sessionInfo, err := sessionStore.GetSessionInfo(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			response := respMessage.CommonResponse{
				Code:    http.StatusUnauthorized,
				Message: "会话异常！",
				Data:    nil,
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// 检查session是否有效
		valid, checkErr := sessionStore.UpdateSession(r, w, "sessionId")
		if !valid || checkErr != nil {
			// session失效，返回未授权错误
			w.WriteHeader(http.StatusUnauthorized)
			response := respMessage.CommonResponse{
				Code:    http.StatusUnauthorized,
				Message: "会话过期！",
				Data:    nil,
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		// 提取IP头信息，获取客户端IP地址
		clientIPs := extractIPHeaders(r)
		fmt.Printf("[拦截器] IP 信息: %+v\n", clientIPs)

		// 将用户名存储在请求上下文中
		ctx := r.Context()
		ctx = context.WithValue(ctx, "username", sessionInfo["username"])
		ctx = context.WithValue(ctx, "clientIPs", clientIPs)
		r = r.WithContext(ctx)

		// 认证通过，调用下一个处理函数
		next(w, r)
	}
}

// GetCurrentUsername 从会话中获取当前用户名
// 参数：
// - r: HTTP请求对象
// - sessionStore: 会话存储对象
// 返回：
// - 用户名和错误信息
func GetCurrentUsername(r *http.Request, sessionStore *sessionManage.SessionStore) (string, error) {
	userSession, err := sessionStore.Get(r, "sessionId")
	if err != nil {
		return "", err
	}

	username, ok := userSession.Values["username"].(string)
	if !ok || username == "" {
		return "", http.ErrNoCookie
	}

	return username, nil
}

type IPInfo struct {
	ClientIP   string            `json:"clientIP"`
	RemoteAddr string            `json:"remoteAddr"`
	Headers    map[string]string `json:"headers"`
	Chain      []string          `json:"chain"`
}

// ********************************************************************
// ip提取拦截器
// extractIPHeaders 从请求头中获取IP相关头信息
// - 功能说明：提取请求中的各种IP相关头信息，用于获取客户端真实IP
// - 参数：
//   - r: HTTP请求对象
//
// - 返回值：
//   - 包含各种IP头信息的映射表（键为头名称，值为头内容）
func extractIPHeaders(r *http.Request) IPInfo {
	headers := []string{
		"CF-Connecting-IP",
		"True-Client-IP",
		"X-Real-IP",
		"X-Forwarded-For",
		"Forwarded",
		"Via",
	}

	ipInfo := IPInfo{
		Headers: make(map[string]string),
	}

	var candidateIPs []string

	for _, h := range headers {
		if v := strings.TrimSpace(r.Header.Get(h)); v != "" {
			ipInfo.Headers[h] = v

			switch h {
			case "X-Forwarded-For":
				for _, part := range strings.Split(v, ",") {
					ip := strings.TrimSpace(part)
					if ip != "" && !isPrivateIP(ip) {
						candidateIPs = append(candidateIPs, ip)
						break
					}
				}
			case "Forwarded":
				re := regexp.MustCompile(`for="?([^";]+)"?`)
				matches := re.FindStringSubmatch(v)
				if len(matches) > 1 {
					ip := strings.TrimSpace(matches[1])
					if ip != "" && !isPrivateIP(ip) {
						candidateIPs = append(candidateIPs, ip)
					}
				}
			default:
				if !isPrivateIP(v) {
					candidateIPs = append(candidateIPs, v)
				}
			}
		}
	}

	// RemoteAddr 兜底
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ipInfo.RemoteAddr = host

	// 选出最终ClientIP
	ipInfo.ClientIP = firstNonEmpty(candidateIPs...)
	if ipInfo.ClientIP == "" {
		ipInfo.ClientIP = host
	}
	ipInfo.Chain = candidateIPs

	return ipInfo
}

func IPInterceptor(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ipInfo := extractIPHeaders(r)
		ctx := context.WithValue(r.Context(), "ipInfo", ipInfo)
		next(w, r.WithContext(ctx))
	}
}

// 判断是否是私有或保留IP
func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return true
	}
	privateCIDRs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8", // loopback
		"::1/128",     // IPv6 loopback
		"fc00::/7",    // unique local address
		"fe80::/10",   // link-local
	}
	for _, cidr := range privateCIDRs {
		_, block, _ := net.ParseCIDR(cidr)
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// 返回第一个非空字符串
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// ********************************************************************
// 会话验证拦截器
type SessionStore interface {
	GetSessionInfo(r *http.Request) (map[string]string, error)
	UpdateSession(r *http.Request, w http.ResponseWriter, sessionKey string) (bool, error)
}

// AuthInterceptor 会话验证拦截器
// - 功能说明：验证HTTP请求中的会话信息，确保用户已登录
// - 参数：
//   - sessionStore: 会话存储对象
//
// - 返回值：
//   - 新的会话验证拦截器实例
func AuthInterceptor(sessionStore *sessionManage.SessionStore) Interceptor {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 获取会话信息
			sessionInfo, err := sessionStore.GetSessionInfo(r)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				response := respMessage.CommonResponse{
					Code:    http.StatusUnauthorized,
					Message: "会话异常！",
					Data:    nil,
				}
				json.NewEncoder(w).Encode(response)
				return
			}

			// 检查 session 是否有效并刷新
			valid, checkErr := sessionStore.UpdateSession(r, w, "sessionId")
			if !valid || checkErr != nil {
				respMessage.SendErrorResponse(w, http.StatusUnauthorized, "会话过期！")
				return
			}

			// 将用户名存入请求上下文
			ctx := context.WithValue(r.Context(), "username", sessionInfo["username"])
			r = r.WithContext(ctx)

			// 调用下一个处理函数
			next(w, r)
		}
	}
}

// ********************************************************************请求限速拦截器***************

// 单个限速器定义
type Limit struct {
	Limiter    *rate.Limiter // 每个IP独立限流器
	LastReq    time.Time     // 最近请求时间
	RateLimit  rate.Limit    // 每秒生成令牌数
	BucketSize int           // 桶容量
}

// 全局限速管理对象
type RateLimiter struct {
	mu            sync.RWMutex
	ipLimitStore  map[string]*Limit // 存储每个IP限流器
	clearTime     time.Duration     // 清理频率
	apiLimitStore map[string]*Limit // 存储每个API限流器
}

// NewRateLimiter 创建一个新的限流器
// - 参数：
//   - r: 每秒生成令牌数
//   - b: 桶容量
//
// - 返回值：
//   - 新的限流器实例
func NewRateLimiter(r rate.Limit, b int, clearTime time.Duration) *RateLimiter {
	rl := &RateLimiter{
		ipLimitStore:  make(map[string]*Limit),
		apiLimitStore: make(map[string]*Limit),
		clearTime:     clearTime, // 初始化清理时间
	}
	go rl.cleanup()
	return rl
}

// getLimiter 获取或创建指定IP的限流器
// - 参数：
//   - ip: 客户端IP地址
//
// - 返回值：
//   - 对应IP的限流器实例
func (rl *RateLimiter) getIPLimiter(ip string, rateLimit rate.Limit, bucketSize int) *rate.Limiter {
	now := time.Now()

	// === 查找 存在直接返回 ===
	rl.mu.RLock()
	if entry, ok := rl.ipLimitStore[ip]; ok {
		entry.LastReq = now
		rl.mu.RUnlock()
		return entry.Limiter
	}
	rl.mu.RUnlock()

	// === 新记录需创建 ===
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 双重检查，防止竞争期间其他协程已创建
	if entry, ok := rl.ipLimitStore[ip]; ok {
		entry.LastReq = now
		return entry.Limiter
	}
	// 新建的限流器配置
	limiter := rate.NewLimiter(rateLimit, bucketSize)
	rl.ipLimitStore[ip] = &Limit{
		Limiter:    limiter,
		LastReq:    now,
		RateLimit:  rateLimit,
		BucketSize: bucketSize,
	}
	return limiter
}

// Allow 判断是否允许新请求
// - 参数：
//   - ip: 客户端IP地址
//
// - 返回值：
//   - 是否允许新请求
func (rl *RateLimiter) AllowIP(ip string, rateLimit rate.Limit, bucketSize int) bool {
	limiter := rl.getIPLimiter(ip, rateLimit, bucketSize)
	return limiter.Allow()
}

// getAPILimiter 获取或创建指定API的限流器
// - 参数：
//   - api: API路径
//
// - 返回值：
//   - 对应API的限流器实例
func (rl *RateLimiter) getAPILimiter(api string, rateLimit rate.Limit, bucketSize int) *rate.Limiter {
	now := time.Now()

	// === 查找 存在直接返回 ===
	rl.mu.RLock()
	if entry, ok := rl.apiLimitStore[api]; ok {
		entry.LastReq = now
		rl.mu.RUnlock()
		return entry.Limiter
	}
	rl.mu.RUnlock()

	// === 新记录需创建 ===
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 双重检查，防止竞争期间其他协程已创建
	if entry, ok := rl.apiLimitStore[api]; ok {
		entry.LastReq = now
		return entry.Limiter
	}

	// 新建的限流器配置
	limiter := rate.NewLimiter(rateLimit, bucketSize)
	rl.apiLimitStore[api] = &Limit{
		Limiter:    limiter,
		LastReq:    now,
		RateLimit:  rateLimit,
		BucketSize: bucketSize,
	}
	return limiter
}

// Allow 判断是否允许新请求
// - 参数：
//   - api: API路径
//
// - 返回值：
//   - 是否允许新请求
func (rl *RateLimiter) AllowAPI(api string, rateLimit rate.Limit, bucketSize int) bool {
	limiter := rl.getAPILimiter(api, rateLimit, bucketSize)
	return limiter.Allow()
}

// cleanup 清理过期限流器
// - 功能说明：定期清理未活动超过clearTime的限流器记录
func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(rl.clearTime)
		now := time.Now()

		rl.mu.Lock()
		for ip, entry := range rl.ipLimitStore {
			if now.Sub(entry.LastReq) > rl.clearTime {
				delete(rl.ipLimitStore, ip)
			}
		}
		for api, entry := range rl.apiLimitStore {
			if now.Sub(entry.LastReq) > rl.clearTime {
				delete(rl.apiLimitStore, api)
			}
		}
		rl.mu.Unlock()
	}
}

// ReqRateIpLimitInterceptor IP请求限速拦截器
// - 参数：
//   - rl: 限流器实例
//   - rateLimit: 每秒生成令牌数
//   - bucketSize: 桶容量
//
// - 返回值：
//   - 限速拦截器函数
func ReqRateIpLimitInterceptor(rl *RateLimiter, rateLimit rate.Limit, bucketSize int) Interceptor {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 从context获取ip
			ipInfo := r.Context().Value("ipInfo").(IPInfo)
			if ipInfo.ClientIP == "" {
				respMessage.SendErrorResponse(w, http.StatusInternalServerError, "获取IP失败！")
				return
			}

			if !rl.AllowIP(ipInfo.ClientIP, rateLimit, bucketSize) {
				respMessage.SendErrorResponse(w, http.StatusTooManyRequests, "请求频率过快！")
				return
			}
			next(w, r)
		}
	}
}

// ReqRateApiLimitInterceptor API请求限速拦截器，用于对某个api定制限速，例如login登录，防止高并发
// - 参数：
//   - rl: 限流器实例
//   - rateLimit: 每秒生成令牌数
//   - bucketSize: 桶容量
//
// - 返回值：
//   - 限速拦截器函数
func ReqRateApiLimitInterceptor(rl *RateLimiter, rateLimit rate.Limit, bucketSize int) Interceptor {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 从context获取url
			url := r.URL.Path
			if url == "" {
				respMessage.SendErrorResponse(w, http.StatusInternalServerError, "获取URL失败！")
				return
			}

			if !rl.AllowAPI(url, rateLimit, bucketSize) {
				respMessage.SendErrorResponse(w, http.StatusTooManyRequests, "请求频率过快！")
				return
			}
			next(w, r)
		}
	}
}

func (rl *RateLimiter) Stats() map[string]int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return map[string]int{
		"ipCount":  len(rl.ipLimitStore),
		"apiCount": len(rl.apiLimitStore),
	}
}

// ***********************************************ip拦截白名单***************************
// - 功能说明：
//   - 根据白名单配置处理请求
//
// - 参数：
// - 返回值：
//   - 是否允许请求
func IsWhiteListInterceptor() Interceptor {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 从context获取ip
			ipInfo := r.Context().Value("ipInfo").(IPInfo)
			if ipInfo.ClientIP == "" {
				respMessage.SendErrorResponse(w, http.StatusInternalServerError, "获取IP失败！")
				return
			}

		}
	}

}
