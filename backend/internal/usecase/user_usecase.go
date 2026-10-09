// Package usecase はビジネスロジックを実装します。
// domain パッケージのみに依存し、handler や repository の実装には依存しません。
package usecase

import (
	"context"
	"fmt"

	"github.com/your-org/webapp-template/internal/domain"
)

// UserUsecase はユーザー参照のユースケースです。
// ユーザーの作成は OAuth ログイン（AuthUsecase.HandleCallback）だけが行います。
type UserUsecase struct {
	repo domain.UserRepository
}

// NewUserUsecase は UserUsecase を生成します。
func NewUserUsecase(repo domain.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
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
