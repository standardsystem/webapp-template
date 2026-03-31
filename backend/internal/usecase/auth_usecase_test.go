package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/usecase"
)

// --- モック: OAuthProvider ---

type mockOAuthProvider struct {
	name      string
	authURL   string
	token     *domain.OAuthToken
	userInfo  *domain.OAuthUserInfo
	exchErr   error
	infoErr   error
}

func (m *mockOAuthProvider) Name() string                { return m.name }
func (m *mockOAuthProvider) AuthURL(state string) string { return m.authURL + "?state=" + state }
func (m *mockOAuthProvider) Exchange(_ context.Context, _ string) (*domain.OAuthToken, error) {
	if m.exchErr != nil {
		return nil, m.exchErr
	}
	return m.token, nil
}
func (m *mockOAuthProvider) UserInfo(_ context.Context, _ *domain.OAuthToken) (*domain.OAuthUserInfo, error) {
	if m.infoErr != nil {
		return nil, m.infoErr
	}
	return m.userInfo, nil
}

// --- モック: UserProviderRepository ---

type mockUserProviderRepository struct {
	providers map[string]*domain.UserProvider // key: provider:providerID
	saveErr   error
}

func newMockUserProviderRepository() *mockUserProviderRepository {
	return &mockUserProviderRepository{providers: make(map[string]*domain.UserProvider)}
}

func (m *mockUserProviderRepository) FindByProviderAndProviderID(_ context.Context, provider, providerID string) (*domain.UserProvider, error) {
	up, ok := m.providers[provider+":"+providerID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return up, nil
}

func (m *mockUserProviderRepository) FindByUserID(_ context.Context, userID string) ([]*domain.UserProvider, error) {
	var result []*domain.UserProvider
	for _, up := range m.providers {
		if up.UserID == userID {
			result = append(result, up)
		}
	}
	return result, nil
}

func (m *mockUserProviderRepository) Save(_ context.Context, up *domain.UserProvider) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.providers[up.Provider+":"+up.ProviderID] = up
	return nil
}

func (m *mockUserProviderRepository) Delete(_ context.Context, id string) error {
	for key, up := range m.providers {
		if up.ID == id {
			delete(m.providers, key)
			break
		}
	}
	return nil
}

// --- モック: SessionService ---

type mockSessionService struct {
	token   string
	claims  *domain.SessionClaims
	issErr  error
	valErr  error
}

func (m *mockSessionService) IssueToken(_ *domain.SessionClaims) (string, error) {
	if m.issErr != nil {
		return "", m.issErr
	}
	return m.token, nil
}

func (m *mockSessionService) ValidateToken(_ string) (*domain.SessionClaims, error) {
	if m.valErr != nil {
		return nil, m.valErr
	}
	return m.claims, nil
}

// --- テスト ---

func TestAuthUsecase_GetAuthURL(t *testing.T) {
	providers := map[string]domain.OAuthProvider{
		"google": &mockOAuthProvider{name: "google", authURL: "https://accounts.google.com/auth"},
	}
	uc := usecase.NewAuthUsecase(nil, nil, nil, providers)

	t.Run("正常系: 認可URLを取得", func(t *testing.T) {
		url, state, err := uc.GetAuthURL("google")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if url == "" {
			t.Error("url should not be empty")
		}
		if state == "" {
			t.Error("state should not be empty")
		}
	})

	t.Run("異常系: 未知のプロバイダ", func(t *testing.T) {
		_, _, err := uc.GetAuthURL("unknown")
		if err == nil {
			t.Error("expected error for unknown provider")
		}
	})
}

func TestAuthUsecase_HandleCallback(t *testing.T) {
	t.Run("正常系: 新規ユーザー（初回=admin）", func(t *testing.T) {
		userRepo := newMockUserRepository()
		providerRepo := newMockUserProviderRepository()
		sessionSvc := &mockSessionService{token: "session-token"}
		provider := &mockOAuthProvider{
			name: "google",
			token: &domain.OAuthToken{AccessToken: "at"},
			userInfo: &domain.OAuthUserInfo{
				ProviderID: "google-123",
				Email:      "test@example.com",
				Name:       "テストユーザー",
				AvatarURL:  "https://example.com/avatar.png",
			},
		}

		uc := usecase.NewAuthUsecase(userRepo, providerRepo, sessionSvc, map[string]domain.OAuthProvider{
			"google": provider,
		})

		result, err := uc.HandleCallback(context.Background(), "google", "auth-code")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsNewUser {
			t.Error("expected new user")
		}
		if result.User.Role != domain.RoleAdmin {
			t.Errorf("first user role = %s, want admin", result.User.Role)
		}
		if result.SessionToken != "session-token" {
			t.Errorf("session token = %s, want session-token", result.SessionToken)
		}
		if result.User.Email != "test@example.com" {
			t.Errorf("email = %s, want test@example.com", result.User.Email)
		}
	})

	t.Run("正常系: 2人目のユーザー（member）", func(t *testing.T) {
		userRepo := newMockUserRepository()
		// 既存ユーザーを追加
		userRepo.users["existing-user"] = &domain.User{
			ID: "existing-user", Name: "Existing", Email: "existing@example.com", Role: domain.RoleAdmin,
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		providerRepo := newMockUserProviderRepository()
		sessionSvc := &mockSessionService{token: "session-token-2"}
		provider := &mockOAuthProvider{
			name: "google",
			token: &domain.OAuthToken{AccessToken: "at"},
			userInfo: &domain.OAuthUserInfo{
				ProviderID: "google-456",
				Email:      "second@example.com",
				Name:       "Second User",
			},
		}

		uc := usecase.NewAuthUsecase(userRepo, providerRepo, sessionSvc, map[string]domain.OAuthProvider{
			"google": provider,
		})

		result, err := uc.HandleCallback(context.Background(), "google", "auth-code")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.User.Role != domain.RoleMember {
			t.Errorf("second user role = %s, want member", result.User.Role)
		}
	})

	t.Run("正常系: 既存ユーザーの再ログイン", func(t *testing.T) {
		userRepo := newMockUserRepository()
		userRepo.users["user-1"] = &domain.User{
			ID: "user-1", Name: "Old Name", Email: "test@example.com", Role: domain.RoleMember,
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		providerRepo := newMockUserProviderRepository()
		providerRepo.providers["google:google-123"] = &domain.UserProvider{
			ID: "up-1", UserID: "user-1", Provider: "google", ProviderID: "google-123",
		}
		sessionSvc := &mockSessionService{token: "session-token"}
		provider := &mockOAuthProvider{
			name: "google",
			token: &domain.OAuthToken{AccessToken: "at"},
			userInfo: &domain.OAuthUserInfo{
				ProviderID: "google-123",
				Email:      "test@example.com",
				Name:       "Updated Name",
				AvatarURL:  "https://example.com/new-avatar.png",
			},
		}

		uc := usecase.NewAuthUsecase(userRepo, providerRepo, sessionSvc, map[string]domain.OAuthProvider{
			"google": provider,
		})

		result, err := uc.HandleCallback(context.Background(), "google", "auth-code")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsNewUser {
			t.Error("expected existing user")
		}
		if result.User.Name != "Updated Name" {
			t.Errorf("name = %s, want Updated Name", result.User.Name)
		}
	})

	t.Run("異常系: コード交換失敗", func(t *testing.T) {
		provider := &mockOAuthProvider{
			name:    "google",
			exchErr: errors.New("exchange failed"),
		}
		uc := usecase.NewAuthUsecase(nil, nil, nil, map[string]domain.OAuthProvider{
			"google": provider,
		})

		_, err := uc.HandleCallback(context.Background(), "google", "bad-code")
		if err == nil {
			t.Error("expected error for failed exchange")
		}
	})
}

func TestAuthUsecase_UpdateUserRole(t *testing.T) {
	userRepo := newMockUserRepository()
	userRepo.users["user-1"] = &domain.User{
		ID: "user-1", Name: "Test", Email: "test@example.com", Role: domain.RoleMember,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	uc := usecase.NewAuthUsecase(userRepo, nil, nil, nil)

	t.Run("正常系: ロール変更", func(t *testing.T) {
		err := uc.UpdateUserRole(context.Background(), "user-1", domain.RoleAdmin)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if userRepo.users["user-1"].Role != domain.RoleAdmin {
			t.Errorf("role = %s, want admin", userRepo.users["user-1"].Role)
		}
	})

	t.Run("異常系: 不正なロール", func(t *testing.T) {
		err := uc.UpdateUserRole(context.Background(), "user-1", "superadmin")
		if err == nil {
			t.Error("expected error for invalid role")
		}
	})
}
