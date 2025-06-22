package repository

import (
	"context"
	"fmt"

	"github.com/Wenrh2004/lark-lite-server/internal/user/domain/service"
	"github.com/Wenrh2004/lark-lite-server/internal/user/infrastructure/model"
)

type UserRepository struct {
	repo *Repository
}

func (u *UserRepository) CreateUser(ctx context.Context, user *service.User) (*service.User, error) {
	if err := u.repo.query.User.WithContext(ctx).Create(&model.User{
		ID:       user.ID,
		Username: user.Username.String(),
		Password: user.Password.String(),
		Nickname: user.Nickname.String(),
	}); err != nil {
		return nil, fmt.Errorf("[Infrastructure.Repository.User]failed to create user: %w", err)
	}
	return user, nil
}

func (u *UserRepository) UpdateUser(ctx context.Context, user *service.User) (*service.User, error) {
	// TODO implement me
	panic("implement me")
}

func (u *UserRepository) GetUserByID(ctx context.Context, id uint64) (*service.User, error) {
	// TODO implement me
	panic("implement me")
}

func (u *UserRepository) GetUser(ctx context.Context, user *service.User) (*service.User, error) {
	// TODO implement me
	panic("implement me")
}

func NewUserRepository(repo *Repository) service.UserRepository {
	return &UserRepository{
		repo: repo,
	}
}
