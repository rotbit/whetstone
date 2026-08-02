// Package conn 管理 app-apis 中的 WebSocket 连接生命周期。
//
// 面试考点（docs/技术方案.md §12）：
//   - goroutine-per-conn 模型与内存开销估算
//   - 心跳保活与半开连接检测（read deadline + ping/pong）
//   - 断线 30s 内凭 sessionId 恢复会话（状态在 Redis，见 interview-rpc）
//   - 优雅关闭：广播关闭帧 → 等待在途消息落盘 → 退出
package conn

import (
	"sync"
	"time"
)

// Session 一条活跃的面试连接。
type Session struct {
	SessionId string
	UserId    int64
	LastPing  time.Time
	// TODO: 持有 *websocket.Conn 与发送队列（带缓冲 channel，写协程独占写）
}

// Manager 并发安全的连接注册表。
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewManager() *Manager {
	return &Manager{sessions: make(map[string]*Session)}
}

func (m *Manager) Register(s *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[s.SessionId] = s
}

func (m *Manager) Unregister(sessionId string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionId)
}

func (m *Manager) Get(sessionId string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[sessionId]
	return s, ok
}

// TODO(M1): 心跳巡检 goroutine —— 定期扫描 LastPing 超时的连接并关闭；
// TODO(M2): 多副本部署时，会话路由改为 sticky（网关层一致性哈希）。
