package application

import (
	"fmt"

	kitexregistry "github.com/cloudwego/kitex/pkg/registry"
	kitexserver "github.com/cloudwego/kitex/server"
	"github.com/hertz-contrib/swagger"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"

	"github.com/Wenrh2004/lark-lite-server/common/kitex_gen/user/userservice"
	"github.com/Wenrh2004/lark-lite-server/internal/user/adapter"
	"github.com/Wenrh2004/lark-lite-server/pkg/application/server/http"
	"github.com/Wenrh2004/lark-lite-server/pkg/application/server/rpc"
	"github.com/Wenrh2004/lark-lite-server/pkg/log"
)

// NewUserHTTPApplication 创建用户应用程序
// @title           Lark Lite User Server API
// @version         1.0.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8888
// @BasePath  /api/v1
// @securityDefinitions.apiKey Bearer
// @in header
// @name Authorization
// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func NewUserHTTPApplication(conf *viper.Viper, logger *log.Logger, handler *adapter.UserHandler) *http.Server {
	h := http.NewServer(conf, logger)

	url := swagger.URL(fmt.Sprintf("http://localhost%s%s/swagger/doc.json", conf.GetString("app.addr"), conf.GetString("app.base_url"))) // The url pointing to API definition
	h.GET("/swagger/*any", swagger.WrapHandler(swaggerFiles.Handler, url))

	v1 := h.Group("/v1")

	userGroup := v1.Group("/user")

	// 不需要认证的路由
	userGroup.POST("/login", handler.Login)
	userGroup.POST("/register", handler.Register)

	// 需要认证的路由
	// userGroup.Use(handler.AuthMiddleware())
	// userGroup.GET("/info", handler.GetUserInfo)
	// userGroup.PUT("/info", handler.UpdateUserInfo)
	userGroup.POST("/logout", handler.Logout)

	// 刷新用户令牌
	userGroup.POST("/refresh", handler.Refresh)

	return h
}

func NewUserRPCApplication(conf *viper.Viper, logger *log.Logger, r kitexregistry.Registry, handler *adapter.UserServiceImpl) *rpc.Server {
	server := userservice.NewServer(handler, kitexserver.WithRegistry(r))
	s := rpc.NewServer(server, logger)

	return s
}
