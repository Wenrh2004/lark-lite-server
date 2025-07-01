package adapter

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"go.uber.org/zap"

	v1 "github.com/Wenrh2004/lark-lite-server/common/api/v1"
	"github.com/Wenrh2004/lark-lite-server/kitex_gen/user"
	"github.com/Wenrh2004/lark-lite-server/kitex_gen/user/userservice"
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
	if resp == nil || resp.Resp == nil || resp.Resp.Code != 0 {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] register failed", zap.Any("resp", resp))
		msg := "internal error"
		code := 500
		if resp != nil && resp.Resp != nil {
			msg = resp.Resp.Message
			code = int(resp.Resp.Code)
		}
		v1.HandlerError(c, v1.Error{
			Code:    code,
			Message: msg,
		})
		return
	}
	if resp.User == nil || resp.User.Token == nil {
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
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
			Certificate: resp.User.Token.AccessToken,
			ExpiresIn:   resp.User.Token.AccessExpiresIn,
		},
	})
}

func (h *UserHandler) Login(ctx context.Context, c *app.RequestContext) {
	var req v1.UserAuthRequest
	if err := c.BindAndValidate(&req); err != nil {
		v1.HandlerError(c, v1.ErrBadRequest)
		return
	}

	resp, err := h.cli.Login(ctx, &user.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] login failed", zap.Error(err))
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}
	if resp == nil || resp.Resp == nil || resp.Resp.Code != 0 {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] login failed", zap.Any("resp", resp))
		msg := "internal error"
		code := 500
		if resp != nil && resp.Resp != nil {
			msg = resp.Resp.Message
			code = int(resp.Resp.Code)
		}
		v1.HandlerError(c, v1.Error{
			Code:    code,
			Message: msg,
		})
		return
	}
	if resp.User == nil || resp.User.Token == nil {
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
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
			Certificate: resp.User.Token.AccessToken,
			ExpiresIn:   resp.User.Token.AccessExpiresIn,
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
	if resp == nil || resp.Resp == nil || resp.Resp.Code != 0 {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] refresh failed", zap.Any("resp", resp))
		msg := "internal error"
		code := 500
		if resp != nil && resp.Resp != nil {
			msg = resp.Resp.Message
			code = int(resp.Resp.Code)
		}
		v1.HandlerError(c, v1.Error{
			Code:    code,
			Message: msg,
		})
		return
	}
	if resp.Token == nil {
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}
	c.SetCookie(
		"UserRefreshToken",
		resp.Token.RefreshToken,
		int(resp.Token.RefreshExpiresIn),
		"/api/v1/user/refresh",
		"",
		protocol.CookieSameSiteStrictMode,
		true,
		true,
	)
	v1.HandlerSuccess(c, &v1.UserCertificateResponseBody{
		Certificate: resp.Token.AccessToken,
		ExpiresIn:   resp.Token.AccessExpiresIn,
	})
}

type UpdateUserRequest struct {
	Nickname  string `json:"nickname"`
	AvatarUrl string `json:"avatar_url"`
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
	if resp == nil || resp.GetResp() == nil || resp.GetResp().Code != 0 {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] GetUserInfo business error", zap.Any("resp", resp))
		msg := "internal error"
		code := 500
		if resp != nil && resp.GetResp() != nil {
			msg = resp.GetResp().Message
			code = int(resp.GetResp().Code)
		}
		v1.HandlerError(c, v1.Error{
			Code:    code,
			Message: msg,
		})
		return
	}
	if resp.GetUser() == nil {
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}
	v1.HandlerSuccess(c, &v1.UserAuthResponseBody{
		UserId:    strconv.FormatInt(resp.GetUser().GetUserId(), 10),
		Username:  resp.GetUser().GetUsername(),
		Nickname:  resp.GetUser().GetNickname(),
		AvatarUrl: resp.GetUser().GetAvatarUrl(),
	})
}

func (h *UserHandler) UpdateUser(ctx context.Context, c *app.RequestContext) {
	var req UpdateUserRequest
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

	resp, err := h.cli.Update(ctx, &user.UpdateRequest{
		UserId:    userID,
		Nickname:  req.Nickname,
		AvatarUrl: req.AvatarUrl,
	})
	if err != nil {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] UpdateUser failed", zap.Error(err))
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}
	if resp == nil || resp.Code != 0 {
		h.srv.Logger.WithContext(ctx).Error("[Adapter.User] UpdateUser business error", zap.Any("resp", resp))
		msg := "internal error"
		code := 500
		if resp != nil {
			msg = resp.Message
			code = int(resp.Code)
		}
		v1.HandlerError(c, v1.Error{
			Code:    code,
			Message: msg,
		})
		return
	}
	// 更新后再查一次用户信息
	userInfoResp, err := h.cli.GetUserInfo(ctx, &user.GetUserInfoRequest{
		UserId: userID,
	})
	if err != nil || userInfoResp == nil || userInfoResp.GetResp() == nil || userInfoResp.GetResp().Code != 0 || userInfoResp.GetUser() == nil {
		v1.HandlerError(c, v1.ErrInternalServerError)
		return
	}
	v1.HandlerSuccess(c, &v1.UserAuthResponseBody{
		UserId:    strconv.FormatInt(userInfoResp.GetUser().GetUserId(), 10),
		Username:  userInfoResp.GetUser().GetUsername(),
		Nickname:  userInfoResp.GetUser().GetNickname(),
		AvatarUrl: userInfoResp.GetUser().GetAvatarUrl(),
	})
}

// UploadFile 文件上传接口
func (h *UserHandler) UploadFile(ctx context.Context, c *app.RequestContext) {
	FileUploadHandler(ctx, c)
}
