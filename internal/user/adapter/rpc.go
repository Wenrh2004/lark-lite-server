package adapter

import (
	"context"

	"github.com/Wenrh2004/lark-lite-server/common/kitex_gen/common"
	"github.com/Wenrh2004/lark-lite-server/common/kitex_gen/user"
	"github.com/Wenrh2004/lark-lite-server/internal/user/domain/service"
)

type UserServiceImpl struct {
	service service.UserService
}

func (u *UserServiceImpl) Register(ctx context.Context, req *user.RegisterRequest) (res *user.UserAuthInfoResponse, err error) {
	newUser := service.NewUser(req.Username, req.Password)
	ur, err := u.service.Register(ctx, newUser)
	if err != nil {
		return nil, err
	}
	res.Resp = &common.BaseResponse{
		Code:    0,
		Message: "success",
	}
	res.User = &user.UserAuthInfo{
		UserId:    int64(ur.ID),
		Username:  string(ur.Username),
		Nickname:  string(ur.Nickname),
		AvatarUrl: ur.AvatarURL,
		Token: &user.TokenPair{
			AccessToken:      ur.AccessToken,
			AccessExpiresIn:  ur.AccessExpiresIn,
			RefreshToken:     ur.RefreshToken,
			RefreshExpiresIn: ur.RefreshExpiresIn,
		},
	}
	return res, nil
}

func (u *UserServiceImpl) Login(ctx context.Context, req *user.LoginRequest) (res *user.UserAuthInfoResponse, err error) {
	newUser := service.NewUser(req.Username, req.Password)
	ur, err := u.service.Login(ctx, newUser)
	if err != nil {
		return nil, err
	}
	res.Resp = &common.BaseResponse{
		Code:    0,
		Message: "success",
	}
	res.User = &user.UserAuthInfo{
		UserId:    int64(ur.ID),
		Username:  string(ur.Username),
		Nickname:  string(ur.Nickname),
		AvatarUrl: ur.AvatarURL,
		Token: &user.TokenPair{
			AccessToken:      ur.AccessToken,
			AccessExpiresIn:  ur.AccessExpiresIn,
			RefreshToken:     ur.RefreshToken,
			RefreshExpiresIn: ur.RefreshExpiresIn,
		},
	}
	return res, nil
}

func (u *UserServiceImpl) Refresh(ctx context.Context, req *user.RefreshRequest) (res *user.AuthResponse, err error) {
	// TODO implement me
	panic("implement me")
}

func (u *UserServiceImpl) Update(ctx context.Context, req *user.UpdateRequest) (res *common.BaseResponse, err error) {
	// TODO implement me
	panic("implement me")
}

func (u *UserServiceImpl) GetUserInfo(ctx context.Context, req *user.GetUserInfoRequest) (res *user.GetUserInfoResponse, err error) {
	// TODO implement me
	panic("implement me")
}
