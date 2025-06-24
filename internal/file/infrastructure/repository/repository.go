package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/Wenrh2004/lark-lite-server/internal/file/infrastructure/repository/query"
	"github.com/Wenrh2004/lark-lite-server/internal/file/infrastructure/third/oss"
	"github.com/Wenrh2004/lark-lite-server/pkg/log"
	"github.com/Wenrh2004/lark-lite-server/pkg/transaction"
)

const ctxTxKey = "TxKey"

type Repository struct {
	rdb    *redis.Client
	oss    oss.Service
	logger *log.Logger
}

func NewRepository(
	logger *log.Logger,
	db *gorm.DB,
	rdb *redis.Client,
	oss oss.Service,
) *Repository {
	query.SetDefault(db)
	return &Repository{
		rdb:    rdb,
		oss:    oss,
		logger: logger,
	}
}

func NewTransaction(r *Repository) transaction.Transaction {
	return r
}

func (r *Repository) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return query.Q.Transaction(func(tx *query.Query) error {
		ctx = context.WithValue(ctx, ctxTxKey, tx)
		return fn(ctx)
	})
}
