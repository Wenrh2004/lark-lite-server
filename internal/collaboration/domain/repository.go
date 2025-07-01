package domain

import (
	"context"
	"time"

	"github.com/hertz-contrib/websocket"
)

// ConnectionRepository 连接仓储接口
type ConnectionRepository interface {
	Add(ctx context.Context, conn *Connection) error
	Remove(ctx context.Context, connectionID string) error
	GetByDocID(ctx context.Context, docID string) ([]*Connection, error)
	GetByID(ctx context.Context, connectionID string) (*Connection, error)
	UpdateLastPing(ctx context.Context, connectionID string, pingTime time.Time) error
	GetStats(ctx context.Context) (*ConnectionStats, error)
	CleanupTimeout(ctx context.Context, timeout time.Duration) (int, error)
}

// MessageBroadcaster 消息广播接口
type MessageBroadcaster interface {
	BroadcastToDocument(ctx context.Context, docID string, msg *Message, excludeUserID string) error
	SendToConnection(ctx context.Context, connectionID string, msg *Message) error
}

// WebSocketManager WebSocket连接管理接口（新增）
type WebSocketManager interface {
	AddConnection(connectionID string, conn *websocket.Conn, ctx context.Context) error
	RemoveConnection(connectionID string) error
	GetConnection(connectionID string) (WebSocketConnection, error)
	WriteMessage(connectionID string, msg *Message) error
	ReadMessage(connectionID string) (*Message, error)
}

// WebSocketConnection WebSocket连接接口（新增）
type WebSocketConnection interface {
	WriteMessage(msg *Message) error
	ReadMessage() (*Message, error)
	Context() context.Context
	IsClosed() bool
	Close() error
}

// CollaborationRepository 协同编辑仓储接口
type CollaborationRepository interface {
	// GetDocumentSession 文档会话管理
	GetDocumentSession(ctx context.Context, documentID string) (*DocumentSession, error)
	SaveDocumentSession(ctx context.Context, session *DocumentSession) error
	DeleteDocumentSession(ctx context.Context, documentID string) error

	// GetUserSession 用户会话管理
	GetUserSession(ctx context.Context, connectionID string) (*UserSession, error)
	SaveUserSession(ctx context.Context, session *UserSession) error
	DeleteUserSession(ctx context.Context, connectionID string) error
	GetDocumentUserSessions(ctx context.Context, documentID string) ([]*UserSession, error)

	// GetActiveUsers 活跃用户管理
	GetActiveUsers(ctx context.Context, documentID string) ([]*ActiveUser, error)
	AddActiveUser(ctx context.Context, documentID string, user *ActiveUser) error
	RemoveActiveUser(ctx context.Context, documentID string, userID string) error
	UpdateUserLastSeen(ctx context.Context, documentID, userID string, lastSeen time.Time) error

	// GetDocumentStats 文档统计
	GetDocumentStats(ctx context.Context, documentID string) (*DocumentStats, error)
	UpdateDocumentStats(ctx context.Context, stats *DocumentStats) error

	// SaveMessage 消息存储（可选，用于审计）
	SaveMessage(ctx context.Context, msg *CollaborationMessage) error
	GetRecentMessages(ctx context.Context, documentID string, limit int) ([]*CollaborationMessage, error)
}

// SyncServiceClient Sync服务RPC客户端接口（调用Rust服务）
type SyncServiceClient interface {
	// Sync 同步文档状态
	Sync(ctx context.Context, documentID string, req *SyncRequest) (*SyncResponse, error)
	// ProcessUpdate 处理文档更新
	ProcessUpdate(ctx context.Context, documentID string, update *UpdateMessage) error
	// GetStateVector 获取文档状态向量
	GetStateVector(ctx context.Context, documentID string) ([]byte, error)
}

// CollaborationBroadcaster 协同编辑消息广播器
type CollaborationBroadcaster interface {
	// BroadcastToDocument 广播消息到文档的所有用户
	BroadcastToDocument(ctx context.Context, documentID string, msg *CollaborationMessage, excludeUserID ...string) error
	// SendToUser 发送消息给特定用户
	SendToUser(ctx context.Context, userID string, msg *CollaborationMessage) error
	// SendToConnection 发送消息给特定连接
	SendToConnection(ctx context.Context, connectionID string, msg *CollaborationMessage) error
	// GetDocumentConnectionCount 获取文档的连接数
	GetDocumentConnectionCount(ctx context.Context, documentID string) (int, error)
}

// SessionManager 会话管理接口
type SessionManager interface {
	// AddUserSession 添加用户会话
	AddUserSession(session *UserSession)
	// RemoveUserSession 移除用户会话
	RemoveUserSession(connectionID string)
	// GetUserSession 获取用户会话
	GetUserSession(connectionID string) (*UserSession, bool)
	// GetDocumentSessions 获取文档的所有会话
	GetDocumentSessions(documentID string) []*UserSession
}
