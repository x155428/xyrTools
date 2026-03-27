/**
 * @Author: 小鱼
 * @Date: 2024-04-18 13:22:35
 * @LastEditors: 小鱼
 * @LastEditTime: 2024-04-18 13:22:35
 * @FilePath: \passwordManageServer\pkg\sessionManage\sessionManage.go
 * @Description: 会话管理模块，用于提供用户会话的创建、验证和管理功能
 *
 * Copyright (c) 2024 by 小鱼, All Rights Reserved.
 */
package sessionManage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

// 会话配置
type SessionConfig struct {
	Domain                     string        `toml:"domain"`
	Path                       string        `toml:"path"`
	MaxAge                     int           `toml:"max_age"`
	Secure                     bool          `toml:"secure"`
	HttpOnly                   bool          `toml:"http_only"`
	SameSite                   http.SameSite `toml:"-"`
	SameSiteMode               string        `toml:"same_site"`
	CleanupExpiredSessionsTime int           `toml:"cleanup_expired_sessions_time"` // 清理过期会话的时间间隔（秒）
}

type SessionStore struct {
	config            SessionConfig
	db                *sql.DB
	stmtInsert        *sql.Stmt
	stmtUpdate        *sql.Stmt
	stmtUpdateExpires *sql.Stmt
	stmtSelect        *sql.Stmt
	stmtDelete        *sql.Stmt
}

// NewSessionStore 创建一个数据库会话存储，并初始化数据库
// 参数：
// - dsn: 数据库连接字符串
// - config: 会话配置对象
// 返回值：
// - 会话存储对象指针
// - 错误信息
func NewSessionStore(dsn string, config SessionConfig) (*SessionStore, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	// 初始化数据库表
	err = initializeDatabase(db)
	if err != nil {
		return nil, err
	}

	store := &SessionStore{
		db:     db,
		config: config, // 存储配置
	}

	// 预编译 SQL 语句
	store.stmtInsert, err = db.Prepare("INSERT INTO sessions (id, data, expires_at) VALUES (?, ?, ?)")
	if err != nil {
		return nil, err
	}

	store.stmtUpdate, err = db.Prepare("UPDATE sessions SET data=?, expires_at=? WHERE id=?")
	if err != nil {
		return nil, err
	}

	store.stmtUpdateExpires, err = db.Prepare("UPDATE sessions SET expires_at=? WHERE id=?")
	if err != nil {
		return nil, err
	}

	store.stmtSelect, err = db.Prepare("SELECT data, expires_at FROM sessions WHERE id=?")
	if err != nil {
		return nil, err
	}

	store.stmtDelete, err = db.Prepare("DELETE FROM sessions WHERE id=?")
	if err != nil {
		return nil, err
	}

	return store, nil
}

// initializeDatabase 初始化数据库表
// 参数：
// - db: 数据库连接对象
// 返回值：
// - 错误信息
func initializeDatabase(db *sql.DB) error {
	// 创建会话表
	query := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		data BLOB NOT NULL,
		expires_at DATETIME NOT NULL
	);
	`
	_, err := db.Exec(query)
	return err
}

// 定期清理过期会话
func (s *SessionStore) CleanupExpiredSessions(config SessionConfig) {
	for {
		time.Sleep(time.Duration(config.CleanupExpiredSessionsTime) * time.Second)
		// 清理过期会话时，使用当前UTC时间与存储的过期时间比较，确保时间基准一致
		currentTimeUTC := time.Now().UTC().Format(time.RFC3339)
		_, err := s.db.Exec("DELETE FROM sessions WHERE expires_at < ?", currentTimeUTC)
		if err != nil {
			log.Printf("Failed to clean up expired sessions: %v\n", err)
		} else {
			//log.Println("Expired sessions cleaned up successfully.")
		}
	}
}

// Get 获取请求会话
func (s *SessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	// 从请求中获取会话 ID
	sessionID, err := r.Cookie(name)
	if err != nil {
		return nil, err
	}
	// 根据sessionid获取session数据
	session := sessions.NewSession(s, name)
	session.ID = sessionID.Value
	err = s.load(session, sessionID.Value)
	if err != nil {
		// 会话失效
		return nil, err
	}

	return session, nil
}

// 新建全新会话，检查是否存在旧会话，存在则删除旧会话，新建新会话
func (s *SessionStore) New(r *http.Request, name string) (*sessions.Session, error) {
	// 创建一个新的 session 对象
	session := sessions.NewSession(s, name)
	session.Options = &sessions.Options{
		Path:     s.config.Path,
		Domain:   s.config.Domain,
		MaxAge:   s.config.MaxAge,
		Secure:   s.config.Secure,
		HttpOnly: s.config.HttpOnly,
		SameSite: s.config.SameSite,
	}
	session.IsNew = true
	session.ID = generateSessionID() // 生成新的 Session ID
	// //从cookie读取sessionid    有漏洞，恶意删除session，暂时不搞
	// cookie, err := r.Cookie(name)
	// if err != nil {
	// 	// 没有cookie，创建新会话
	// 	return session, nil
	// }
	// // 有cookie，从cookie读取sessionid
	// session.ID = cookie.Value
	// // 尝试从数据库中删除该sessionid
	// _, err = s.stmtDelete.Exec(session.ID)
	// if err != nil {
	// 	log.Printf("Failed to clean up expired sessions: %v\n", err)
	// }
	return session, nil
}

// Save 保存会话数据
// - 功能说明：将会话数据保存到数据库，并设置响应cookie
// - 参数：
//   - r: HTTP请求对象
//   - w: HTTP响应对象
//   - session: 会话对象
// - 返回值：
//   - 错误信息（如果保存失败）
//sessions.Store接口的Save方法实现
func (s *SessionStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	// 获取sessionid
	normalized := make(map[string]interface{})
	for key, value := range session.Values {
		strKey, ok := key.(string) // 转换键为 string
		if !ok {
			fmt.Printf("session数据解析失败: %v\n", key)
			// session无效
			return fmt.Errorf("session数据解析失败")
		}
		normalized[strKey] = value
	}
	data, err := json.Marshal(normalized)
	if err != nil {
		return err
	}

	// 更新过期时间（使用UTC时间，确保跨时区一致性）
	expiresAt := time.Now().UTC().Add(time.Duration(s.config.MaxAge) * time.Second).Format(time.RFC3339)

	if session.IsNew {
		// 插入新会话
		_, err = s.stmtInsert.Exec(session.ID, data, expiresAt)
		if err != nil {
			return err
		}

	} else {
		// 更新现有会话
		_, err = s.stmtUpdate.Exec(data, expiresAt, session.ID)
		if err != nil {
			return err
		}
	}
	// 设置 cookie，统一使用固定名称"sessionId"避免大小写问题
	cookie := &http.Cookie{
		Name:     "sessionId", // 统一使用固定名称
		Value:    session.ID,
		Path:     "/",               // 固定路径
		MaxAge:   s.config.MaxAge,   // 使用统一配置
		HttpOnly: true,              // 强制HttpOnly
		Secure:   s.config.Secure,   // 继承配置的安全设置
		SameSite: s.config.SameSite, // 继承SameSite配置
	}
	http.SetCookie(w, cookie)
	return nil
}

// load 从数据库加载会话
// - 功能说明：根据会话ID从数据库加载会话数据并检查是否过期
// - 参数：
//   - session: 会话对象，用于存储加载的数据
//   - id: 会话ID
// - 返回值：
//   - 错误信息（如果加载失败或会话不存在/已过期）
func (s *SessionStore) load(session *sessions.Session, id string) error {
	var data []byte
	var expiresAt time.Time

	// 通过session的ID查询数据库中对应的会话数据和过期时间
	err := s.stmtSelect.QueryRow(id).Scan(&data, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("会话不存在")
		}
		// fmt.Printf("查询会话数据失败: %v\n", err)
		return err
	}

	// 检查会话是否过期（确保与存储时使用相同的UTC时间基准）
	currentTime := time.Now().UTC()
	if expiresAt.Before(currentTime) {
		// 会话已过期，删除该会话
		_, _ = s.stmtDelete.Exec(id)
		return errors.New("会话已过期")
	}

	// 会话没过期，更新会话
	// 创建一个临时 map 来存储解码后的 JSON 数据
	var tempMap map[string]interface{}
	err = json.Unmarshal(data, &tempMap)
	if err != nil {
		return fmt.Errorf("反序列化原session数据出错: %w", err)
	}

	// 将临时 map 的内容复制到 session.Values 中
	for key, value := range tempMap {
		session.Values[key] = value
	}

	return nil
}

// Delete 删除会话
// - 功能说明：从数据库中删除会话数据，并设置cookie过期
// - 参数：
//   - r: HTTP请求对象
//   - w: HTTP响应对象
//   - session: 要删除的会话对象
// - 返回值：
//   - 错误信息（如果删除失败）
func (s *SessionStore) Delete(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	_, err := s.stmtDelete.Exec(session.ID)
	if err != nil {
		return err
	}
	// 统一使用固定名称"sessionId"设置过期cookie
	expireCookie := &http.Cookie{
		Name:     "sessionId",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Expires:  time.Unix(0, 0), // 设置为过去时间立即过期
	}
	http.SetCookie(w, expireCookie)
	return nil
}

// DeleteSession 根据会话ID删除会话
// - 功能说明：仅从数据库中删除指定ID的会话数据
// - 参数：
//   - sessionID: 要删除的会话ID
// - 返回值：
//   - 错误信息（如果删除失败）
func (s *SessionStore) DeleteSession(sessionID string) error {
	_, err := s.stmtDelete.Exec(sessionID)
	if err != nil {
		return err
	}
	return nil
}

// CheckSessionValid 检查会话是否有效（外部调用）
// - 功能说明：从请求cookie中获取会话ID并检查其有效性
// - 参数：
//   - r: HTTP请求对象
//   - w: HTTP响应对象
// - 返回值：
//   - 布尔值，表示会话是否有效
//   - 错误信息（如果获取cookie失败）
func (s *SessionStore) CheckSessionValid(r *http.Request, w http.ResponseWriter) (bool, error) {
	// 从请求cookie中获取sessionId（注意大小写统一）
	cookie, err := r.Cookie("sessionId")
	if err != nil {
		return false, err
	}
	sessionID := cookie.Value

	return s.checkSession(sessionID)
}

// checkSession 检查会话是否有效
// - 功能说明：根据会话ID检查会话是否存在且未过期
// - 参数：
//   - sessionID: 要检查的会话ID
// - 返回值：
//   - 布尔值，表示会话是否有效
//   - 错误信息（如果数据库查询失败）
func (s *SessionStore) checkSession(sessionID string) (bool, error) {
	var expiresAt time.Time
	var data []byte
	err := s.stmtSelect.QueryRow(sessionID).Scan(&data, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // 会话不存在
		}
		return false, err // 数据库查询错误
	}
	return expiresAt.After(time.Now().UTC()), nil // 会话有效，使用UTC时间比较确保一致性
}

// Close 关闭数据库连接
// - 功能说明：关闭会话存储使用的数据库连接
// - 返回值：
//   - 错误信息（如果关闭连接失败）
func (s *SessionStore) Close() error {
	return s.db.Close()
}

// generateSessionID 生成一个唯一的会话 ID
// - 功能说明：生成一个基于UUID的唯一会话标识符
// - 返回值：
//   - 唯一的会话ID字符串
func generateSessionID() string {
	//return fmt.Sprintf("%d", time.Now().UnixNano())
	return uuid.New().String()
}

// UpdateSession 更新会话
// - 功能说明：检查会话有效性并更新其过期时间
// - 参数：
//   - r: HTTP请求对象
//   - w: HTTP响应对象
//   - name: 会话名称
// - 返回值：
//   - 布尔值，表示更新是否成功
//   - 错误信息（如果更新失败）
func (s *SessionStore) UpdateSession(r *http.Request, w http.ResponseWriter, name string) (bool, error) {
	//从请求中获取sessionid
	cookie, err := r.Cookie(name)
	if err != nil {
		return false, err
	}
	sessionID := cookie.Value
	// 检查session是否有效
	valid, err := s.checkSession(sessionID)
	if err != nil || !valid {
		return false, err
	}
	// 获取session
	session, err := s.Get(r, name)
	if err != nil {
		return false, err
	}
	// 更新session
	err = s.Save(r, w, session)
	if err != nil {
		return false, err
	}
	return true, nil

}

// SetMaxAge 更新会话最大存活时间
// - 功能说明：设置会话的最大存活时间（秒）
// - 参数：
//   - maxAge: 会话最大存活时间（秒）
func (s *SessionStore) SetMaxAge(maxAge int) {
	s.config.MaxAge = maxAge
}

// GetSessionInfo 获取会话信息
// - 功能说明：从请求cookie中获取会话ID并返回会话中存储的所有信息
// - 参数：
//   - r: HTTP请求对象
// - 返回值：
//   - 包含会话信息的map
//   - 错误信息（如果获取失败）
func (s *SessionStore) GetSessionInfo(r *http.Request) (map[string]interface{}, error) {
	var data []byte
	var expiresAt time.Time
	// 从请求cookie中获取sessionId（注意大小写统一）
	cookie, err := r.Cookie("sessionId")
	if err != nil {
		return nil, err
	}
	sessionID := cookie.Value
	err = s.stmtSelect.QueryRow(sessionID).Scan(&data, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 会话不存在
		}
		return nil, err // 数据库查询错误
	}
	// 会话没过期，更新会话
	// 创建一个临时 map 存储解码后的 JSON 数据
	var tempMap map[string]interface{}
	err = json.Unmarshal(data, &tempMap)
	if err != nil {
		return nil, fmt.Errorf("反序列化session数据出错: %w", err)
	}
	return tempMap, nil
}
