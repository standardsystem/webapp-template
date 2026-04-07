package mock

import (
	"github.com/your-org/webapp-template/internal/domain"
)

// SessionService は domain.SessionService のモック実装です。
type SessionService struct {
	Token    string
	Claims   *domain.SessionClaims
	IssueErr error
	ValidErr error
}

func (m *SessionService) IssueToken(_ *domain.SessionClaims) (string, error) {
	if m.IssueErr != nil {
		return "", m.IssueErr
	}
	return m.Token, nil
}

func (m *SessionService) ValidateToken(_ string) (*domain.SessionClaims, error) {
	if m.ValidErr != nil {
		return nil, m.ValidErr
	}
	return m.Claims, nil
}
