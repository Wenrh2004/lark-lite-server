package application

import (
	"github.com/spf13/viper"

	"github.com/Wenrh2004/lark-lite-server/common/kitex_gen/user/userservice"
	"github.com/Wenrh2004/lark-lite-server/internal/user/adapter"
	"github.com/Wenrh2004/lark-lite-server/pkg/application/server/http"
	"github.com/Wenrh2004/lark-lite-server/pkg/application/server/rpc"
	"github.com/Wenrh2004/lark-lite-server/pkg/log"
)

// NewUserHTTPApplication 创建用户应用程序
// @title           Brain Hub API
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

// @Summary 用户注册
// @Description 用户注册接口
// @Tags 用户
// @Accept json
// @Produce json
// @Param data body UserAuthRequest true "注册参数"
// @Success 200 {object} UserAuthResponseBody
// @Router /v1/user/register [post]

// @Summary 用户登录
// @Description 用户登录接口
// @Tags 用户
// @Accept json
// @Produce json
// @Param data body UserAuthRequest true "登录参数"
// @Success 200 {object} UserAuthResponseBody
// @Router /v1/user/login [post]
// @Summary 获取用户信息
// @Description 获取当前登录用户的信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} UserAuthResponseBody
// @Router /v1/user/info [get]

// @Summary 更新用户信息
// @Description 更新当前登录用户的信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param data body UpdateUserRequest true "更新参数"
// @Success 200 {object} UserAuthResponseBody
// @Router /v1/user/update [put]
func NewUserHTTPApplication(conf *viper.Viper, logger *log.Logger, handler *adapter.UserHandler) *http.Server {
	h := http.NewServer(conf, logger)

	v1 := h.Group("/v1")

	userGroup := v1.Group("/user")

	// 不需要认证的路由
	userGroup.POST("/login", handler.Login)
	userGroup.POST("/register", handler.Register)

	// 需要认证的路由
    userGroup.GET("/info", handler.GetUserInfo)
    userGroup.PUT("/update", handler.UpdateUser)
	return h
}

func NewUserRPCApplication(conf *viper.Viper, logger *log.Logger, handler *adapter.UserServiceImpl) *rpc.Server {
	server := userservice.NewServer(handler)
	s := rpc.NewServer(server, logger)

	return s
}
