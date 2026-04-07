package mock

import (
	"context"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/usecase"
)

// AuthService は handler.AuthService のモック実装です。
type AuthService struct {
	AuthURL_        string
	State_          string
	GetAuthURLErr   error
	CallbackResult  *usecase.AuthCallbackResult
	CallbackErr     error
	CurrentUser     *domain.User
	CurrentUserErr  error
	UpdateRoleErr   error
}

func (m *AuthService) GetAuthURL(_ string) (string, string, error) {
	if m.GetAuthURLErr != nil {
		return "", "", m.GetAuthURLErr
	}
	return m.AuthURL_, m.State_, nil
}

func (m *AuthService) HandleCallback(_ context.Context, _, _ string) (*usecase.AuthCallbackResult, error) {
	if m.CallbackErr != nil {
		return nil, m.CallbackErr
	}
	return m.CallbackResult, nil
}

func (m *AuthService) GetCurrentUser(_ context.Context, _ string) (*domain.User, error) {
	if m.CurrentUserErr != nil {
		return nil, m.CurrentUserErr
	}
	return m.CurrentUser, nil
}

func (m *AuthService) UpdateUserRole(_ context.Context, _ string, _ domain.Role) error {
	return m.UpdateRoleErr
}
