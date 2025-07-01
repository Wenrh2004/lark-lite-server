package domain

import (
	"time"
)

// Connection 连接聚合根
type Connection struct {
	ID        string
	DocID     string
	UserID    string
	CreatedAt time.Time
	LastPing  time.Time
	Status    ConnectionStatus
}

// ConnectionStatus 连接状态
type ConnectionStatus int

const (
	ConnectionStatusConnected ConnectionStatus = iota
	ConnectionStatusDisconnected
	ConnectionStatusTimeout
)

// Message 消息值对象
type Message struct {
	Type      MessageType `json:"type"`
	DocID     string      `json:"doc_id"`
	UserID    string      `json:"user_id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// MessageType 消息类型
type MessageType string

const (
	MessageTypeJoin      MessageType = "join"
	MessageTypeLeave     MessageType = "leave"
	MessageTypeHeartbeat MessageType = "heartbeat"
	MessageTypeBroadcast MessageType = "broadcast"

	MessageTypeSync      MessageType = "sync"
	MessageTypeUpdate    MessageType = "update"
	MessageTypeAwareness MessageType = "awareness"
)

// ConnectionStats 连接统计
type ConnectionStats struct {
	TotalConnections int64 `json:"total_connections"`
	MessagesSent     int64 `json:"messages_sent"`
	MessagesReceived int64 `json:"messages_received"`
	ShardCount       int   `json:"shard_count"`
}

type ConnectionEstablishedEvent struct {
	ConnectionID string
	DocID        string
	UserID       string
	Timestamp    time.Time
}

type ConnectionClosedEvent struct {
	ConnectionID string
	DocID        string
	UserID       string
	Timestamp    time.Time
}

type MessageReceivedEvent struct {
	ConnectionID string
	Message      *Message
	Timestamp    time.Time
}

// CollaborationMessage 协同编辑消息
type CollaborationMessage struct {
	UserID      string               `json:"user_id"`
	DocumentID  string               `json:"document_id"`
	Timestamp   int64                `json:"timestamp"`
	MessageType CollaborationMsgType `json:"message_type"`
	Content     interface{}          `json:"content"`
}

// CollaborationMsgType 协同编辑消息类型
type CollaborationMsgType string

const (
	MsgTypeAwareness     CollaborationMsgType = "awareness"
	MsgTypeJoinDocument  CollaborationMsgType = "join_document"
	MsgTypeLeaveDocument CollaborationMsgType = "leave_document"
	MsgTypeHeartbeat     CollaborationMsgType = "heartbeat"
	MsgTypeSync          CollaborationMsgType = "sync"
	MsgTypeUpdate        CollaborationMsgType = "update"

	MsgTypeUserJoined CollaborationMsgType = "user_joined"
	MsgTypeUserLeft   CollaborationMsgType = "user_left"
	MsgTypeActiveUser CollaborationMsgType = "active_user"
	MsgTypeError      CollaborationMsgType = "error"
)

// AwarenessUpdate 用户感知信息更新（光标位置、选择等）
type AwarenessUpdate struct {
	ClientID       string `json:"client_id"`
	UserInfo       string `json:"user_info"`       // JSON格式用户信息
	AwarenessState string `json:"awareness_state"` // JSON格式感知状态
	Timestamp      int64  `json:"timestamp"`
}

// JoinDocument 加入文档
type JoinDocument struct {
	UserID       string            `json:"user_id"`
	UserName     string            `json:"user_name"`
	UserColor    string            `json:"user_color"`
	UserMetadata map[string]string `json:"user_metadata"`
}

// LeaveDocument 离开文档
type LeaveDocument struct {
	UserID string `json:"user_id"`
}

// HeartBeat 心跳消息
type HeartBeat struct {
	Timestamp int64 `json:"timestamp"`
}

// UserJoined 用户加入通知
type UserJoined struct {
	UserID       string            `json:"user_id"`
	UserName     string            `json:"user_name"`
	UserColor    string            `json:"user_color"`
	ClientID     string            `json:"client_id"`
	UserMetadata map[string]string `json:"user_metadata"`
}

// UserLeft 用户离开通知
type UserLeft struct {
	UserID   string `json:"user_id"`
	ClientID string `json:"client_id"`
}

// ActiveUser 活跃用户信息
type ActiveUser struct {
	UserID       string            `json:"user_id"`
	UserName     string            `json:"user_name"`
	UserColor    string            `json:"user_color"`
	IsActive     bool              `json:"is_active"`
	UserMetadata map[string]string `json:"user_metadata"`
	LastSeen     time.Time         `json:"last_seen"`
}

// DocumentSession 文档协作会话
type DocumentSession struct {
	DocumentID   string                 `json:"document_id"`
	ActiveUsers  map[string]*ActiveUser `json:"active_users"`
	CreatedAt    time.Time              `json:"created_at"`
	LastActivity time.Time              `json:"last_activity"`
	StateVector  []byte                 `json:"state_vector"`  // Y.js状态向量
	DocumentData []byte                 `json:"document_data"` // 文档数据
}

// SyncRequest Y.js同步请求
type SyncRequest struct {
	StateVector []byte `json:"state_vector"`
}

// SyncResponse Y.js同步响应
type SyncResponse struct {
	UpdateData []byte `json:"update_data"`
}

// UpdateMessage Y.js更新消息
type UpdateMessage struct {
	UpdateData     []byte `json:"update_data"`
	SequenceNumber int64  `json:"sequence_number"`
}

// UserSession 用户会话信息
type UserSession struct {
	UserID        string      `json:"user_id"`
	DocumentID    string      `json:"document_id"`
	ClientID      string      `json:"client_id"`
	ConnectionID  string      `json:"connection_id"`
	JoinedAt      time.Time   `json:"joined_at"`
	LastHeartbeat time.Time   `json:"last_heartbeat"`
	UserInfo      *ActiveUser `json:"user_info"`
	IsActive      bool        `json:"is_active"`
}

// DocumentStats 文档统计信息
type DocumentStats struct {
	DocumentID       string    `json:"document_id"`
	ActiveUserCount  int       `json:"active_user_count"`
	TotalConnections int       `json:"total_connections"`
	MessageCount     int64     `json:"message_count"`
	LastActivity     time.Time `json:"last_activity"`
}
