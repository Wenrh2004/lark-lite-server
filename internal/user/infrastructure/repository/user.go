package repository

import (
	"context"
	"fmt"

	"github.com/Wenrh2004/lark-lite-server/internal/user/domain"
	"github.com/Wenrh2004/lark-lite-server/internal/user/infrastructure/model"
)

type UserRepository struct {
	repo *Repository
}

func (u *UserRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
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

func (u *UserRepository) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	update := map[string]interface{}{
		"nickname":       user.Nickname.String(),
		"avatar_url":     user.AvatarURL,
		"background_url": user.BackgroundURL,
		"signature":      user.Signature,
		"email":          user.Email,
		"phone":          user.Phone,
		"gender":         byte(user.Gender),
	}
	_, err := u.repo.query.User.WithContext(ctx).
		Where(u.repo.query.User.ID.Eq(user.ID)).
		Updates(update)
	if err != nil {
		return nil, fmt.Errorf("[Infrastructure.Repository.User]failed to update user: %w", err)
	}
	return user, nil
}

func (u *UserRepository) GetUserByID(ctx context.Context, id uint64) (*domain.User, error) {
	// TODO implement me
	panic("implement me")
}

func (u *UserRepository) GetUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	// TODO implement me
	panic("implement me")
}

func NewUserRepository(repo *Repository) domain.UserRepository {
	return &UserRepository{
		repo: repo,
	}
}
