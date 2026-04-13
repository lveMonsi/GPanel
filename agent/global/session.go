package global

import (
	"sync"
	"sync/atomic"
	"time"
)

// WsSSHSession 表示一个通过 WebSocket 建立的 SSH 会话
type WsSSHSession struct {
	ID        int64
	Username  string
	Host      string
	Port      int
	LoginTime string
	Close     func() // 关闭回调
}

var (
	wsSSHSessions sync.Map // map[int64]WsSSHSession
	sessionIDSeq  int64
)

// RegisterWsSSHSession 注册一个 WebSocket SSH 会话，返回会话 ID
func RegisterWsSSHSession(username, host string, port int, closeFn func()) int64 {
	id := atomic.AddInt64(&sessionIDSeq, 1)
	wsSSHSessions.Store(id, WsSSHSession{
		ID:        id,
		Username:  username,
		Host:      host,
		Port:      port,
		LoginTime: time.Now().Format("2006-01-02 15:04:05"),
		Close:     closeFn,
	})
	return id
}

// UnregisterWsSSHSession 注销一个 WebSocket SSH 会话
func UnregisterWsSSHSession(id int64) {
	wsSSHSessions.Delete(id)
}

// CloseWsSSHSession 关闭并注销一个 WebSocket SSH 会话
func CloseWsSSHSession(id int64) bool {
	val, ok := wsSSHSessions.LoadAndDelete(id)
	if !ok {
		return false
	}
	session := val.(WsSSHSession)
	if session.Close != nil {
		session.Close()
	}
	return true
}

// GetWsSSHSessions 获取所有活跃的 WebSocket SSH 会话
func GetWsSSHSessions() []WsSSHSession {
	var sessions []WsSSHSession
	wsSSHSessions.Range(func(key, value any) bool {
		sessions = append(sessions, value.(WsSSHSession))
		return true
	})
	return sessions
}
