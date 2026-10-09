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
		userRepo.ProviderRepo = providerRepo
		sessionSvc := &mock.SessionService{Token: "session-token"}
		provider := &mock.OAuthProvider{
			ProviderName: "google",
			Token:        &domain.OAuthToken{AccessToken: "at"},
			UserInfo_: &domain.OAuthUserInfo{
				ProviderID:    "google-123",
				Email:         "test@example.com",
				EmailVerified: true,
				Name:          "テストユーザー",
				AvatarURL:     "https://example.com/avatar.png",
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
		userRepo.ProviderRepo = providerRepo
		sessionSvc := &mock.SessionService{Token: "session-token-2"}
		provider := &mock.OAuthProvider{
			ProviderName: "google",
			Token:        &domain.OAuthToken{AccessToken: "at"},
			UserInfo_: &domain.OAuthUserInfo{
				ProviderID:    "google-456",
				Email:         "second@example.com",
				EmailVerified: true,
				Name:          "Second User",
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

	t.Run("異常系: 検証済みメールが一致しても既存ユーザーに連携しない", func(t *testing.T) {
		userRepo := mock.NewUserRepository()
		userRepo.Users["user-1"] = &domain.User{
			ID: "user-1", Name: "Victim", Email: "victim@example.com", Role: domain.RoleAdmin,
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		providerRepo := mock.NewUserProviderRepository()
		providerRepo.Providers["google:google-123"] = &domain.UserProvider{
			ID: "up-1", UserID: "user-1", Provider: "google", ProviderID: "google-123",
		}
		sessionSvc := &mock.SessionService{Token: "session-token"}
		provider := &mock.OAuthProvider{
			ProviderName: "github",
			Token:        &domain.OAuthToken{AccessToken: "at"},
			UserInfo_: &domain.OAuthUserInfo{
				ProviderID:    "github-789",
				Email:         "victim@example.com",
				EmailVerified: true,
				Name:          "Victim",
			},
		}

		uc := usecase.NewAuthUsecase(userRepo, providerRepo, sessionSvc, map[string]domain.OAuthProvider{
			"github": provider,
		})

		result, err := uc.HandleCallback(context.Background(), "github", "auth-code")
		if !errors.Is(err, domain.ErrAccountLinkRequired) {
			t.Fatalf("err = %v, want ErrAccountLinkRequired", err)
		}
		if result != nil {
			t.Error("result should be nil when login is rejected")
		}
		if len(userRepo.Users) != 1 {
			t.Errorf("users = %d, want 1", len(userRepo.Users))
		}
		if len(providerRepo.Providers) != 1 {
			t.Errorf("provider links = %d, want 1 (must not be linked)", len(providerRepo.Providers))
		}
	})

	t.Run("異常系: 未検証メールが一致しても既存ユーザーに連携しない", func(t *testing.T) {
		userRepo := mock.NewUserRepository()
		userRepo.Users["user-1"] = &domain.User{
			ID: "user-1", Name: "Victim", Email: "victim@example.com", Role: domain.RoleAdmin,
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		providerRepo := mock.NewUserProviderRepository()
		providerRepo.Providers["google:google-123"] = &domain.UserProvider{
			ID: "up-1", UserID: "user-1", Provider: "google", ProviderID: "google-123",
		}
		sessionSvc := &mock.SessionService{Token: "session-token"}
		provider := &mock.OAuthProvider{
			ProviderName: "microsoft",
			Token:        &domain.OAuthToken{AccessToken: "at"},
			UserInfo_: &domain.OAuthUserInfo{
				ProviderID:    "attacker-oid",
				Email:         "victim@example.com",
				EmailVerified: false,
				Name:          "Attacker",
			},
		}

		uc := usecase.NewAuthUsecase(userRepo, providerRepo, sessionSvc, map[string]domain.OAuthProvider{
			"microsoft": provider,
		})

		result, err := uc.HandleCallback(context.Background(), "microsoft", "auth-code")
		if !errors.Is(err, domain.ErrEmailNotVerified) {
			t.Fatalf("err = %v, want ErrEmailNotVerified", err)
		}
		if result != nil {
			t.Error("result should be nil when login is rejected")
		}
		if len(providerRepo.Providers) != 1 {
			t.Errorf("provider links = %d, want 1 (attacker must not be linked)", len(providerRepo.Providers))
		}
	})

	t.Run("異常系: 未検証メールでは新規ユーザーを作成しない", func(t *testing.T) {
		userRepo := mock.NewUserRepository()
		providerRepo := mock.NewUserProviderRepository()
		sessionSvc := &mock.SessionService{Token: "session-token"}
		provider := &mock.OAuthProvider{
			ProviderName: "microsoft",
			Token:        &domain.OAuthToken{AccessToken: "at"},
			UserInfo_: &domain.OAuthUserInfo{
				ProviderID:    "attacker-oid",
				Email:         "victim@example.com",
				EmailVerified: false,
				Name:          "Attacker",
			},
		}

		uc := usecase.NewAuthUsecase(userRepo, providerRepo, sessionSvc, map[string]domain.OAuthProvider{
			"microsoft": provider,
		})

		_, err := uc.HandleCallback(context.Background(), "microsoft", "auth-code")
		if !errors.Is(err, domain.ErrEmailNotVerified) {
			t.Fatalf("err = %v, want ErrEmailNotVerified", err)
		}
		if len(userRepo.Users) != 0 {
			t.Errorf("users = %d, want 0", len(userRepo.Users))
		}
		if len(providerRepo.Providers) != 0 {
			t.Errorf("provider links = %d, want 0", len(providerRepo.Providers))
		}
	})

	t.Run("正常系: 連携済みなら未検証メールでも再ログインできる", func(t *testing.T) {
		userRepo := mock.NewUserRepository()
		userRepo.Users["user-1"] = &domain.User{
			ID: "user-1", Name: "Old Name", Email: "test@example.com", Role: domain.RoleMember,
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		providerRepo := mock.NewUserProviderRepository()
		providerRepo.Providers["microsoft:ms-1"] = &domain.UserProvider{
			ID: "up-1", UserID: "user-1", Provider: "microsoft", ProviderID: "ms-1",
		}
		sessionSvc := &mock.SessionService{Token: "session-token"}
		provider := &mock.OAuthProvider{
			ProviderName: "microsoft",
			Token:        &domain.OAuthToken{AccessToken: "at"},
			UserInfo_: &domain.OAuthUserInfo{
				ProviderID: "ms-1",
				Email:      "changed@example.com",
				Name:       "New Name",
			},
		}

		uc := usecase.NewAuthUsecase(userRepo, providerRepo, sessionSvc, map[string]domain.OAuthProvider{
			"microsoft": provider,
		})

		result, err := uc.HandleCallback(context.Background(), "microsoft", "auth-code")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.User.ID != "user-1" {
			t.Errorf("user id = %s, want user-1", result.User.ID)
		}
		if result.User.Email != "test@example.com" {
			t.Errorf("email = %s, want test@example.com", result.User.Email)
		}
	})

	t.Run("異常系: 紐付けの保存に失敗してもユーザーを残さず、再試行で登録できる", func(t *testing.T) {
		userRepo := mock.NewUserRepository()
		providerRepo := mock.NewUserProviderRepository()
		userRepo.ProviderRepo = providerRepo
		providerRepo.SaveErr = errors.New("connection reset")
		sessionSvc := &mock.SessionService{Token: "session-token"}
		provider := &mock.OAuthProvider{
			ProviderName: "google",
			Token:        &domain.OAuthToken{AccessToken: "at"},
			UserInfo_: &domain.OAuthUserInfo{
				ProviderID:    "google-123",
				Email:         "test@example.com",
				EmailVerified: true,
				Name:          "テストユーザー",
			},
		}

		uc := usecase.NewAuthUsecase(userRepo, providerRepo, sessionSvc, map[string]domain.OAuthProvider{
			"google": provider,
		})

		if _, err := uc.HandleCallback(context.Background(), "google", "auth-code"); err == nil {
			t.Fatal("expected error when saving the provider link fails")
		}
		if len(userRepo.Users) != 0 {
			t.Fatalf("users = %d, want 0 (user must not remain without a provider link)", len(userRepo.Users))
		}

		// エラーが解消したあとの再試行で、新規登録が完了する
		providerRepo.SaveErr = nil
		result, err := uc.HandleCallback(context.Background(), "google", "auth-code")
		if err != nil {
			t.Fatalf("unexpected error on retry: %v", err)
		}
		if !result.IsNewUser {
			t.Error("expected new user on retry")
		}
		if result.User.Role != domain.RoleAdmin {
			t.Errorf("role = %s, want admin", result.User.Role)
		}

		// 以後は連携済みのユーザーとしてログインできる
		again, err := uc.HandleCallback(context.Background(), "google", "auth-code")
		if err != nil {
			t.Fatalf("unexpected error on login: %v", err)
		}
		if again.IsNewUser {
			t.Error("expected existing user on login")
		}
		if again.User.ID != result.User.ID {
			t.Errorf("user id = %s, want %s", again.User.ID, result.User.ID)
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
