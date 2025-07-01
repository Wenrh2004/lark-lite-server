package application

import (
	"github.com/spf13/viper"

	"github.com/Wenrh2004/lark-lite-server/internal/collaboration/adapter"
	"github.com/Wenrh2004/lark-lite-server/pkg/application/server/http"
	"github.com/Wenrh2004/lark-lite-server/pkg/log"
)

func NewUserHTTPApplication(conf *viper.Viper, logger *log.Logger, handler *adapter.WebSocketAdapter) *http.Server {
	h := http.NewServer(conf, logger)

	v1 := h.Group("/v1")

	collaGroup := v1.Group("/collaboration")

	collaGroup.GET("/:doc_id", handler.HandleWebSocketUpgrade)
	collaGroup.GET("/status/:doc_id", handler.GetDocumentStats)
	collaGroup.GET("/users/:doc_id", handler.GetActiveUsers)
	collaGroup.GET("/conn/:doc_id", handler.GetDocumentConnections)
	collaGroup.GET("/conn", handler.GetConnectionStats)

	return h
}
