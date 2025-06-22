package adapter

import (
	"github.com/Wenrh2004/lark-lite-server/pkg/log"
)

type Service struct {
	*log.Logger
}

func NewService(logger *log.Logger) *Service {
	return &Service{
		Logger: logger,
	}
}
