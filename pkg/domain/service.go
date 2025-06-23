package domain

import (
	"github.com/Wenrh2004/lark-lite-server/pkg/jwt"
	"github.com/Wenrh2004/lark-lite-server/pkg/log"
	"github.com/Wenrh2004/lark-lite-server/pkg/sid"
	"github.com/Wenrh2004/lark-lite-server/pkg/transaction"
)

type Service struct {
	Logger *log.Logger
	Sid    *sid.Sid
	Jwt    *jwt.JWT
	Tx     transaction.Transaction
}

func NewService(log *log.Logger, s *sid.Sid, j *jwt.JWT, tx transaction.Transaction) *Service {
	return &Service{
		Logger: log,
		Sid:    s,
		Jwt:    j,
		Tx:     tx,
	}
}
