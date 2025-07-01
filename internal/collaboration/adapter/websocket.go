package adapter

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"
	"go.uber.org/zap"

	v1 "github.com/Wenrh2004/lark-lite-server/common/api/v1"
	"github.com/Wenrh2004/lark-lite-server/pkg/adapter"
)

// WebSocketAdapter HTTP WebSocket适配器（入口层）
type WebSocketAdapter struct {
	srv                  *adapter.Service
	collaborationAdapter *CollaborationAdapter
	upgrader             websocket.HertzUpgrader
}

// NewWebSocketAdapter 创建WebSocket适配器
func NewWebSocketAdapter(srv *adapter.Service, collaborationAdapter *CollaborationAdapter) *WebSocketAdapter {
	upgrader := websocket.HertzUpgrader{
		CheckOrigin: func(ctx *app.RequestContext) bool {
			// 在生产环境中应该有更严格的检查
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return &WebSocketAdapter{
		srv:                  srv,
		collaborationAdapter: collaborationAdapter,
		upgrader:             upgrader,
	}
}

// HandleWebSocketUpgrade 处理WebSocket升级请求
func (a *WebSocketAdapter) HandleWebSocketUpgrade(ctx context.Context, c *app.RequestContext) {
	documentID := c.Param("doc_id")
	userID := c.Query("user_id")

	if documentID == "" || userID == "" {
		a.srv.Logger.WithContext(ctx).Info("[Adapter.WebSocket] document_id or user_id is missing")
		v1.HandlerError(c, v1.ErrBadRequest)
		return
	}

	err := a.upgrader.Upgrade(c, func(conn *websocket.Conn) {
		defer conn.Close()

		// 直接调用协同编辑适配器处理WebSocket连接
		if err := a.collaborationAdapter.HandleWebSocketConnection(ctx, conn, documentID, userID); err != nil {
			a.srv.Logger.WithContext(ctx).Error("[Adapter.WebSocket] failed to handle WebSocket connection", zap.Error(err))
		}
	})

	if err != nil {
		a.srv.Logger.WithContext(ctx).Error("Failed to upgrade WebSocket: %v\n", zap.Error(err))
		v1.HandlerError(c, v1.ErrInternalServerError)
	}
}

// GetActiveUsers 获取活跃用户列表
func (a *WebSocketAdapter) GetActiveUsers(ctx context.Context, c *app.RequestContext) {
	documentID := c.Query("document_id")
	if documentID == "" {
		a.srv.Logger.WithContext(ctx).Info("[Adapter.WebSocket] document_id is missing")
		v1.HandlerError(c, v1.ErrBadRequest)
		return
	}

	users, err := a.collaborationAdapter.GetActiveUsers(ctx, documentID)
	if err != nil {
		a.srv.Logger.WithContext(ctx).Error("[Adapter.WebSocket] failed to get active users: %v", zap.Error(err))
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}

	v1.HandlerSuccess(c, users)
}

// GetDocumentStats 获取文档统计信息
func (a *WebSocketAdapter) GetDocumentStats(ctx context.Context, c *app.RequestContext) {
	documentID := c.Query("document_id")
	if documentID == "" {
		a.srv.Logger.WithContext(ctx).Info("[Adapter.WebSocket] document_id is missing")
		v1.HandlerError(c, v1.ErrBadRequest)
		return
	}

	stats, err := a.collaborationAdapter.GetDocumentStats(ctx, documentID)
	if err != nil {
		a.srv.Logger.WithContext(ctx).Error("[Adapter.WebSocket] failed to get stats: %v", zap.Error(err))
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}

	v1.HandlerSuccess(c, stats)
}

// GetConnectionStats 获取连接统计信息
func (a *WebSocketAdapter) GetConnectionStats(ctx context.Context, c *app.RequestContext) {
	stats, err := a.collaborationAdapter.GetConnectionStats(ctx)
	if err != nil {
		a.srv.Logger.WithContext(ctx).Error("[Adapter.WebSocket] failed to get connection stats: %v", zap.Error(err))
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}

	v1.HandlerSuccess(c, stats)
}

// GetDocumentConnections 获取文档连接列表
func (a *WebSocketAdapter) GetDocumentConnections(ctx context.Context, c *app.RequestContext) {
	documentID := c.Param("doc_id")
	if documentID == "" {
		a.srv.Logger.WithContext(ctx).Info("[Adapter.WebSocket] document_id is missing")
		v1.HandlerError(c, v1.ErrBadRequest)
		return
	}

	// TODO: Implement the logic to get document connections
	v1.HandlerSuccess(c, []string{})
}
