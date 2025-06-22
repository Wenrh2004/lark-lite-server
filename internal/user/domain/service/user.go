package service

import (
	"context"
	"fmt"

	"github.com/Wenrh2004/lark-lite-server/common/kitex_gen/user"
	"github.com/Wenrh2004/lark-lite-server/pkg/domain"
)

type Gender int

const (
	GenderUnknown Gender = iota
	GenderMale
	GenderFemale
)

type Username string

func NewUsername(name string) Username {
	var username Username
	username = Username(name)
	return username
}

func (u Username) String() string {
	return string(u)
}

type Password string

func NewPassword(password string) Password {
	var pwd Password
	// TODO: add the password encryption logic here
	pwd = Password(password)
	return pwd
}

func (p Password) String() string {
	return string(p)
}

// User 实体
type User struct {
	ID            uint64
	Username      Username
	Password      Password
	Nickname      Username
	AvatarURL     string
	BackgroundURL string
	Signature     string
	Email         string
	Phone         string
	Gender        Gender
	user.TokenPair
}

func NewUser(username, password string) *User {
	return &User{
		Username: NewUsername(username),
		Password: NewPassword(password),
		Nickname: NewUsername(username),
		Gender:   GenderUnknown,
	}
}

type UserService interface {
	Login(ctx context.Context, user *User) (*User, error)
	Register(ctx context.Context, user *User) (*User, error)
	GetUserByID(ctx context.Context, id uint64) (*User, error)
	GetUser(ctx context.Context, user *User) (*User, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
	GetUserByID(ctx context.Context, id uint64) (*User, error)
	GetUser(ctx context.Context, user *User) (*User, error)
}

type userService struct {
	srv  *domain.Service
	repo UserRepository
}

func (u *userService) Login(ctx context.Context, user *User) (*User, error) {
	res, err := u.repo.GetUserByID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("[Domain.User] get user by id: %w", err)
	}
	ack, rfk, err := u.srv.Jwt.GenTokenPair(res.ID)
	if err != nil {
		return nil, fmt.Errorf("[Domain.User] generate token pair: %w", err)
	}
	res.TokenPair.AccessToken = ack
	res.TokenPair.RefreshToken = rfk
	res.TokenPair.AccessExpiresIn = u.srv.Jwt.GetAckExpires()
	res.TokenPair.RefreshExpiresIn = u.srv.Jwt.GetRefreshExpires()
	return res, nil
}

func (u *userService) Register(ctx context.Context, user *User) (*User, error) {
	uid, err := u.srv.Sid.GenUint64()
	if err != nil {
		return nil, fmt.Errorf("[Domain.User] gen sid field: %w", err)
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
	user.TokenPair.AccessToken = akc
	user.TokenPair.RefreshToken = rfk
	user.TokenPair.AccessExpiresIn = u.srv.Jwt.GetAckExpires()
	user.TokenPair.RefreshExpiresIn = u.srv.Jwt.GetRefreshExpires()
	return user, nil
}

func (u *userService) GetUserByID(ctx context.Context, id uint64) (*User, error) {
	res, err := u.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("[Domain.User] get user by id: %w", err)
	}
	return res, nil
}

func (u *userService) GetUser(ctx context.Context, user *User) (*User, error) {
	// TODO implement me
	panic("implement me")
}

func NewUserService(srv *domain.Service, repo UserRepository) UserService {
	return &userService{
		srv:  srv,
		repo: repo,
	}
}
