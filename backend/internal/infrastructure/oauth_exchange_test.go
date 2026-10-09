package infrastructure_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/infrastructure"
)

const (
	googleTokenURL    = "https://oauth2.googleapis.com/token"
	googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	githubTokenURL    = "https://github.com/login/oauth/access_token"
	githubUserURL     = "https://api.github.com/user"
	githubEmailsURL   = "https://api.github.com/user/emails"
	microsoftTenant   = "contoso.onmicrosoft.com"
	microsoftTokenURL = "https://login.microsoftonline.com/" + microsoftTenant + "/oauth2/v2.0/token"
	microsoftGraphURL = "https://graph.microsoft.com/v1.0/me"
)

// newProvider は、stub の応答を返す HTTP クライアントを持つプロバイダを生成します。
func newProvider(name string, bodies map[string]string) domain.OAuthProvider {
	client, _ := newStubClient(bodies)
	switch name {
	case "google":
		return infrastructure.NewGoogleOAuthProvider("client-id", "secret", "http://localhost/cb", client)
	case "github":
		return infrastructure.NewGitHubOAuthProvider("client-id", "secret", "http://localhost/cb", client)
	default:
		return infrastructure.NewMicrosoftOAuthProvider("client-id", "secret", "http://localhost/cb", microsoftTenant, client)
	}
}

func TestOAuthProviders_NameAndAuthURL(t *testing.T) {
	tests := []struct {
		provider   string
		wantPrefix string
		wantScope  string
	}{
		{"google", "https://accounts.google.com/o/oauth2/v2/auth?", "openid email profile"},
		{"github", "https://github.com/login/oauth/authorize?", "read:user user:email"},
		{"microsoft", "https://login.microsoftonline.com/" + microsoftTenant + "/oauth2/v2.0/authorize?", "openid email profile User.Read"},
	}
	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			p := newProvider(tt.provider, nil)

			if p.Name() != tt.provider {
				t.Errorf("Name() = %s, want %s", p.Name(), tt.provider)
			}

			authURL := p.AuthURL("state-123")
			if !strings.HasPrefix(authURL, tt.wantPrefix) {
				t.Fatalf("AuthURL = %s, want prefix %s", authURL, tt.wantPrefix)
			}
			parsed, err := url.Parse(authURL)
			if err != nil {
				t.Fatalf("AuthURL is not a valid URL: %v", err)
			}
			q := parsed.Query()
			for key, want := range map[string]string{
				"client_id":    "client-id",
				"redirect_uri": "http://localhost/cb",
				"state":        "state-123",
				"scope":        tt.wantScope,
			} {
				if got := q.Get(key); got != want {
					t.Errorf("query %s = %q, want %q", key, got, want)
				}
			}
		})
	}
}

func TestOAuthProviders_Exchange(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		bodies      map[string]string
		wantAccess  string
		wantRefresh string
		wantErr     bool
	}{
		{
			name:        "google: トークンを取得",
			provider:    "google",
			bodies:      map[string]string{googleTokenURL: `{"access_token":"at","refresh_token":"rt"}`},
			wantAccess:  "at",
			wantRefresh: "rt",
		},
		{
			name:     "google: トークンエンドポイントが 200 以外",
			provider: "google",
			bodies:   map[string]string{},
			wantErr:  true,
		},
		{
			name:     "google: 応答が JSON でない",
			provider: "google",
			bodies:   map[string]string{googleTokenURL: `<html>`},
			wantErr:  true,
		},
		{
			name:       "github: トークンを取得",
			provider:   "github",
			bodies:     map[string]string{githubTokenURL: `{"access_token":"at"}`},
			wantAccess: "at",
		},
		{
			name:     "github: 200 でも error フィールドがあれば失敗",
			provider: "github",
			bodies:   map[string]string{githubTokenURL: `{"error":"bad_verification_code"}`},
			wantErr:  true,
		},
		{
			name:     "github: トークンエンドポイントが 200 以外",
			provider: "github",
			bodies:   map[string]string{},
			wantErr:  true,
		},
		{
			name:     "github: 応答が JSON でない",
			provider: "github",
			bodies:   map[string]string{githubTokenURL: `<html>`},
			wantErr:  true,
		},
		{
			name:        "microsoft: トークンを取得",
			provider:    "microsoft",
			bodies:      map[string]string{microsoftTokenURL: `{"access_token":"at","refresh_token":"rt"}`},
			wantAccess:  "at",
			wantRefresh: "rt",
		},
		{
			name:     "microsoft: トークンエンドポイントが 200 以外",
			provider: "microsoft",
			bodies:   map[string]string{},
			wantErr:  true,
		},
		{
			name:     "microsoft: 応答が JSON でない",
			provider: "microsoft",
			bodies:   map[string]string{microsoftTokenURL: `<html>`},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newProvider(tt.provider, tt.bodies)

			token, err := p.Exchange(context.Background(), "auth-code")

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if token.AccessToken != tt.wantAccess {
				t.Errorf("AccessToken = %q, want %q", token.AccessToken, tt.wantAccess)
			}
			if token.RefreshToken != tt.wantRefresh {
				t.Errorf("RefreshToken = %q, want %q", token.RefreshToken, tt.wantRefresh)
			}
		})
	}
}

func TestOAuthProviders_UserInfo_Errors(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		bodies   map[string]string
	}{
		{"google: userinfo が 200 以外", "google", map[string]string{}},
		{"google: 応答が JSON でない", "google", map[string]string{googleUserInfoURL: `<html>`}},
		{"github: user が 200 以外", "github", map[string]string{}},
		{"github: user の応答が JSON でない", "github", map[string]string{githubUserURL: `<html>`}},
		{"github: emails の応答が JSON でない", "github", map[string]string{githubUserURL: `{"id":1,"login":"octo"}`, githubEmailsURL: `<html>`}},
		{"microsoft: me が 200 以外", "microsoft", map[string]string{}},
		{"microsoft: 応答が JSON でない", "microsoft", map[string]string{microsoftGraphURL: `<html>`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newProvider(tt.provider, tt.bodies)

			info, err := p.UserInfo(context.Background(), &domain.OAuthToken{AccessToken: "at"})
			if err == nil {
				t.Fatalf("expected error but got %+v", info)
			}
		})
	}
}

func TestOAuthProviders_UserInfo_Fallbacks(t *testing.T) {
	t.Run("github: name が空なら login を名前にする", func(t *testing.T) {
		p := newProvider("github", map[string]string{
			githubUserURL:   `{"id":42,"login":"octo","name":"","avatar_url":"https://example.com/a.png"}`,
			githubEmailsURL: `[{"email":"octo@example.com","primary":true,"verified":true}]`,
		})

		info, err := p.UserInfo(context.Background(), &domain.OAuthToken{AccessToken: "at"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Name != "octo" {
			t.Errorf("Name = %q, want octo", info.Name)
		}
		if info.AvatarURL != "https://example.com/a.png" {
			t.Errorf("AvatarURL = %q", info.AvatarURL)
		}
	})

	t.Run("microsoft: mail が空なら userPrincipalName をメールにする", func(t *testing.T) {
		p := newProvider("microsoft", map[string]string{
			microsoftGraphURL: `{"id":"oid-1","displayName":"A","mail":"","userPrincipalName":"a@contoso.onmicrosoft.com"}`,
		})

		info, err := p.UserInfo(context.Background(), &domain.OAuthToken{AccessToken: "at"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Email != "a@contoso.onmicrosoft.com" {
			t.Errorf("Email = %q, want a@contoso.onmicrosoft.com", info.Email)
		}
	})
}
