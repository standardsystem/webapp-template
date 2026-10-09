package mock

import (
	"context"

	"github.com/your-org/webapp-template/internal/domain"
)

// UserService は handler.UserService のモック実装です。
type UserService struct {
	User    *domain.User
	GetErr  error
	Users   []*domain.User
	ListErr error
	// GotIDs は GetUser に渡された ID の記録です。
	GotIDs []string
}

func (m *UserService) GetUser(_ context.Context, id string) (*domain.User, error) {
	m.GotIDs = append(m.GotIDs, id)
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	return m.User, nil
}

func (m *UserService) ListUsers(_ context.Context) ([]*domain.User, error) {
	if m.ListErr != nil {
		return nil, m.ListErr
	}
	return m.Users, nil
}
