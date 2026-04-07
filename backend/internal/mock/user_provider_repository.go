package mock

import (
	"context"

	"github.com/your-org/webapp-template/internal/domain"
)

// UserProviderRepository は domain.UserProviderRepository のインメモリモック実装です。
type UserProviderRepository struct {
	Providers map[string]*domain.UserProvider // key: provider:providerID
	SaveErr   error
}

// NewUserProviderRepository は UserProviderRepository を生成します。
func NewUserProviderRepository() *UserProviderRepository {
	return &UserProviderRepository{Providers: make(map[string]*domain.UserProvider)}
}

func (m *UserProviderRepository) FindByProviderAndProviderID(_ context.Context, provider, providerID string) (*domain.UserProvider, error) {
	up, ok := m.Providers[provider+":"+providerID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return up, nil
}

func (m *UserProviderRepository) FindByUserID(_ context.Context, userID string) ([]*domain.UserProvider, error) {
	var result []*domain.UserProvider
	for _, up := range m.Providers {
		if up.UserID == userID {
			result = append(result, up)
		}
	}
	return result, nil
}

func (m *UserProviderRepository) Save(_ context.Context, up *domain.UserProvider) error {
	if m.SaveErr != nil {
		return m.SaveErr
	}
	m.Providers[up.Provider+":"+up.ProviderID] = up
	return nil
}

func (m *UserProviderRepository) Delete(_ context.Context, id string) error {
	for key, up := range m.Providers {
		if up.ID == id {
			delete(m.Providers, key)
			break
		}
	}
	return nil
}
