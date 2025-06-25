package adapter

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"go.uber.org/zap"

	v1 "github.com/Wenrh2004/lark-lite-server/common/api/v1"
	"github.com/Wenrh2004/lark-lite-server/common/kitex_gen/user"
	"github.com/Wenrh2004/lark-lite-server/common/kitex_gen/user/userservice"
	"github.com/Wenrh2004/lark-lite-server/pkg/adapter"
)

type UserHandler struct {
	srv *adapter.Service
	cli userservice.Client
}

func NewUserHandler(cli userservice.Client) *UserHandler {
	return &UserHandler{
		cli: cli,
	}
}

func (h *UserHandler) Register(ctx context.Context, c *app.RequestContext) {
	var req v1.UserAuthRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] register has bind and validate filed", zap.Error(err))
		v1.HandlerError(c, v1.ErrBadRequest)
		return
	}

	resp, err := h.cli.Register(ctx, &user.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] register failed", zap.Error(err))
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}
	if resp.Resp.Code != 0 {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] register failed", zap.Any("resp", resp))
		v1.HandlerError(c, v1.Error{
			Code:    int(resp.Resp.Code),
			Message: resp.Resp.Message,
		})
	}
	c.SetCookie(
		"UserRefreshToken",
		resp.User.Token.RefreshToken,
		int(resp.User.Token.GetRefreshExpiresIn()),
		"/api/v1/user/refresh",
		"",
		protocol.CookieSameSiteStrictMode,
		true,
		true,
	)
	v1.HandlerSuccess(c, &v1.UserAuthResponseBody{
		UserId:    strconv.FormatInt(resp.User.UserId, 10),
		Username:  resp.User.Username,
		Nickname:  resp.User.Nickname,
		AvatarUrl: resp.User.AvatarUrl,
		Certificate: v1.UserCertificateResponseBody{
			Certificate: resp.User.Token.GetAccessToken(),
			ExpiresIn:   resp.User.Token.GetAccessExpiresIn(),
		},
	})
}

func (h *UserHandler) Login(ctx context.Context, c *app.RequestContext) {
	var req v1.UserAuthRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] register has bind and validate filed", zap.Error(err))
		v1.HandlerError(c, v1.ErrBadRequest)
		return
	}

	resp, err := h.cli.Login(ctx, &user.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] register failed", zap.Error(err))
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}
	if resp.Resp.Code != 0 {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] register failed", zap.Any("resp", resp))
		v1.HandlerError(c, v1.Error{
			Code:    int(resp.Resp.Code),
			Message: resp.Resp.Message,
		})
	}
	c.SetCookie(
		"UserRefreshToken",
		resp.User.Token.RefreshToken,
		int(resp.User.Token.RefreshExpiresIn),
		"/api/v1/user/refresh",
		"",
		protocol.CookieSameSiteStrictMode,
		true,
		true,
	)
	v1.HandlerSuccess(c, &v1.UserAuthResponseBody{
		UserId:    strconv.FormatInt(resp.User.UserId, 10),
		Username:  resp.User.Username,
		Nickname:  resp.User.Nickname,
		AvatarUrl: resp.User.AvatarUrl,
		Certificate: v1.UserCertificateResponseBody{
			Certificate: resp.User.Token.GetAccessToken(),
			ExpiresIn:   resp.User.Token.GetAccessExpiresIn(),
		},
	})
}

func (h *UserHandler) Refresh(ctx context.Context, c *app.RequestContext) {
	refreshToken := c.Request.Header.Cookie("UserrefreshToken")

	if len(refreshToken) == 0 {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] refresh token is empty")
		v1.HandlerError(c, v1.ErrUnauthorized)
		return
	}

	resp, err := h.cli.Refresh(ctx, &user.RefreshRequest{
		RefreshToken: string(refreshToken),
	})
	if err != nil {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] refresh failed", zap.Error(err))
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}
	if resp.Resp.Code != 0 {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] refresh failed", zap.Any("resp", resp))
		v1.HandlerError(c, v1.Error{
			Code:    int(resp.Resp.Code),
			Message: resp.Resp.Message,
		})
		return
	}
	c.SetCookie(
		"UserRefreshToken",
		resp.Token.RefreshToken,
		int(resp.Token.GetRefreshExpiresIn()),
		"/api/v1/user/refresh",
		"",
		protocol.CookieSameSiteStrictMode,
		true,
		true,
	)
	v1.HandlerSuccess(c, &v1.UserCertificateResponseBody{
		Certificate: resp.Token.GetAccessToken(),
		ExpiresIn:   resp.Token.GetAccessExpiresIn(),
	})
}

func (h *UserHandler) GetUserInfo(ctx context.Context, c *app.RequestContext) {
	userIDStr := c.GetString("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] invalid user_id", zap.Error(err))
		v1.HandlerError(c, v1.ErrBadRequest)
		return
	}

	resp, err := h.cli.GetUserInfo(ctx, &user.GetUserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] GetUserInfo failed", zap.Error(err))
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}
	if resp.Resp.Code != 0 {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] GetUserInfo business error", zap.Any("resp", resp))
		v1.HandlerError(c, v1.Error{
			Code:    int(resp.Resp.Code),
			Message: resp.Resp.Message,
		})
		return
	}

	v1.HandlerSuccess(c, &v1.UserAuthResponseBody{
		UserId:    strconv.FormatInt(resp.User.UserId, 10),
		Username:  resp.User.Username,
		Nickname:  resp.User.Nickname,
		AvatarUrl: resp.User.AvatarUrl,
	})
}

func (h *UserHandler) UpdateUser(ctx context.Context, c *app.RequestContext) {
	var req v1.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		v1.HandlerError(c, v1.ErrBadRequest)
		return
	}

	userIDStr := c.GetString("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] invalid user_id", zap.Error(err))
		v1.HandlerError(c, v1.ErrBadRequest)
		return
	}

	resp, err := h.cli.UpdateUser(ctx, &user.UpdateUserRequest{
		UserId:    userID,
		Nickname:  req.Nickname,
		AvatarUrl: req.AvatarUrl,
	})
	if err != nil {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] UpdateUser failed", zap.Error(err))
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}
	if resp.Resp.Code != 0 {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] UpdateUser business error", zap.Any("resp", resp))
		v1.HandlerError(c, v1.Error{
			Code:    int(resp.Resp.Code),
			Message: resp.Resp.Message,
		})
		return
	}

	v1.HandlerSuccess(c, &v1.UserAuthResponseBody{
		UserId:    strconv.FormatInt(resp.User.UserId, 10),
		Username:  resp.User.Username,
		Nickname:  resp.User.Nickname,
		AvatarUrl: resp.User.AvatarUrl,
	})
}

