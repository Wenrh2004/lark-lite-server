package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bytedance/sonic"

	"github.com/Wenrh2004/lark-lite-server/internal/collaboration/domain"
)

// collaborationRepository 协同编辑仓储实现
// 使用内存存储，生产环境建议使用Redis等持久化存储
type collaborationRepository struct {
	documentSessions map[string]*domain.DocumentSession
	userSessions     map[string]*domain.UserSession
	activeUsers      map[string]map[string]*domain.ActiveUser // documentID -> userID -> ActiveUser
	documentStats    map[string]*domain.DocumentStats
	messages         map[string][]*domain.CollaborationMessage // documentID -> messages
	mu               sync.RWMutex
}

// NewCollaborationRepository 创建协同编辑仓储
func NewCollaborationRepository() domain.CollaborationRepository {
	return &collaborationRepository{
		documentSessions: make(map[string]*domain.DocumentSession),
		userSessions:     make(map[string]*domain.UserSession),
		activeUsers:      make(map[string]map[string]*domain.ActiveUser),
		documentStats:    make(map[string]*domain.DocumentStats),
		messages:         make(map[string][]*domain.CollaborationMessage),
	}
}

// GetDocumentSession 获取文档会话
func (r *collaborationRepository) GetDocumentSession(ctx context.Context, documentID string) (*domain.DocumentSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.documentSessions[documentID]
	if !exists {
		return nil, fmt.Errorf("document session not found: %s", documentID)
	}

	// 深拷贝返回
	sessionCopy := *session
	sessionCopy.ActiveUsers = make(map[string]*domain.ActiveUser)
	for k, v := range session.ActiveUsers {
		userCopy := *v
		sessionCopy.ActiveUsers[k] = &userCopy
	}

	return &sessionCopy, nil
}

// SaveDocumentSession 保存文档会话
func (r *collaborationRepository) SaveDocumentSession(ctx context.Context, session *domain.DocumentSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 深拷贝保存
	sessionCopy := *session
	sessionCopy.ActiveUsers = make(map[string]*domain.ActiveUser)
	for k, v := range session.ActiveUsers {
		userCopy := *v
		sessionCopy.ActiveUsers[k] = &userCopy
	}

	r.documentSessions[session.DocumentID] = &sessionCopy
	return nil
}

// DeleteDocumentSession 删除文档会话
func (r *collaborationRepository) DeleteDocumentSession(ctx context.Context, documentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.documentSessions, documentID)
	delete(r.activeUsers, documentID)
	delete(r.documentStats, documentID)
	delete(r.messages, documentID)
	return nil
}

// GetUserSession 获取用户会话
func (r *collaborationRepository) GetUserSession(ctx context.Context, connectionID string) (*domain.UserSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.userSessions[connectionID]
	if !exists {
		return nil, fmt.Errorf("user session not found: %s", connectionID)
	}

	// 深拷贝返回
	sessionCopy := *session
	if session.UserInfo != nil {
		userInfoCopy := *session.UserInfo
		sessionCopy.UserInfo = &userInfoCopy
	}

	return &sessionCopy, nil
}

// SaveUserSession 保存用户会话
func (r *collaborationRepository) SaveUserSession(_ context.Context, session *domain.UserSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 深拷贝保存
	sessionCopy := *session
	if session.UserInfo != nil {
		userInfoCopy := *session.UserInfo
		sessionCopy.UserInfo = &userInfoCopy
	}

	r.userSessions[session.ConnectionID] = &sessionCopy
	return nil
}

// DeleteUserSession 删除用户会话
func (r *collaborationRepository) DeleteUserSession(_ context.Context, connectionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.userSessions, connectionID)
	return nil
}

// GetDocumentUserSessions 获取文档的所有用户会话
func (r *collaborationRepository) GetDocumentUserSessions(_ context.Context, documentID string) ([]*domain.UserSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sessions []*domain.UserSession
	for _, session := range r.userSessions {
		if session.DocumentID == documentID {
			sessionCopy := *session
			if session.UserInfo != nil {
				userInfoCopy := *session.UserInfo
				sessionCopy.UserInfo = &userInfoCopy
			}
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions, nil
}

// GetActiveUsers 获取活跃用户
func (r *collaborationRepository) GetActiveUsers(_ context.Context, documentID string) ([]*domain.ActiveUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users, exists := r.activeUsers[documentID]
	if !exists {
		return []*domain.ActiveUser{}, nil
	}

	var activeUsers []*domain.ActiveUser
	for _, user := range users {
		userCopy := *user
		activeUsers = append(activeUsers, &userCopy)
	}

	return activeUsers, nil
}

// AddActiveUser 添加活跃用户
func (r *collaborationRepository) AddActiveUser(_ context.Context, documentID string, user *domain.ActiveUser) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.activeUsers[documentID] == nil {
		r.activeUsers[documentID] = make(map[string]*domain.ActiveUser)
	}

	userCopy := *user
	r.activeUsers[documentID][user.UserID] = &userCopy

	// 更新文档统计
	r.updateDocumentStatsUnsafe(documentID)
	return nil
}

// RemoveActiveUser 移除活跃用户
func (r *collaborationRepository) RemoveActiveUser(_ context.Context, documentID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if users, exists := r.activeUsers[documentID]; exists {
		delete(users, userID)
		if len(users) == 0 {
			delete(r.activeUsers, documentID)
		}
	}

	// 更新文档统计
	r.updateDocumentStatsUnsafe(documentID)
	return nil
}

// UpdateUserLastSeen 更新用户最后活跃时间
func (r *collaborationRepository) UpdateUserLastSeen(_ context.Context, documentID, userID string, lastSeen time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if users, exists := r.activeUsers[documentID]; exists {
		if user, userExists := users[userID]; userExists {
			user.LastSeen = lastSeen
			user.IsActive = true
		}
	}

	return nil
}

// GetDocumentStats 获取文档统计
func (r *collaborationRepository) GetDocumentStats(_ context.Context, documentID string) (*domain.DocumentStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats, exists := r.documentStats[documentID]
	if !exists {
		return &domain.DocumentStats{
			DocumentID:       documentID,
			ActiveUserCount:  0,
			TotalConnections: 0,
			MessageCount:     0,
			LastActivity:     time.Now(),
		}, nil
	}

	statsCopy := *stats
	return &statsCopy, nil
}

// UpdateDocumentStats 更新文档统计
func (r *collaborationRepository) UpdateDocumentStats(_ context.Context, stats *domain.DocumentStats) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	statsCopy := *stats
	r.documentStats[stats.DocumentID] = &statsCopy
	return nil
}

// SaveMessage 保存消息（用于审计）
func (r *collaborationRepository) SaveMessage(ctx context.Context, msg *domain.CollaborationMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	msgCopy := *msg
	// 序列化Content以避免引用问题
	if msg.Content != nil {
		contentBytes, err := sonic.Marshal(msg.Content)
		if err != nil {
			return fmt.Errorf("[Infrastruct.Repository.CollaborationRepository] failed to marshal message content: %w", err)
		}
		var content interface{}
		if err := sonic.Unmarshal(contentBytes, &content); err != nil {
			return fmt.Errorf("[Infrastruct.Repository.CollaborationRepository] failed to marshal message content: %w", err)
		}
		msgCopy.Content = content
	}

	r.messages[msg.DocumentID] = append(r.messages[msg.DocumentID], &msgCopy)

	// 限制消息数量，保留最近1000条
	if len(r.messages[msg.DocumentID]) > 1000 {
		r.messages[msg.DocumentID] = r.messages[msg.DocumentID][len(r.messages[msg.DocumentID])-1000:]
	}

	// 更新文档统计
	r.updateDocumentStatsUnsafe(msg.DocumentID)
	return nil
}

// GetRecentMessages 获取最近的消息
func (r *collaborationRepository) GetRecentMessages(ctx context.Context, documentID string, limit int) ([]*domain.CollaborationMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	messages := r.messages[documentID]
	if len(messages) == 0 {
		return []*domain.CollaborationMessage{}, nil
	}

	start := 0
	if len(messages) > limit {
		start = len(messages) - limit
	}

	result := make([]*domain.CollaborationMessage, 0, limit)
	for i := start; i < len(messages); i++ {
		msgCopy := *messages[i]
		result = append(result, &msgCopy)
	}

	return result, nil
}

// updateDocumentStatsUnsafe 更新文档统计（不加锁版本）
func (r *collaborationRepository) updateDocumentStatsUnsafe(documentID string) {
	if r.documentStats[documentID] == nil {
		r.documentStats[documentID] = &domain.DocumentStats{
			DocumentID:   documentID,
			LastActivity: time.Now(),
		}
	}

	stats := r.documentStats[documentID]

	// 更新活跃用户数
	if users, exists := r.activeUsers[documentID]; exists {
		stats.ActiveUserCount = len(users)
	} else {
		stats.ActiveUserCount = 0
	}

	// 更新连接数
	connectionCount := 0
	for _, session := range r.userSessions {
		if session.DocumentID == documentID && session.IsActive {
			connectionCount++
		}
	}
	stats.TotalConnections = connectionCount

	// 更新消息数
	if messages, exists := r.messages[documentID]; exists {
		stats.MessageCount = int64(len(messages))
	}

	stats.LastActivity = time.Now()
}
