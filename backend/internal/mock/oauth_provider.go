package mock

import (
	"context"

	"github.com/your-org/webapp-template/internal/domain"
)

// OAuthProvider は domain.OAuthProvider のモック実装です。
type OAuthProvider struct {
	ProviderName string
	AuthBaseURL  string
	Token        *domain.OAuthToken
	UserInfo_    *domain.OAuthUserInfo
	ExchangeErr  error
	UserInfoErr  error
}

func (m *OAuthProvider) Name() string                { return m.ProviderName }
func (m *OAuthProvider) AuthURL(state string) string { return m.AuthBaseURL + "?state=" + state }

func (m *OAuthProvider) Exchange(_ context.Context, _ string) (*domain.OAuthToken, error) {
	if m.ExchangeErr != nil {
		return nil, m.ExchangeErr
	}
	return m.Token, nil
}

func (m *OAuthProvider) UserInfo(_ context.Context, _ *domain.OAuthToken) (*domain.OAuthUserInfo, error) {
	if m.UserInfoErr != nil {
		return nil, m.UserInfoErr
	}
	return m.UserInfo_, nil
}
