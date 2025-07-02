package http

import (
	"errors"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/spf13/viper"

	"github.com/Wenrh2004/lark-lite-server/pkg/log"
)

type Server struct {
	*server.Hertz
	logger *log.Logger
}

type Option func(s *Server)

func NewServer(conf *viper.Viper, logger *log.Logger, opts ...Option) *Server {
	h := server.Default(
		server.WithHostPorts(conf.GetString("app.addr")),
		server.WithBasePath(conf.GetString("app.base_url")),
	)

	s := &Server{
		Hertz:  h,
		logger: logger,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (h *Server) Start() {
	if err := h.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		h.logger.Sugar().Fatalf("listen: %s\n", err)
	}
}

func (h *Server) Stop() {
	h.logger.Sugar().Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	h.Spin()

	h.logger.Sugar().Info("Server exiting")
}
