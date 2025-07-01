package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Wenrh2004/lark-lite-server/internal/collaboration/domain"
)

// ConnectionRepository 内存连接仓储实现
type ConnectionRepository struct {
	connections map[string]*domain.Connection
	stats       *domain.ConnectionStats
	mu          sync.RWMutex
}

func NewConnectionRepository() *ConnectionRepository {
	return &ConnectionRepository{
		connections: make(map[string]*domain.Connection),
		stats: &domain.ConnectionStats{
			TotalConnections: 0,
			MessagesSent:     0,
			MessagesReceived: 0,
			ShardCount:       1,
		},
	}
}

// Add 添加连接
func (r *ConnectionRepository) Add(_ context.Context, conn *domain.Connection) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.connections[conn.ID]; exists {
		return fmt.Errorf("[Infrastructure.Repository.ConnectionRepository] connection %s already exists", conn.ID)
	}

	connCopy := *conn
	r.connections[conn.ID] = &connCopy
	r.stats.TotalConnections++

	return nil
}

// Remove 移除连接
func (r *ConnectionRepository) Remove(ctx context.Context, connectionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.connections[connectionID]; !exists {
		return fmt.Errorf("[Infrastructure.Repository.ConnectionRepository] connection %s not found", connectionID)
	}

	delete(r.connections, connectionID)
	if r.stats.TotalConnections > 0 {
		r.stats.TotalConnections--
	}

	return nil
}

// GetByDocID 根据文档ID获取连接
func (r *ConnectionRepository) GetByDocID(ctx context.Context, docID string) ([]*domain.Connection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var connections []*domain.Connection
	for _, conn := range r.connections {
		if conn.DocID == docID {
			connCopy := *conn
			connections = append(connections, &connCopy)
		}
	}

	return connections, nil
}

// GetByID 根据连接ID获取连接
func (r *ConnectionRepository) GetByID(ctx context.Context, connectionID string) (*domain.Connection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conn, exists := r.connections[connectionID]
	if !exists {
		return nil, fmt.Errorf("[Infrastructure.Repository.ConnectionRepository] connection %s not found", connectionID)
	}

	connCopy := *conn
	return &connCopy, nil
}

// UpdateLastPing 更新最后ping时间
func (r *ConnectionRepository) UpdateLastPing(ctx context.Context, connectionID string, pingTime time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, exists := r.connections[connectionID]
	if !exists {
		return fmt.Errorf("[Infrastructure.Repository.ConnectionRepository] connection %s not found", connectionID)
	}

	conn.LastPing = pingTime
	return nil
}

// GetStats 获取统计信息
func (r *ConnectionRepository) GetStats(ctx context.Context) (*domain.ConnectionStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	statsCopy := *r.stats
	return &statsCopy, nil
}

// CleanupTimeout 清理超时连接
func (r *ConnectionRepository) CleanupTimeout(ctx context.Context, timeout time.Duration) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	var cleanedCount int
	var toDelete []string

	for id, conn := range r.connections {
		if now.Sub(conn.LastPing) > timeout {
			toDelete = append(toDelete, id)
		}
	}

	for _, id := range toDelete {
		delete(r.connections, id)
		cleanedCount++
		if r.stats.TotalConnections > 0 {
			r.stats.TotalConnections--
		}
	}

	return cleanedCount, nil
}

func (r *ConnectionRepository) IncrementMessagesSent() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.MessagesSent++
}

func (r *ConnectionRepository) IncrementMessagesReceived() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.MessagesReceived++
}
