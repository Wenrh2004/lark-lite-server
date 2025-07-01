package domain

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ConnectionService 连接领域服务接口
type ConnectionService interface {
	EstablishConnection(ctx context.Context, docID, userID string) (*Connection, error)
	CloseConnection(ctx context.Context, connectionID string) error
	ProcessMessage(ctx context.Context, connectionID string, msg *Message) error
	GetDocumentConnections(ctx context.Context, docID string) ([]*Connection, error)
	GetConnectionStats(ctx context.Context) (*ConnectionStats, error)
	CleanupTimeoutConnections(ctx context.Context, timeout time.Duration) (int, error)
}

// connectionService 连接领域服务实现
type connectionService struct {
	connectionRepo ConnectionRepository
	broadcaster    MessageBroadcaster
}

// NewConnectionService 创建连接服务
func NewConnectionService(repo ConnectionRepository, broadcaster MessageBroadcaster) ConnectionService {
	return &connectionService{
		connectionRepo: repo,
		broadcaster:    broadcaster,
	}
}

// EstablishConnection 建立连接
func (s *connectionService) EstablishConnection(ctx context.Context, docID, userID string) (*Connection, error) {
	connectionID := fmt.Sprintf("%s:%s", docID, userID)

	// 检查是否已存在连接
	existing, _ := s.connectionRepo.GetByID(ctx, connectionID)
	if existing != nil {
		return nil, fmt.Errorf("[Domain.Connection] connection already exists for user %s in document %s", userID, docID)
	}

	// 创建新连接
	conn := &Connection{
		ID:        connectionID,
		DocID:     docID,
		UserID:    userID,
		CreatedAt: time.Now(),
		LastPing:  time.Now(),
		Status:    ConnectionStatusConnected,
	}

	err := s.connectionRepo.Add(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("[Domain.Connection] failed to add connection: %w", err)
	}

	// 发送加入消息给其他用户
	joinMsg := &Message{
		Type:      MessageTypeJoin,
		DocID:     docID,
		UserID:    userID,
		Data:      map[string]string{"status": "joined"},
		Timestamp: time.Now(),
	}

	if err := s.broadcaster.BroadcastToDocument(ctx, docID, joinMsg, userID); err != nil {
		return nil, fmt.Errorf("[Domain.Connection] failed to broadcast join message: %w", err)
	}

	return conn, nil
}

// CloseConnection 关闭连接
func (s *connectionService) CloseConnection(ctx context.Context, connectionID string) error {
	conn, err := s.connectionRepo.GetByID(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("[Domain.Connection] connection not found: %w", err)
	}

	// 发送离开消息给其他用户
	leaveMsg := &Message{
		Type:      MessageTypeLeave,
		DocID:     conn.DocID,
		UserID:    conn.UserID,
		Data:      map[string]string{"status": "left"},
		Timestamp: time.Now(),
	}

	if err := s.broadcaster.BroadcastToDocument(ctx, conn.DocID, leaveMsg, conn.UserID); err != nil {
		return fmt.Errorf("[Domain.Connection] failed to broadcast leave message: %w", err)
	}

	return s.connectionRepo.Remove(ctx, connectionID)
}

// ProcessMessage 处理消息
func (s *connectionService) ProcessMessage(ctx context.Context, connectionID string, msg *Message) error {
	conn, err := s.connectionRepo.GetByID(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("[Domain.Connection] connection not found: %w", err)
	}

	// 更新最后ping时间
	if err := s.connectionRepo.UpdateLastPing(ctx, connectionID, time.Now()); err != nil {
		return fmt.Errorf("[Domain.Connection] failed to update last ping: %w", err)
	}

	switch msg.Type {
	case MessageTypeHeartbeat:
		// 处理心跳消息
		return s.handleHeartbeat(ctx, conn)
	case MessageTypeBroadcast:
		// 广播消息给所有用户
		return s.broadcaster.BroadcastToDocument(ctx, conn.DocID, msg, "")
	default:
		return fmt.Errorf("[Domain.Connection] unknown message type: %s", msg.Type)
	}
}

// GetDocumentConnections 获取文档连接数
func (s *connectionService) GetDocumentConnections(ctx context.Context, docID string) ([]*Connection, error) {
	return s.connectionRepo.GetByDocID(ctx, docID)
}

// GetConnectionStats 获取连接统计
func (s *connectionService) GetConnectionStats(ctx context.Context) (*ConnectionStats, error) {
	return s.connectionRepo.GetStats(ctx)
}

// CleanupTimeoutConnections 清理超时连接
func (s *connectionService) CleanupTimeoutConnections(ctx context.Context, timeout time.Duration) (int, error) {
	return s.connectionRepo.CleanupTimeout(ctx, timeout)
}

// handleHeartbeat 处理心跳消息
func (s *connectionService) handleHeartbeat(ctx context.Context, conn *Connection) error {
	response := &Message{
		Type:      MessageTypeHeartbeat,
		DocID:     conn.DocID,
		UserID:    conn.UserID,
		Data:      map[string]string{"status": "pong"},
		Timestamp: time.Now(),
	}

	return s.broadcaster.SendToConnection(ctx, conn.ID, response)
}

// CollaborationService 协同编辑领域服务接口
type CollaborationService interface {
	// JoinDocument 用户加入文档协作
	JoinDocument(ctx context.Context, connectionID string, joinReq *JoinDocument) error
	// LeaveDocument 用户离开文档协作
	LeaveDocument(ctx context.Context, connectionID string, leaveReq *LeaveDocument) error
	// ProcessAwarenessUpdate 处理用户感知信息更新
	ProcessAwarenessUpdate(ctx context.Context, connectionID string, awareness *AwarenessUpdate) error
	// ProcessHeartbeat 处理心跳消息
	ProcessHeartbeat(ctx context.Context, connectionID string, heartbeat *HeartBeat) error
	// ProcessSyncRequest 处理同步请求（调用Rust服务）
	ProcessSyncRequest(ctx context.Context, connectionID string, syncReq *SyncRequest) error
	// ProcessUpdateMessage 处理文档更新（调用Rust服务）
	ProcessUpdateMessage(ctx context.Context, connectionID string, update *UpdateMessage) error
	// GetActiveUsers 获取活跃用户列表
	GetActiveUsers(ctx context.Context, documentID string) ([]*ActiveUser, error)
	// CleanupTimeoutSessions 清理超时连接
	CleanupTimeoutSessions(ctx context.Context, timeout time.Duration) error
	// GetDocumentStats 获取文档统计信息
	GetDocumentStats(ctx context.Context, documentID string) (*DocumentStats, error)
}

// collaborationService 协同编辑服务实现
type collaborationService struct {
	collabRepo       CollaborationRepository
	syncClient       SyncServiceClient
	broadcaster      CollaborationBroadcaster
	wsManager        WebSocketManager
	heartbeatTimeout time.Duration
}

// NewCollaborationService 创建协同编辑服务
func NewCollaborationService(
	collabRepo CollaborationRepository,
	syncClient SyncServiceClient,
	broadcaster CollaborationBroadcaster,
	wsManager WebSocketManager,
) CollaborationService {
	return &collaborationService{
		collabRepo:       collabRepo,
		syncClient:       syncClient,
		broadcaster:      broadcaster,
		wsManager:        wsManager,
		heartbeatTimeout: 15 * time.Second, // 3次心跳超时（5s * 3）
	}
}

// JoinDocument 处理用户加入文档协作
func (s *collaborationService) JoinDocument(ctx context.Context, connectionID string, joinReq *JoinDocument) error {
	// 从connectionID解析documentID (格式: doc_id:user_id)
	documentID := extractDocumentIDFromConnectionID(connectionID)

	// 创建用户会话
	userSession := &UserSession{
		UserID:        joinReq.UserID,
		DocumentID:    documentID,
		ClientID:      generateClientID(joinReq.UserID),
		ConnectionID:  connectionID,
		JoinedAt:      time.Now(),
		LastHeartbeat: time.Now(),
		UserInfo: &ActiveUser{
			UserID:       joinReq.UserID,
			UserName:     joinReq.UserName,
			UserColor:    joinReq.UserColor,
			IsActive:     true,
			UserMetadata: joinReq.UserMetadata,
			LastSeen:     time.Now(),
		},
		IsActive: true,
	}

	// 保存用户会话
	if err := s.collabRepo.SaveUserSession(ctx, userSession); err != nil {
		return fmt.Errorf("[Domain.Connection] failed to save user session: %w", err)
	}

	// 添加到活跃用户列表
	if err := s.collabRepo.AddActiveUser(ctx, documentID, userSession.UserInfo); err != nil {
		return fmt.Errorf("[Domain.Connection] failed to add active user: %w", err)
	}

	// 获取或创建文档会话
	docSession, err := s.collabRepo.GetDocumentSession(ctx, documentID)
	if err != nil {
		docSession = &DocumentSession{
			DocumentID:   documentID,
			ActiveUsers:  make(map[string]*ActiveUser),
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}
	}

	// 更新文档会话
	docSession.ActiveUsers[joinReq.UserID] = userSession.UserInfo
	docSession.LastActivity = time.Now()
	if err := s.collabRepo.SaveDocumentSession(ctx, docSession); err != nil {
		return fmt.Errorf("[Domain.Connection] failed to save document session: %w", err)
	}

	// 广播用户加入事件
	joinedMsg := &CollaborationMessage{
		UserID:      joinReq.UserID,
		DocumentID:  documentID,
		Timestamp:   time.Now().Unix(),
		MessageType: MsgTypeUserJoined,
		Content: &UserJoined{
			UserID:       joinReq.UserID,
			UserName:     joinReq.UserName,
			UserColor:    joinReq.UserColor,
			ClientID:     userSession.ClientID,
			UserMetadata: joinReq.UserMetadata,
		},
	}

	return s.broadcaster.BroadcastToDocument(ctx, documentID, joinedMsg, joinReq.UserID)
}

// LeaveDocument 处理用户离开文档协作
func (s *collaborationService) LeaveDocument(ctx context.Context, connectionID string, leaveReq *LeaveDocument) error {
	// 获取用户会话
	userSession, err := s.collabRepo.GetUserSession(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("[Domain.Connection] failed to get user session: %w", err)
	}

	documentID := userSession.DocumentID

	// 从活跃用户列表移除
	if err := s.collabRepo.RemoveActiveUser(ctx, documentID, leaveReq.UserID); err != nil {
		return fmt.Errorf("[Domain.Connection] failed to remove active user: %w", err)
	}

	// 删除用户会话
	if err := s.collabRepo.DeleteUserSession(ctx, connectionID); err != nil {
		return fmt.Errorf("[Domain.Connection] failed to delete user session: %w", err)
	}

	// 更新文档会话
	docSession, err := s.collabRepo.GetDocumentSession(ctx, documentID)
	if err == nil && docSession != nil {
		delete(docSession.ActiveUsers, leaveReq.UserID)
		docSession.LastActivity = time.Now()
		if err := s.collabRepo.SaveDocumentSession(ctx, docSession); err != nil {
			return fmt.Errorf("[Domain.Connection] failed to save document session: %w", err)
		}
	}

	// 广播用户离开事件
	leftMsg := &CollaborationMessage{
		UserID:      leaveReq.UserID,
		DocumentID:  documentID,
		Timestamp:   time.Now().Unix(),
		MessageType: MsgTypeUserLeft,
		Content: &UserLeft{
			UserID:   leaveReq.UserID,
			ClientID: userSession.ClientID,
		},
	}

	return s.broadcaster.BroadcastToDocument(ctx, documentID, leftMsg)
}

// ProcessAwarenessUpdate 处理用户感知信息更新
func (s *collaborationService) ProcessAwarenessUpdate(ctx context.Context, connectionID string, awareness *AwarenessUpdate) error {
	// 获取用户会话
	userSession, err := s.collabRepo.GetUserSession(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("[Domain.Connection] failed to get user session: %w", err)
	}

	documentID := userSession.DocumentID

	// 更新用户最后活跃时间
	if err := s.collabRepo.UpdateUserLastSeen(ctx, documentID, userSession.UserID, time.Now()); err != nil {
		return fmt.Errorf("[Domain.Connection] failed to update user last seen: %w", err)
	}

	// 广播感知信息更新
	awarenessMsg := &CollaborationMessage{
		UserID:      userSession.UserID,
		DocumentID:  documentID,
		Timestamp:   time.Now().Unix(),
		MessageType: MsgTypeAwareness,
		Content:     awareness,
	}

	return s.broadcaster.BroadcastToDocument(ctx, documentID, awarenessMsg, userSession.UserID)
}

// ProcessHeartbeat 处理心跳消息
func (s *collaborationService) ProcessHeartbeat(ctx context.Context, connectionID string, heartbeat *HeartBeat) error {
	// 获取用户会话
	userSession, err := s.collabRepo.GetUserSession(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("[Domain.Connection] failed to get user session: %w", err)
	}

	// 更新心跳时间
	userSession.LastHeartbeat = time.Now()
	userSession.IsActive = true

	if err := s.collabRepo.SaveUserSession(ctx, userSession); err != nil {
		return fmt.Errorf("[Domain.Connection] failed to save user session: %w", err)
	}

	// 更新用户最后活跃时间
	return s.collabRepo.UpdateUserLastSeen(ctx, userSession.DocumentID, userSession.UserID, time.Now())
}

// ProcessSyncRequest 处理同步请求（调用Rust服务）
func (s *collaborationService) ProcessSyncRequest(ctx context.Context, connectionID string, syncReq *SyncRequest) error {
	// 获取用户会话
	userSession, err := s.collabRepo.GetUserSession(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("[Domain.Connection] failed to get user session: %w", err)
	}

	documentID := userSession.DocumentID

	// 调用Rust同步服务
	syncResp, err := s.syncClient.Sync(ctx, documentID, syncReq)
	if err != nil {
		return fmt.Errorf("[Domain.Connection] failed to sync with rust service: %w", err)
	}

	// 发送同步响应给客户端
	responseMsg := &CollaborationMessage{
		UserID:      userSession.UserID,
		DocumentID:  documentID,
		Timestamp:   time.Now().Unix(),
		MessageType: MsgTypeSync,
		Content:     syncResp,
	}

	return s.broadcaster.SendToConnection(ctx, connectionID, responseMsg)
}

// ProcessUpdateMessage 处理文档更新（调用Rust服务）
func (s *collaborationService) ProcessUpdateMessage(ctx context.Context, connectionID string, update *UpdateMessage) error {
	// 获取用户会话
	userSession, err := s.collabRepo.GetUserSession(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("[Domain.Connection] failed to get user session: %w", err)
	}

	documentID := userSession.DocumentID

	// 调用Rust更新服务
	if err := s.syncClient.ProcessUpdate(ctx, documentID, update); err != nil {
		return fmt.Errorf("[Domain.Connection] failed to process update with rust service: %w", err)
	}

	// 广播更新给其他用户
	updateMsg := &CollaborationMessage{
		UserID:      userSession.UserID,
		DocumentID:  documentID,
		Timestamp:   time.Now().Unix(),
		MessageType: MsgTypeUpdate,
		Content:     update,
	}

	return s.broadcaster.BroadcastToDocument(ctx, documentID, updateMsg, userSession.UserID)
}

// GetActiveUsers 获取活跃用户列表
func (s *collaborationService) GetActiveUsers(ctx context.Context, documentID string) ([]*ActiveUser, error) {
	return s.collabRepo.GetActiveUsers(ctx, documentID)
}

// CleanupTimeoutSessions 清理超时会话
func (s *collaborationService) CleanupTimeoutSessions(ctx context.Context, timeout time.Duration) error {
	// 这里可以实现定期清理逻辑
	// TODO: 获取所有超时的会话并清理
	return nil
}

// GetDocumentStats 获取文档统计信息
func (s *collaborationService) GetDocumentStats(ctx context.Context, documentID string) (*DocumentStats, error) {
	return s.collabRepo.GetDocumentStats(ctx, documentID)
}

// 辅助函数
func generateClientID(userID string) string {
	return fmt.Sprintf("client_%s_%d", userID, time.Now().UnixNano())
}

func extractDocumentIDFromConnectionID(connectionID string) string {
	// 假设connectionID格式为 "doc_id:user_id"
	parts := strings.Split(connectionID, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
