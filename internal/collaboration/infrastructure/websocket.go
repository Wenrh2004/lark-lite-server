package infrastructure

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hertz-contrib/websocket"

	"github.com/Wenrh2004/lark-lite-server/internal/collaboration/domain"
)

// WebSocketConnection WebSocket连接包装
type WebSocketConnection struct {
	conn   *websocket.Conn
	mu     sync.Mutex
	closed bool
	ctx    context.Context
	cancel context.CancelFunc
}

// WebSocketBroadcaster WebSocket消息广播器，实现Domain接口
type WebSocketBroadcaster struct {
	connections map[string]*WebSocketConnection
	mu          sync.RWMutex
	repo        ConnectionStatsUpdater
}

// ConnectionStatsUpdater 连接统计更新接口
type ConnectionStatsUpdater interface {
	IncrementMessagesSent()
	IncrementMessagesReceived()
}

// NewWebSocketBroadcaster 创建WebSocket广播器
func NewWebSocketBroadcaster(repo ConnectionStatsUpdater) *WebSocketBroadcaster {
	return &WebSocketBroadcaster{
		connections: make(map[string]*WebSocketConnection),
		repo:        repo,
	}
}

// 实现 domain.WebSocketManager 接口

// AddConnection 添加WebSocket连接
func (b *WebSocketBroadcaster) AddConnection(connectionID string, conn *websocket.Conn, ctx context.Context) error {
	connCtx, cancel := context.WithCancel(ctx)

	wsConn := &WebSocketConnection{
		conn:   conn,
		ctx:    connCtx,
		cancel: cancel,
	}

	b.mu.Lock()
	b.connections[connectionID] = wsConn
	b.mu.Unlock()

	return nil
}

// RemoveConnection 移除WebSocket连接
func (b *WebSocketBroadcaster) RemoveConnection(connectionID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if wsConn, exists := b.connections[connectionID]; exists {
		wsConn.close()
		delete(b.connections, connectionID)
	}
	return nil
}

// GetConnection 获取WebSocket连接
func (b *WebSocketBroadcaster) GetConnection(connectionID string) (domain.WebSocketConnection, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	conn, exists := b.connections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection %s not found", connectionID)
	}
	return conn, nil
}

// WriteMessage 向指定连接写入消息
func (b *WebSocketBroadcaster) WriteMessage(connectionID string, msg *domain.Message) error {
	wsConn, err := b.GetConnection(connectionID)
	if err != nil {
		return err
	}
	return wsConn.WriteMessage(msg)
}

// ReadMessage 从指定连接读取消息
func (b *WebSocketBroadcaster) ReadMessage(connectionID string) (*domain.Message, error) {
	wsConn, err := b.GetConnection(connectionID)
	if err != nil {
		return nil, err
	}
	return wsConn.ReadMessage()
}

// 实现 domain.MessageBroadcaster 接口

// BroadcastToDocument 向文档的所有连接广播消息
func (b *WebSocketBroadcaster) BroadcastToDocument(ctx context.Context, docID string, msg *domain.Message, excludeUserID string) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	count := 0
	for connectionID, wsConn := range b.connections {
		// 从connectionID解析出docID和userID
		if b.shouldSendToConnection(connectionID, docID, excludeUserID) {
			if err := wsConn.WriteMessage(msg); err != nil {
				log.Printf("Failed to broadcast to connection %s: %v", connectionID, err)
			} else {
				count++
			}
		}
	}

	// 更新统计
	if b.repo != nil {
		for i := 0; i < count; i++ {
			b.repo.IncrementMessagesSent()
		}
	}

	log.Printf("Broadcasted message to %d connections in document %s", count, docID)
	return nil
}

// SendToConnection 向特定连接发送消息
func (b *WebSocketBroadcaster) SendToConnection(ctx context.Context, connectionID string, msg *domain.Message) error {
	b.mu.RLock()
	wsConn, exists := b.connections[connectionID]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("connection %s not found", connectionID)
	}

	err := wsConn.WriteMessage(msg)
	if err == nil && b.repo != nil {
		b.repo.IncrementMessagesSent()
	}

	return err
}

// shouldSendToConnection 判断是否应该向连接发送消息
func (b *WebSocketBroadcaster) shouldSendToConnection(connectionID, docID, excludeUserID string) bool {
	// connectionID 格式: docID:userID
	// 简单的解析逻辑，可以根据实际需要优化
	if len(connectionID) < len(docID)+1 {
		return false
	}

	if connectionID[:len(docID)] != docID {
		return false
	}

	if excludeUserID != "" {
		userIDStart := len(docID) + 1
		if len(connectionID) > userIDStart {
			userID := connectionID[userIDStart:]
			if userID == excludeUserID {
				return false
			}
		}
	}

	return true
}

// WriteMessage 写入消息
func (w *WebSocketConnection) WriteMessage(msg *domain.Message) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return fmt.Errorf("connection is closed")
	}

	// 设置写入超时
	_ = w.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return w.conn.WriteJSON(msg)
}

// ReadMessage 读取消息
func (w *WebSocketConnection) ReadMessage() (*domain.Message, error) {
	var msg domain.Message
	err := w.conn.ReadJSON(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// Context 获取连接上下文
func (w *WebSocketConnection) Context() context.Context {
	return w.ctx
}

// IsClosed 检查连接是否已关闭
func (w *WebSocketConnection) IsClosed() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.closed
}

// Close 关闭连接
func (w *WebSocketConnection) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.closed {
		w.closed = true
		w.cancel()
		_ = w.conn.Close()
	}
	return nil
}

// close 内部关闭方法
func (w *WebSocketConnection) close() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.closed {
		w.closed = true
		w.cancel()
		_ = w.conn.Close()
	}
}

// CollaborationBroadcaster 协同编辑消息广播器实现
// 这个类属于Infrastructure层，实现Domain层的接口
type CollaborationBroadcaster struct {
	wsManager      *WebSocketBroadcaster
	userSessions   map[string]*domain.UserSession // connectionID -> UserSession
	docConnections map[string][]string            // documentID -> []connectionID
	mu             sync.RWMutex
}

// NewCollaborationBroadcaster 创建协同编辑消息广播器
func NewCollaborationBroadcaster(wsManager *WebSocketBroadcaster) *CollaborationBroadcaster {
	return &CollaborationBroadcaster{
		wsManager:      wsManager,
		userSessions:   make(map[string]*domain.UserSession),
		docConnections: make(map[string][]string),
	}
}

// BroadcastToDocument 广播消息到文档的所有用户
func (b *CollaborationBroadcaster) BroadcastToDocument(_ context.Context, documentID string, msg *domain.CollaborationMessage, excludeUserID ...string) error {
	b.mu.RLock()
	connections := b.docConnections[documentID]
	b.mu.RUnlock()

	excludeMap := make(map[string]bool)
	for _, userID := range excludeUserID {
		excludeMap[userID] = true
	}

	var errors []error
	for _, connectionID := range connections {
		b.mu.RLock()
		session := b.userSessions[connectionID]
		b.mu.RUnlock()

		if session != nil && !excludeMap[session.UserID] {
			if err := b.sendMessageToConnection(connectionID, msg); err != nil {
				errors = append(errors, fmt.Errorf("[Infrastructure.CollaborationBroadcaster] failed to send to connection %s: %w", connectionID, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("[Infrastructure.CollaborationBroadcaster] failed to broadcast to some connections: %v", errors)
	}

	return nil
}

// SendToUser 发送消息给特定用户
func (b *CollaborationBroadcaster) SendToUser(_ context.Context, userID string, msg *domain.CollaborationMessage) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for connectionID, session := range b.userSessions {
		if session.UserID == userID && session.IsActive {
			return b.sendMessageToConnection(connectionID, msg)
		}
	}

	return fmt.Errorf("[Infrastructure.CollaborationBroadcaster] user %s not found or not active", userID)
}

// SendToConnection 发送消息给特定连接
func (b *CollaborationBroadcaster) SendToConnection(_ context.Context, connectionID string, msg *domain.CollaborationMessage) error {
	return b.sendMessageToConnection(connectionID, msg)
}

// GetDocumentConnectionCount 获取文档的连接数
func (b *CollaborationBroadcaster) GetDocumentConnectionCount(_ context.Context, documentID string) (int, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	connections := b.docConnections[documentID]
	activeCount := 0

	for _, connectionID := range connections {
		if session := b.userSessions[connectionID]; session != nil && session.IsActive {
			activeCount++
		}
	}

	return activeCount, nil
}

// 实现 domain.SessionManager 接口

// AddUserSession 添加用户会话
func (b *CollaborationBroadcaster) AddUserSession(session *domain.UserSession) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.userSessions[session.ConnectionID] = session

	// 添加到文档连接列表
	documentID := session.DocumentID
	if b.docConnections[documentID] == nil {
		b.docConnections[documentID] = make([]string, 0)
	}

	// 检查是否已存在
	exists := false
	for _, connID := range b.docConnections[documentID] {
		if connID == session.ConnectionID {
			exists = true
			break
		}
	}

	if !exists {
		b.docConnections[documentID] = append(b.docConnections[documentID], session.ConnectionID)
	}
}

// RemoveUserSession 移除用户会话
func (b *CollaborationBroadcaster) RemoveUserSession(connectionID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	session := b.userSessions[connectionID]
	if session != nil {
		documentID := session.DocumentID

		// 从文档连接列表移除
		if connections := b.docConnections[documentID]; connections != nil {
			for i, connID := range connections {
				if connID == connectionID {
					b.docConnections[documentID] = append(connections[:i], connections[i+1:]...)
					break
				}
			}

			// 如果文档没有连接了，清理
			if len(b.docConnections[documentID]) == 0 {
				delete(b.docConnections, documentID)
			}
		}
	}

	delete(b.userSessions, connectionID)
}

// GetUserSession 获取用户会话
func (b *CollaborationBroadcaster) GetUserSession(connectionID string) (*domain.UserSession, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	session, exists := b.userSessions[connectionID]
	if !exists {
		return nil, false
	}

	// 返回副本
	sessionCopy := *session
	if session.UserInfo != nil {
		userInfoCopy := *session.UserInfo
		sessionCopy.UserInfo = &userInfoCopy
	}

	return &sessionCopy, true
}

// GetDocumentSessions 获取文档的所有会话
func (b *CollaborationBroadcaster) GetDocumentSessions(documentID string) []*domain.UserSession {
	b.mu.RLock()
	defer b.mu.RUnlock()

	connections := b.docConnections[documentID]
	var sessions []*domain.UserSession

	for _, connectionID := range connections {
		if session, exists := b.userSessions[connectionID]; exists {
			sessionCopy := *session
			if session.UserInfo != nil {
				userInfoCopy := *session.UserInfo
				sessionCopy.UserInfo = &userInfoCopy
			}
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions
}

// sendMessageToConnection 发送消息到指定连接
func (b *CollaborationBroadcaster) sendMessageToConnection(connectionID string, msg *domain.CollaborationMessage) error {
	// 将协同编辑消息转换为WebSocket消息
	wsMsg := &domain.Message{
		Type:      convertCollaborationMsgType(msg.MessageType),
		DocID:     msg.DocumentID,
		UserID:    msg.UserID,
		Data:      msg.Content,
		Timestamp: time.Unix(msg.Timestamp, 0),
	}

	// 直接调用同层的WebSocketBroadcaster方法，避免接口调用
	return b.wsManager.WriteMessage(connectionID, wsMsg)
}

// convertCollaborationMsgType 转换消息类型
func convertCollaborationMsgType(msgType domain.CollaborationMsgType) domain.MessageType {
	switch msgType {
	case domain.MsgTypeAwareness:
		return domain.MessageTypeAwareness
	case domain.MsgTypeJoinDocument, domain.MsgTypeUserJoined:
		return domain.MessageTypeJoin
	case domain.MsgTypeLeaveDocument, domain.MsgTypeUserLeft:
		return domain.MessageTypeLeave
	case domain.MsgTypeHeartbeat:
		return domain.MessageTypeHeartbeat
	case domain.MsgTypeSync, domain.MsgTypeUpdate:
		return domain.MessageTypeSync
	default:
		return domain.MessageTypeBroadcast
	}
}
