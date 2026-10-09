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

func TestParseSignupPolicy(t *testing.T) {
	tests := []struct {
		name           string
		adminEmails    string
		allowedDomains string
		wantRestricts  bool
		wantHasAdmins  bool
	}{
		{"どちらも未設定", "", "", false, false},
		{"空白とカンマだけは未設定と同じ", " , ,", " ,", false, false},
		{"@ だけのドメインは無視する", "", "@", false, false},
		{"両方を設定", "admin@example.com", "example.com", true, true},
		{"複数の値", "a@example.com, b@example.com", "example.com,example.co.jp", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := usecase.ParseSignupPolicy(tt.adminEmails, tt.allowedDomains)

			if got := p.RestrictsDomains(); got != tt.wantRestricts {
				t.Errorf("RestrictsDomains() = %v, want %v", got, tt.wantRestricts)
			}
			if got := p.HasInitialAdmins(); got != tt.wantHasAdmins {
				t.Errorf("HasInitialAdmins() = %v, want %v", got, tt.wantHasAdmins)
			}
		})
	}
}

// signupFixture は、プロバイダ google が指定のメールアドレスを返す AuthUsecase を組み立てます。
type signupFixture struct {
	uc           *usecase.AuthUsecase
	userRepo     *mock.UserRepository
	providerRepo *mock.UserProviderRepository
}

func newSignupFixture(email string, policy usecase.SignupPolicy) *signupFixture {
	userRepo := mock.NewUserRepository()
	providerRepo := mock.NewUserProviderRepository()
	userRepo.ProviderRepo = providerRepo
	provider := &mock.OAuthProvider{
		ProviderName: "google",
		Token:        &domain.OAuthToken{AccessToken: "at"},
		UserInfo_: &domain.OAuthUserInfo{
			ProviderID:    "google-123",
			Email:         email,
			EmailVerified: true,
			Name:          "テストユーザー",
		},
	}
	uc := usecase.NewAuthUsecase(
		userRepo, providerRepo, &mock.SessionService{Token: "session-token"},
		map[string]domain.OAuthProvider{"google": provider},
		usecase.WithSignupPolicy(policy),
	)
	return &signupFixture{uc: uc, userRepo: userRepo, providerRepo: providerRepo}
}

func TestAuthUsecase_HandleCallback_InitialRole(t *testing.T) {
	tests := []struct {
		name        string
		adminEmails string
		email       string
		wantRole    domain.Role
	}{
		{"初期管理者が未設定なら、最初のユーザーでも member", "", "first@example.com", domain.RoleMember},
		{"初期管理者に指定したメールは admin", "admin@example.com", "admin@example.com", domain.RoleAdmin},
		{"初期管理者の指定は大文字と小文字を区別しない", " Admin@Example.com ", "admin@EXAMPLE.com", domain.RoleAdmin},
		{"複数指定のどれかに一致すれば admin", "a@example.com,b@example.com", "b@example.com", domain.RoleAdmin},
		{"指定していないユーザーは、最初のログインでも member", "admin@example.com", "first@example.com", domain.RoleMember},
		{"前方一致では admin にならない", "admin@example.com", "admin@example.com.attacker.test", domain.RoleMember},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newSignupFixture(tt.email, usecase.ParseSignupPolicy(tt.adminEmails, ""))

			result, err := f.uc.HandleCallback(context.Background(), "google", "auth-code")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !result.IsNewUser {
				t.Error("expected new user")
			}
			if result.User.Role != tt.wantRole {
				t.Errorf("role = %s, want %s", result.User.Role, tt.wantRole)
			}
		})
	}
}

func TestAuthUsecase_HandleCallback_AllowedEmailDomains(t *testing.T) {
	tests := []struct {
		name           string
		allowedDomains string
		email          string
		wantAllowed    bool
	}{
		{"未設定ならどのドメインでもログインできる", "", "user@anywhere.test", true},
		{"許可したドメインはログインできる", "example.com", "user@example.com", true},
		{"ドメインの大文字と小文字は区別しない", "Example.COM", "user@EXAMPLE.com", true},
		{"先頭に @ を付けて指定してもよい", "@example.com", "user@example.com", true},
		{"複数指定のどれかに一致すればログインできる", "example.com, example.co.jp", "user@example.co.jp", true},
		{"許可していないドメインは拒否する", "example.com", "user@other.test", false},
		{"サブドメインは別のドメインとして拒否する", "example.com", "user@sub.example.com", false},
		{"後方一致するだけのドメインは拒否する", "example.com", "user@evil-example.com", false},
		{"ローカル部に許可ドメインを含むだけでは拒否する", "example.com", "user@example.com@other.test", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newSignupFixture(tt.email, usecase.ParseSignupPolicy("", tt.allowedDomains))

			result, err := f.uc.HandleCallback(context.Background(), "google", "auth-code")

			if tt.wantAllowed {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if result.SessionToken == "" {
					t.Error("session token should be issued")
				}
				return
			}
			if !errors.Is(err, domain.ErrEmailDomainNotAllowed) {
				t.Fatalf("err = %v, want ErrEmailDomainNotAllowed", err)
			}
			if result != nil {
				t.Error("result should be nil when login is rejected")
			}
			if len(f.userRepo.Users) != 0 {
				t.Errorf("users = %d, want 0 (user must not be created)", len(f.userRepo.Users))
			}
			if len(f.providerRepo.Providers) != 0 {
				t.Errorf("provider links = %d, want 0", len(f.providerRepo.Providers))
			}
		})
	}
}

func TestAuthUsecase_HandleCallback_DomainRestrictionOnExistingUsers(t *testing.T) {
	seed := func(f *signupFixture, email string) {
		f.userRepo.Users["user-1"] = &domain.User{
			ID: "user-1", Name: "Existing", Email: email, Role: domain.RoleMember,
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		f.providerRepo.Providers["google:google-123"] = &domain.UserProvider{
			ID: "up-1", UserID: "user-1", Provider: "google", ProviderID: "google-123",
		}
	}

	t.Run("異常系: 許可ドメインを後から絞ると、登録済みのユーザーもログインできない", func(t *testing.T) {
		f := newSignupFixture("user@old.test", usecase.ParseSignupPolicy("", "example.com"))
		seed(f, "user@old.test")

		result, err := f.uc.HandleCallback(context.Background(), "google", "auth-code")
		if !errors.Is(err, domain.ErrEmailDomainNotAllowed) {
			t.Fatalf("err = %v, want ErrEmailDomainNotAllowed", err)
		}
		if result != nil {
			t.Error("result should be nil when login is rejected")
		}
	})

	t.Run("異常系: 別のユーザーが登録済みのメールでも、許可外のドメインなら登録状況を答えない", func(t *testing.T) {
		f := newSignupFixture("victim@other.test", usecase.ParseSignupPolicy("", "example.com"))
		// 同じメールのユーザーが別のプロバイダで登録済み
		f.userRepo.Users["user-1"] = &domain.User{
			ID: "user-1", Name: "Victim", Email: "victim@other.test", Role: domain.RoleMember,
		}

		_, err := f.uc.HandleCallback(context.Background(), "google", "auth-code")
		if !errors.Is(err, domain.ErrEmailDomainNotAllowed) {
			t.Fatalf("err = %v, want ErrEmailDomainNotAllowed (not ErrAccountLinkRequired)", err)
		}
	})

	t.Run("正常系: 登録済みのユーザーは、初期管理者に指定されても再ログインで昇格しない", func(t *testing.T) {
		f := newSignupFixture("user@example.com", usecase.ParseSignupPolicy("user@example.com", ""))
		seed(f, "user@example.com")

		result, err := f.uc.HandleCallback(context.Background(), "google", "auth-code")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.User.Role != domain.RoleMember {
			t.Errorf("role = %s, want member", result.User.Role)
		}
	})
}
