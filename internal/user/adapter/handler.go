package adapter

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"go.uber.org/zap"

	v1 "github.com/Wenrh2004/lark-lite-server/common/api/v1"
	"github.com/Wenrh2004/lark-lite-server/internal/user/domain"
	"github.com/Wenrh2004/lark-lite-server/pkg/adapter"
)

const (
	EmptyToken           = ""
	UserRefreshTokenName = "UserRefreshToken"
	UserRefreshTokenPath = "/api/v1/user/refresh"
)

type UserHandler struct {
	srv *adapter.Service
	us  domain.UserService
}

func NewUserHandler(srv *adapter.Service, cli domain.UserService) *UserHandler {
	return &UserHandler{
		srv: srv,
		us:  cli,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body v1.UserAuthRequest true "用户注册请求"
// @Success 200 {object} v1.UserAuthResponse "注册成功"
// @Failure 400 {object} v1.ErrorResponse "请求参数错误"
// @Failure 500 {object} v1.ErrorResponse "服务器内部错误"
// @Router /api/v1/user/register [post]
func (h *UserHandler) Register(ctx context.Context, c *app.RequestContext) {
	var req v1.UserAuthRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] register has bind and validate filed", zap.Error(err))
		v1.HandlerError(c, v1.ErrBadRequest)
		return
	}

	user := domain.NewUser(req.Username, req.Password)
	resp, err := h.us.Register(ctx, user)
	if err != nil {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] register failed", zap.Error(err))
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}

	c.SetCookie(
		"UserRefreshToken",
		resp.TokenPair.RefreshToken.Token,
		int(resp.TokenPair.RefreshToken.ExpiresIn),
		"/api/v1/user/refresh",
		"",
		protocol.CookieSameSiteStrictMode,
		true,
		true,
	)
	v1.HandlerSuccess(c, &v1.UserAuthResponseBody{
		UserId:    strconv.FormatUint(resp.ID, 10),
		Username:  resp.Username.String(),
		Nickname:  resp.Nickname.String(),
		AvatarUrl: resp.AvatarURL,
		Certificate: v1.UserCertificateResponseBody{
			Certificate: resp.TokenPair.AccessToken.Token,
			ExpiresIn:   resp.TokenPair.AccessToken.ExpiresIn,
		},
	})
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户账户登录认证
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body v1.UserAuthRequest true "用户登录请求"
// @Success 200 {object} v1.UserAuthResponse "登录成功"
// @Failure 400 {object} v1.ErrorResponse "请求参数错误"
// @Failure 401 {object} v1.ErrorResponse "认证失败"
// @Failure 500 {object} v1.ErrorResponse "服务器内部错误"
// @Router /api/v1/user/login [post]
func (h *UserHandler) Login(ctx context.Context, c *app.RequestContext) {
	var req v1.UserAuthRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] login has bind and validate filed", zap.Error(err))
		v1.HandlerError(c, v1.ErrBadRequest)
		return
	}

	user := domain.NewUser(req.Username, req.Password)
	resp, err := h.us.Login(ctx, user)
	if err != nil {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] login failed", zap.Error(err))
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}

	c.SetCookie(
		UserRefreshTokenName,
		resp.TokenPair.RefreshToken.Token,
		int(resp.TokenPair.RefreshToken.ExpiresIn),
		UserRefreshTokenPath,
		"",
		protocol.CookieSameSiteStrictMode,
		true,
		true,
	)
	v1.HandlerSuccess(c, &v1.UserAuthResponseBody{
		UserId:    strconv.FormatUint(resp.ID, 10),
		Username:  resp.Username.String(),
		Nickname:  resp.Nickname.String(),
		AvatarUrl: resp.AvatarURL,
		Certificate: v1.UserCertificateResponseBody{
			Certificate: resp.TokenPair.AccessToken.Token,
			ExpiresIn:   resp.TokenPair.AccessToken.ExpiresIn,
		},
	})
}

// Refresh 刷新令牌
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param UserRefreshToken header string false "刷新令牌"
// @Success 200 {object} v1.RefreshResponse "刷新成功"
// @Failure 401 {object} v1.ErrorResponse "认证失败或令牌过期"
// @Failure 500 {object} v1.ErrorResponse "服务器内部错误"
// @Router /api/v1/user/refresh [post]
func (h *UserHandler) Refresh(ctx context.Context, c *app.RequestContext) {
	refreshToken := c.Request.Header.Cookie("UserRefreshToken")

	if len(refreshToken) == 0 {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] refresh token is empty")
		v1.HandlerError(c, v1.ErrUnauthorized)
		return
	}

	resp, err := h.us.Refresh(ctx, string(refreshToken))
	if err != nil {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] refresh failed", zap.Error(err))
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}

	c.SetCookie(
		UserRefreshTokenName,
		resp.RefreshToken.Token,
		int(resp.RefreshToken.ExpiresIn),
		UserRefreshTokenPath,
		"",
		protocol.CookieSameSiteStrictMode,
		true,
		true,
	)
	v1.HandlerSuccess(c, &v1.UserCertificateResponseBody{
		Certificate: resp.AccessToken.Token,
		ExpiresIn:   resp.AccessToken.ExpiresIn,
	})
}

// Logout 用户登出
// @Summary 用户登出
// @Description 清除用户的登录状态
// @Tags 用户认证
// @Accept json
// @Produce json
// @Success 200 {object} v1.SuccessResponse "登出成功"
// @Failure 500 {object} v1.ErrorResponse "服务器内部错误"
// @Router /api/v1/user/logout [post]
func (h *UserHandler) Logout(ctx context.Context, c *app.RequestContext) {
	c.SetCookie(
		UserRefreshTokenName,
		EmptyToken,
		0,
		UserRefreshTokenPath,
		"",
		protocol.CookieSameSiteStrictMode,
		true,
		true,
	)
	v1.HandlerSuccess(c, nil)
}
