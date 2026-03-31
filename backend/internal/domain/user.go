package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// --- エラー定義 ---

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
)

// --- ロール定義 ---

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

func (r Role) IsValid() bool {
	return r == RoleAdmin || r == RoleMember
}

// --- エンティティ ---

// User はユーザーエンティティです。パスワードは保持しません（認証は外部プロバイダに委譲）。
type User struct {
	ID        string
	Name      string
	Email     string
	AvatarURL string
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate はユーザーのバリデーションを行います。
func (u *User) Validate() error {
	if u.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if u.Email == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidInput)
	}
	if !u.Role.IsValid() {
		return fmt.Errorf("%w: invalid role: %s", ErrInvalidInput, u.Role)
	}
	return nil
}

// IsAdmin はユーザーが管理者かどうかを返します。
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// --- リポジトリインターフェース ---

// UserRepository はユーザーのデータアクセスインターフェースです。
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindAll(ctx context.Context) ([]*User, error)
	Save(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}
