package domain

import (
	"context"
	"fmt"

	"github.com/Wenrh2004/lark-lite-server/pkg/domain"
)

type UserService interface {
	Login(ctx context.Context, user *User) (*User, error)
	Register(ctx context.Context, user *User) (*User, error)
	GetUserByID(ctx context.Context, id uint64) (*User, error)
	GetUser(ctx context.Context, user *User) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
}

type userService struct {
	srv  *domain.Service
	repo UserRepository
}

func (u *userService) Login(ctx context.Context, user *User) (*User, error) {
	res, err := u.repo.GetUserByID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("[Domain.Service.User] get user by id: %w", err)
	}
	ack, rfk, err := u.srv.Jwt.GenTokenPair(res.ID)
	if err != nil {
		return nil, fmt.Errorf("[Domain.Service.User] generate token pair: %w", err)
	}
	res.TokenPair = NewCertificatePair(ack, rfk, u.srv.Jwt.GetAckExpires(), u.srv.Jwt.GetRefreshExpires())
	return res, nil
}

func (u *userService) Register(ctx context.Context, user *User) (*User, error) {
	uid, err := u.srv.Sid.GenUint64()
	if err != nil {
		return nil, fmt.Errorf("[Domain.Service.User] gen sid field: %w", err)
	}
	user.ID = uid
	res, err := u.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	akc, rfk, err := u.srv.Jwt.GenTokenPair(res.ID)
	if err != nil {
		return nil, err
	}
	res.TokenPair = NewCertificatePair(akc, rfk, u.srv.Jwt.GetAckExpires(), u.srv.Jwt.GetRefreshExpires())
	return user, nil
}

func (u *userService) GetUserByID(ctx context.Context, id uint64) (*User, error) {
	res, err := u.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("[Domain.Service.User] get user by id: %w", err)
	}
	return res, nil
}

func (u *userService) GetUser(ctx context.Context, user *User) (*User, error) {
	// TODO implement me
	panic("implement me")
}

func (u *userService) UpdateUser(ctx context.Context, user *User) (*User, error) {
	_, err := u.repo.GetUserByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	updated, err := u.repo.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func NewUserService(srv *domain.Service, repo UserRepository) UserService {
	return &userService{
		srv:  srv,
		repo: repo,
	}
}
