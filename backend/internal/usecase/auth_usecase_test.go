package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/mock"
	"github.com/your-org/webapp-template/internal/usecase"
)

func TestAuthUsecase_GetAuthURL(t *testing.T) {
	providers := map[string]domain.OAuthProvider{
		"google": &mock.OAuthProvider{ProviderName: "google", AuthBaseURL: "https://accounts.google.com/auth"},
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
		userRepo := mock.NewUserRepository()
		providerRepo := mock.NewUserProviderRepository()
		sessionSvc := &mock.SessionService{Token: "session-token"}
		provider := &mock.OAuthProvider{
			ProviderName: "google",
			Token:        &domain.OAuthToken{AccessToken: "at"},
			UserInfo_: &domain.OAuthUserInfo{
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
		userRepo := mock.NewUserRepository()
		userRepo.Users["existing-user"] = &domain.User{
			ID: "existing-user", Name: "Existing", Email: "existing@example.com", Role: domain.RoleAdmin,
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		providerRepo := mock.NewUserProviderRepository()
		sessionSvc := &mock.SessionService{Token: "session-token-2"}
		provider := &mock.OAuthProvider{
			ProviderName: "google",
			Token:        &domain.OAuthToken{AccessToken: "at"},
			UserInfo_: &domain.OAuthUserInfo{
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
		userRepo := mock.NewUserRepository()
		userRepo.Users["user-1"] = &domain.User{
			ID: "user-1", Name: "Old Name", Email: "test@example.com", Role: domain.RoleMember,
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		providerRepo := mock.NewUserProviderRepository()
		providerRepo.Providers["google:google-123"] = &domain.UserProvider{
			ID: "up-1", UserID: "user-1", Provider: "google", ProviderID: "google-123",
		}
		sessionSvc := &mock.SessionService{Token: "session-token"}
		provider := &mock.OAuthProvider{
			ProviderName: "google",
			Token:        &domain.OAuthToken{AccessToken: "at"},
			UserInfo_: &domain.OAuthUserInfo{
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
		provider := &mock.OAuthProvider{
			ProviderName: "google",
			ExchangeErr:  errors.New("exchange failed"),
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
	userRepo := mock.NewUserRepository()
	userRepo.Users["user-1"] = &domain.User{
		ID: "user-1", Name: "Test", Email: "test@example.com", Role: domain.RoleMember,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	uc := usecase.NewAuthUsecase(userRepo, nil, nil, nil)

	t.Run("正常系: ロール変更", func(t *testing.T) {
		err := uc.UpdateUserRole(context.Background(), "user-1", domain.RoleAdmin)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if userRepo.Users["user-1"].Role != domain.RoleAdmin {
			t.Errorf("role = %s, want admin", userRepo.Users["user-1"].Role)
		}
	})

	t.Run("異常系: 不正なロール", func(t *testing.T) {
		err := uc.UpdateUserRole(context.Background(), "user-1", "superadmin")
		if err == nil {
			t.Error("expected error for invalid role")
		}
	})
}
