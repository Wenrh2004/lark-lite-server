package infrastructure

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/Wenrh2004/lark-lite-server/internal/collaboration/domain"
	"github.com/Wenrh2004/lark-lite-server/pkg/log"
)

// ConnectionManager 连接管理器 - 负责基础设施关注点
type ConnectionManager struct {
	logger        *log.Logger
	domainService domain.ConnectionService
	broadcaster   *WebSocketBroadcaster
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager(
	domainService domain.ConnectionService,
	broadcaster *WebSocketBroadcaster,
) *ConnectionManager {
	manager := &ConnectionManager{
		domainService: domainService,
		broadcaster:   broadcaster,
		stopCleanup:   make(chan struct{}),
	}

	// 启动清理定时任务
	manager.startCleanupRoutine()

	return manager
}

// GetBroadcaster 获取广播器
func (m *ConnectionManager) GetBroadcaster() *WebSocketBroadcaster {
	return m.broadcaster
}

// GetDomainService 获取领域服务
func (m *ConnectionManager) GetDomainService() domain.ConnectionService {
	return m.domainService
}

// startCleanupRoutine 启动清理定时任务
func (m *ConnectionManager) startCleanupRoutine() {
	m.cleanupTicker = time.NewTicker(time.Minute)

	go func() {
		for {
			select {
			case <-m.cleanupTicker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				cleaned, err := m.domainService.CleanupTimeoutConnections(ctx, 5*time.Minute)
				if err != nil {
					m.logger.Error("[Infrastructure.ConnectionManager] failed to cleanup connections", zap.Error(err))
				} else if cleaned > 0 {
					m.logger.Info("[Infrastructure.ConnectionManager] cleaned up connections", zap.Int("count", cleaned))
				} else {
					m.logger.Debug("[Infrastructure.ConnectionManager] no connections to cleanup")
				}
				cancel()

			case <-m.stopCleanup:
				m.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// Shutdown 关闭管理器
func (m *ConnectionManager) Shutdown() {
	close(m.stopCleanup)
	if m.cleanupTicker != nil {
		m.cleanupTicker.Stop()
	}
}
