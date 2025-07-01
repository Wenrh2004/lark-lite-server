package adapter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hertz-contrib/websocket"
	"go.uber.org/zap"

	"github.com/Wenrh2004/lark-lite-server/internal/collaboration/domain"
	"github.com/Wenrh2004/lark-lite-server/pkg/adapter"
)

// CollaborationAdapter 协同编辑适配器
type CollaborationAdapter struct {
	srv                  *adapter.Service
	collaborationService domain.CollaborationService
	connectionService    domain.ConnectionService
	sessionManager       domain.SessionManager
	wsManager            domain.WebSocketManager
}

// NewCollaborationAdapter 创建协同编辑适配器
func NewCollaborationAdapter(
	srv *adapter.Service,
	collaborationService domain.CollaborationService,
	connectionService domain.ConnectionService,
	sessionManager domain.SessionManager,
	wsManager domain.WebSocketManager,
) *CollaborationAdapter {
	return &CollaborationAdapter{
		srv:                  srv,
		collaborationService: collaborationService,
		connectionService:    connectionService,
		sessionManager:       sessionManager,
		wsManager:            wsManager,
	}
}

// HandleWebSocketConnection 处理WebSocket连接（技术细节）
func (a *CollaborationAdapter) HandleWebSocketConnection(ctx context.Context, conn *websocket.Conn, documentID, userID string) error {
	connectionID := fmt.Sprintf("%s:%s", documentID, userID)

	// 添加WebSocket连接到管理器
	if err := a.wsManager.AddConnection(connectionID, conn, ctx); err != nil {
		return fmt.Errorf("[Adapter.Collaboration] failed to add connection: %w", err)
	}

	defer func() {
		// 清理连接
		a.wsManager.RemoveConnection(connectionID)
		a.sessionManager.RemoveUserSession(connectionID)

		// 处理用户离开
		leaveReq := &domain.LeaveDocument{UserID: userID}
		a.collaborationService.LeaveDocument(ctx, connectionID, leaveReq)
	}()

	// 开始消息处理循环
	return a.messageLoop(ctx, connectionID)
}

// messageLoop 消息处理循环
func (a *CollaborationAdapter) messageLoop(ctx context.Context, connectionID string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 读取消息
			msg, err := a.wsManager.ReadMessage(connectionID)
			if err != nil {
				a.srv.Logger.WithContext(ctx).Error("[Adapter.Collaboration] failed to read message: ", zap.Error(err))
				return err
			}

			// 处理消息
			if err := a.ProcessMessage(ctx, connectionID, msg); err != nil {
				a.srv.Logger.WithContext(ctx).Error("[Adapter.Collaboration] failed to process message: ", zap.Error(err))
				if err := a.sendErrorMessage(connectionID, err.Error()); err != nil {
					a.srv.Logger.WithContext(ctx).Error("[Adapter.Collaboration] failed to send error message: ", zap.Error(err))
				}
			}
		}
	}
}

// ProcessMessage 处理业务消息
func (a *CollaborationAdapter) ProcessMessage(ctx context.Context, connectionID string, msg *domain.Message) error {
	switch msg.Type {
	case domain.MessageTypeJoin:
		return a.HandleJoinMessage(ctx, connectionID, msg.Data)
	case domain.MessageTypeLeave:
		return a.HandleLeaveMessage(ctx, connectionID, msg.Data)
	case domain.MessageTypeAwareness:
		return a.HandleAwarenessMessage(ctx, connectionID, msg.Data)
	case domain.MessageTypeHeartbeat:
		return a.HandleHeartbeatMessage(ctx, connectionID, msg.Data)
	case domain.MessageTypeSync:
		return a.HandleSyncMessage(ctx, connectionID, msg.Data)
	case domain.MessageTypeUpdate:
		return a.HandleUpdateMessage(ctx, connectionID, msg.Data)
	default:
		// 使用现有的连接服务处理其他消息
		return a.connectionService.ProcessMessage(ctx, connectionID, msg)
	}
}

// HandleJoinMessage 处理加入消息
func (a *CollaborationAdapter) HandleJoinMessage(ctx context.Context, connectionID string, data interface{}) error {
	joinReq, ok := data.(*domain.JoinDocument)
	if !ok {
		return fmt.Errorf("[Adapter.Collaboration] invalid join document data")
	}

	if err := a.collaborationService.JoinDocument(ctx, connectionID, joinReq); err != nil {
		return fmt.Errorf("[Adapter.Collaboration] failed to process join document: %w", err)
	}

	// 添加用户会话到会话管理器
	documentID := extractDocumentIDFromConnectionID(connectionID)
	userSession := &domain.UserSession{
		UserID:        joinReq.UserID,
		DocumentID:    documentID,
		ClientID:      generateClientID(joinReq.UserID),
		ConnectionID:  connectionID,
		JoinedAt:      time.Now(),
		LastHeartbeat: time.Now(),
		UserInfo: &domain.ActiveUser{
			UserID:       joinReq.UserID,
			UserName:     joinReq.UserName,
			UserColor:    joinReq.UserColor,
			IsActive:     true,
			UserMetadata: joinReq.UserMetadata,
			LastSeen:     time.Now(),
		},
		IsActive: true,
	}
	a.sessionManager.AddUserSession(userSession)

	return nil
}

// HandleLeaveMessage 处理离开消息
func (a *CollaborationAdapter) HandleLeaveMessage(ctx context.Context, connectionID string, data interface{}) error {
	leaveReq, ok := data.(*domain.LeaveDocument)
	if !ok {
		return fmt.Errorf("[Adapter.Collaboration] invalid leave document data")
	}
	return a.collaborationService.LeaveDocument(ctx, connectionID, leaveReq)
}

// HandleAwarenessMessage 处理感知信息消息
func (a *CollaborationAdapter) HandleAwarenessMessage(ctx context.Context, connectionID string, data interface{}) error {
	awareness, ok := data.(*domain.AwarenessUpdate)
	if !ok {
		return fmt.Errorf("[Adapter.Collaboration] invalid awareness update data")
	}
	return a.collaborationService.ProcessAwarenessUpdate(ctx, connectionID, awareness)
}

// HandleHeartbeatMessage 处理心跳消息
func (a *CollaborationAdapter) HandleHeartbeatMessage(ctx context.Context, connectionID string, data interface{}) error {
	heartbeat, ok := data.(*domain.HeartBeat)
	if !ok {
		return fmt.Errorf("[Adapter.Collaboration] invalid heartbeat data")
	}
	return a.collaborationService.ProcessHeartbeat(ctx, connectionID, heartbeat)
}

// HandleSyncMessage 处理同步消息
func (a *CollaborationAdapter) HandleSyncMessage(ctx context.Context, connectionID string, data interface{}) error {
	syncReq, ok := data.(*domain.SyncRequest)
	if !ok {
		return fmt.Errorf("[Adapter.Collaboration] invalid sync request data")
	}
	return a.collaborationService.ProcessSyncRequest(ctx, connectionID, syncReq)
}

// HandleUpdateMessage 处理更新消息
func (a *CollaborationAdapter) HandleUpdateMessage(ctx context.Context, connectionID string, data interface{}) error {
	updateMsg, ok := data.(*domain.UpdateMessage)
	if !ok {
		return fmt.Errorf("[Adapter.Collaboration] invalid update message data")
	}
	return a.collaborationService.ProcessUpdateMessage(ctx, connectionID, updateMsg)
}

// GetActiveUsers 获取活跃用户列表
func (a *CollaborationAdapter) GetActiveUsers(ctx context.Context, documentID string) ([]*domain.ActiveUser, error) {
	return a.collaborationService.GetActiveUsers(ctx, documentID)
}

// GetDocumentStats 获取文档统计信息
func (a *CollaborationAdapter) GetDocumentStats(ctx context.Context, documentID string) (*domain.DocumentStats, error) {
	return a.collaborationService.GetDocumentStats(ctx, documentID)
}

// GetConnectionStats 获取连接统计信息
func (a *CollaborationAdapter) GetConnectionStats(ctx context.Context) (*domain.ConnectionStats, error) {
	return a.connectionService.GetConnectionStats(ctx)
}

// LeaveSession 用户离开会话
func (a *CollaborationAdapter) LeaveSession(ctx context.Context, connectionID string, userID string) error {
	a.sessionManager.RemoveUserSession(connectionID)

	leaveReq := &domain.LeaveDocument{UserID: userID}
	return a.collaborationService.LeaveDocument(ctx, connectionID, leaveReq)
}

// sendErrorMessage 发送错误消息给客户端
func (a *CollaborationAdapter) sendErrorMessage(connectionID, errorMsg string) error {
	errorMessage := &domain.Message{
		Type:      domain.MessageTypeBroadcast,
		Data:      map[string]string{"error": errorMsg},
		Timestamp: time.Now(),
	}

	if err := a.wsManager.WriteMessage(connectionID, errorMessage); err != nil {
		return fmt.Errorf("[Adapter.Collaboration] failed to send error message: %w", err)
	}

	return nil
}

// 辅助函数
func extractDocumentIDFromConnectionID(connectionID string) string {
	parts := strings.Split(connectionID, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func generateClientID(userID string) string {
	return fmt.Sprintf("client_%s_%d", userID, time.Now().UnixNano())
}
