// Package mock はテスト用のモック実装を提供します。
package mock

import (
	"context"

	"github.com/your-org/webapp-template/internal/domain"
)

// UserRepository は domain.UserRepository のインメモリモック実装です。
type UserRepository struct {
	Users   map[string]*domain.User
	SaveErr error
	FindErr error
}

// NewUserRepository は UserRepository を生成します。
func NewUserRepository() *UserRepository {
	return &UserRepository{Users: make(map[string]*domain.User)}
}

func (m *UserRepository) FindByID(_ context.Context, id string) (*domain.User, error) {
	if m.FindErr != nil {
		return nil, m.FindErr
	}
	u, ok := m.Users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (m *UserRepository) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	if m.FindErr != nil {
		return nil, m.FindErr
	}
	for _, u := range m.Users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *UserRepository) FindAll(_ context.Context) ([]*domain.User, error) {
	if m.FindErr != nil {
		return nil, m.FindErr
	}
	result := make([]*domain.User, 0, len(m.Users))
	for _, u := range m.Users {
		result = append(result, u)
	}
	return result, nil
}

func (m *UserRepository) Save(_ context.Context, user *domain.User) error {
	if m.SaveErr != nil {
		return m.SaveErr
	}
	m.Users[user.ID] = user
	return nil
}

func (m *UserRepository) Delete(_ context.Context, id string) error {
	delete(m.Users, id)
	return nil
}

func (m *UserRepository) Count(_ context.Context) (int64, error) {
	return int64(len(m.Users)), nil
}
