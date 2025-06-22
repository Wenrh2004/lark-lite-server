package repository

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/Wenrh2004/lark-lite-server/internal/user/infrastructure/repository/query"
	"github.com/Wenrh2004/lark-lite-server/pkg/log"
)

const ctxTxKey = "TxKey"

type Repository struct {
	query  *query.Query
	rdb    *redis.Client
	logger *log.Logger
}

func NewRepository(
	logger *log.Logger,
	db *gorm.DB,
	rdb *redis.Client,
) *Repository {
	query.SetDefault(db)
	return &Repository{
		query:  query.Q,
		rdb:    rdb,
		logger: logger,
	}
}
