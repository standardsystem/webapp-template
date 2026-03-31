package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/handler"
)

// --- モック: SessionService ---

type mockSessionService struct {
	claims *domain.SessionClaims
	err    error
}

func (m *mockSessionService) IssueToken(_ *domain.SessionClaims) (string, error) {
	return "mock-token", nil
}

func (m *mockSessionService) ValidateToken(_ string) (*domain.SessionClaims, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.claims, nil
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		cookie     *http.Cookie
		claims     *domain.SessionClaims
		valErr     error
		wantStatus int
	}{
		{
			name:       "正常系: 有効なセッション",
			cookie:     &http.Cookie{Name: "session_token", Value: "valid-token"},
			claims:     &domain.SessionClaims{UserID: "user-1", Email: "test@example.com", Role: domain.RoleMember},
			wantStatus: http.StatusOK,
		},
		{
			name:       "異常系: Cookie なし",
			cookie:     nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "異常系: 無効なトークン",
			cookie:     &http.Cookie{Name: "session_token", Value: "invalid"},
			valErr:     domain.ErrUnauthorized,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockSessionService{claims: tt.claims, err: tt.valErr}
			mw := handler.NewAuthMiddleware(svc)

			var capturedUserID string
			inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedUserID = handler.UserIDFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			rec := httptest.NewRecorder()

			mw.Handler()(inner).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusOK && capturedUserID != tt.claims.UserID {
				t.Errorf("userID = %s, want %s", capturedUserID, tt.claims.UserID)
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name       string
		claims     *domain.SessionClaims
		required   domain.Role
		wantStatus int
	}{
		{
			name:       "正常系: admin が admin エンドポイントにアクセス",
			claims:     &domain.SessionClaims{UserID: "u1", Role: domain.RoleAdmin},
			required:   domain.RoleAdmin,
			wantStatus: http.StatusOK,
		},
		{
			name:       "異常系: member が admin エンドポイントにアクセス",
			claims:     &domain.SessionClaims{UserID: "u1", Role: domain.RoleMember},
			required:   domain.RoleAdmin,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "正常系: member が member エンドポイントにアクセス",
			claims:     &domain.SessionClaims{UserID: "u1", Role: domain.RoleMember},
			required:   domain.RoleMember,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockSessionService{claims: tt.claims}
			authMW := handler.NewAuthMiddleware(svc)
			roleMW := handler.RequireRole(tt.required)

			inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.AddCookie(&http.Cookie{Name: "session_token", Value: "token"})
			rec := httptest.NewRecorder()

			authMW.Handler()(roleMW(inner)).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
