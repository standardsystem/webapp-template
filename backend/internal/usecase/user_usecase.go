// Package usecase はビジネスロジックを実装します。
// domain パッケージのみに依存し、handler や repository の実装には依存しません。
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/your-org/webapp-template/internal/domain"
)

// UserUsecase はユーザー操作のユースケースです。
type UserUsecase struct {
	repo domain.UserRepository
}

// NewUserUsecase は UserUsecase を生成します。
func NewUserUsecase(repo domain.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

// CreateUserInput はユーザー作成の入力値です。
type CreateUserInput struct {
	Name  string
	Email string
}

// CreateUser は新しいユーザーを作成します。
func (u *UserUsecase) CreateUser(ctx context.Context, input CreateUserInput) (*domain.User, error) {
	user := &domain.User{
		ID:        generateID(),
		Name:      input.Name,
		Email:     input.Email,
		Role:      domain.RoleMember,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := u.repo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

// GetUser は指定 ID のユーザーを取得します。
func (u *UserUsecase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return user, nil
}

// ListUsers は全ユーザーを取得します。
func (u *UserUsecase) ListUsers(ctx context.Context) ([]*domain.User, error) {
	users, err := u.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func generateID() string {
	return uuid.New().String()
}
