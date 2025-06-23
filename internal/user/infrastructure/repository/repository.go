package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/Wenrh2004/lark-lite-server/internal/user/infrastructure/repository/query"
	"github.com/Wenrh2004/lark-lite-server/pkg/log"
	"github.com/Wenrh2004/lark-lite-server/pkg/transaction"
)

// ctxTxKeyType is an unexported type to avoid collisions in context keys.
type ctxTxKeyType string

// ctxTxKey is the key for storing transaction in context.
const ctxTxKey ctxTxKeyType = "TxKey"

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

func NewTransaction(r *Repository) transaction.Transaction {
	return r
}

func (r *Repository) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return query.Q.Transaction(func(tx *query.Query) error {
		ctx = context.WithValue(ctx, ctxTxKey, tx)
		return fn(ctx)
	})
}
