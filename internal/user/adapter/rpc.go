package adapter

import (
	"context"

	"github.com/Wenrh2004/lark-lite-server/internal/user/domain"
	"github.com/Wenrh2004/lark-lite-server/kitex_gen/common"
	"github.com/Wenrh2004/lark-lite-server/kitex_gen/user"
	"github.com/Wenrh2004/lark-lite-server/pkg/adapter"
)

type UserServiceImpl struct {
	srv         *adapter.Service
	userService domain.UserService
}

func (u *UserServiceImpl) Register(ctx context.Context, req *user.RegisterRequest) (res *user.UserAuthInfoResponse, err error) {
	newUser := domain.NewUser(req.Username, req.Password)
	ur, err := u.userService.Register(ctx, newUser)
	if err != nil {
		return nil, err
	}
	res = &user.UserAuthInfoResponse{
		Resp: &common.BaseResponse{
			Code:    0,
			Message: "success",
		},
		User: &user.UserAuthInfo{
			UserId:    int64(ur.ID),
			Username:  string(ur.Username),
			Nickname:  string(ur.Nickname),
			AvatarUrl: ur.AvatarURL,
			Token: &user.TokenPair{
				AccessToken:      ur.TokenPair.AccessToken.Token,
				AccessExpiresIn:  ur.TokenPair.AccessToken.ExpiresIn,
				RefreshToken:     ur.TokenPair.RefreshToken.Token,
				RefreshExpiresIn: ur.TokenPair.RefreshToken.ExpiresIn,
			},
		},
	}
	return res, nil
}

func (u *UserServiceImpl) Login(ctx context.Context, req *user.LoginRequest) (res *user.UserAuthInfoResponse, err error) {
	newUser := domain.NewUser(req.Username, req.Password)
	ur, err := u.userService.Login(ctx, newUser)
	if err != nil {
		return nil, err
	}
	res = &user.UserAuthInfoResponse{
		Resp: &common.BaseResponse{
			Code:    0,
			Message: "success",
		},
		User: &user.UserAuthInfo{
			UserId:    int64(ur.ID),
			Username:  string(ur.Username),
			Nickname:  string(ur.Nickname),
			AvatarUrl: ur.AvatarURL,
			Token: &user.TokenPair{
				AccessToken:      ur.TokenPair.AccessToken.Token,
				AccessExpiresIn:  ur.TokenPair.AccessToken.ExpiresIn,
				RefreshToken:     ur.TokenPair.RefreshToken.Token,
				RefreshExpiresIn: ur.TokenPair.RefreshToken.ExpiresIn,
			},
		},
	}
	return res, nil
}

func (u *UserServiceImpl) Refresh(ctx context.Context, req *user.RefreshRequest) (res *user.AuthResponse, err error) {
	// TODO implement me
	panic("implement me")
}

func (u *UserServiceImpl) Update(ctx context.Context, req *user.UpdateRequest) (res *common.BaseResponse, err error) {
	updateUser := &domain.User{
		ID:            uint64(req.UserId),
		Nickname:      domain.Username(req.Nickname),
		AvatarURL:     req.AvatarUrl,
		Email:         req.Email,
		BackgroundURL: req.BackgroundUrl,
		Signature:     req.Signature,
		Phone:         req.Phone,
		Gender:        domain.Gender(req.Gender),
	}
	_, err = u.userService.GetUserByID(ctx, uint64(req.UserId))
	if err != nil {
		return &common.BaseResponse{Code: 1, Message: "user not found"}, nil
	}
	_, err = u.userService.UpdateUser(ctx, updateUser)
	if err != nil {
		return &common.BaseResponse{Code: 3, Message: err.Error()}, nil
	}
	return &common.BaseResponse{Code: 0, Message: "success"}, nil
}

func (u *UserServiceImpl) GetUserInfo(ctx context.Context, req *user.GetUserInfoRequest) (res *user.GetUserInfoResponse, err error) {
	ur, err := u.userService.GetUserByID(ctx, uint64(req.UserId))
	if err != nil || ur == nil {
		return &user.GetUserInfoResponse{
			Resp: &common.BaseResponse{Code: 1, Message: "user not found"},
		}, nil
	}
	return &user.GetUserInfoResponse{
		Resp: &common.BaseResponse{Code: 0, Message: "success"},
		User: &user.UserAuthInfo{
			UserId:    int64(ur.ID),
			Username:  ur.Username.String(),
			Nickname:  ur.Nickname.String(),
			AvatarUrl: ur.AvatarURL,
		},
	}, nil
}

func NewUserServiceImpl(srv *adapter.Service, userService domain.UserService) *UserServiceImpl {
	return &UserServiceImpl{
		srv:         srv,
		userService: userService,
	}
}
