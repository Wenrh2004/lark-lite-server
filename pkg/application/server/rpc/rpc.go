package rpc

import (
	"github.com/cloudwego/kitex/server"

	"github.com/Wenrh2004/lark-lite-server/pkg/log"
)

type Server struct {
	server.Server
	logger *log.Logger
}

type Option func(s *Server)

func NewServer(server server.Server, logger *log.Logger, opts ...Option) *Server {
	s := &Server{
		Server: server,
		logger: logger,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Server) Start() {
	if err := s.Run(); err != nil {
		s.logger.Sugar().Fatalf("listen: %s\n", err)
	}
}

func (s *Server) Stop() {
	s.logger.Sugar().Info("Shutting down server...")
	if err := s.Server.Stop(); err != nil {
		s.logger.Sugar().Errorf("server stop error: %v", err)
	}
	s.logger.Sugar().Info("Server exiting")
}
