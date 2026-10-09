package infrastructure_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/infrastructure"
)

// stubTransport は URL ごとに決まった JSON を返す http.RoundTripper です。
type stubTransport struct {
	bodies   map[string]string
	requests []string
}

func (s *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	u := req.URL.Scheme + "://" + req.URL.Host + req.URL.Path
	s.requests = append(s.requests, u)

	body, ok := s.bodies[u]
	status := http.StatusOK
	if !ok {
		status = http.StatusNotFound
	}
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func newStubClient(bodies map[string]string) (*http.Client, *stubTransport) {
	transport := &stubTransport{bodies: bodies}
	return &http.Client{Transport: transport}, transport
}

func TestGoogleOAuthProvider_UserInfo(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		wantVerified bool
	}{
		{"検証済みメール", `{"id":"1","email":"a@example.com","verified_email":true,"name":"A"}`, true},
		{"未検証メール", `{"id":"1","email":"a@example.com","verified_email":false,"name":"A"}`, false},
		{"verified_email がない", `{"id":"1","email":"a@example.com","name":"A"}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := newStubClient(map[string]string{
				"https://www.googleapis.com/oauth2/v2/userinfo": tt.body,
			})
			p := infrastructure.NewGoogleOAuthProvider("id", "secret", "http://localhost/cb", client)

			info, err := p.UserInfo(context.Background(), &domain.OAuthToken{AccessToken: "at"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info.EmailVerified != tt.wantVerified {
				t.Errorf("EmailVerified = %v, want %v", info.EmailVerified, tt.wantVerified)
			}
			if info.Email != "a@example.com" {
				t.Errorf("Email = %s, want a@example.com", info.Email)
			}
		})
	}
}

func TestGitHubOAuthProvider_UserInfo(t *testing.T) {
	const userBody = `{"id":42,"login":"octo","email":"public@example.com"}`

	t.Run("正常系: 検証済みのプライマリメールを使う", func(t *testing.T) {
		client, _ := newStubClient(map[string]string{
			"https://api.github.com/user": userBody,
			"https://api.github.com/user/emails": `[
				{"email":"unverified@example.com","primary":false,"verified":false},
				{"email":"primary@example.com","primary":true,"verified":true}
			]`,
		})
		p := infrastructure.NewGitHubOAuthProvider("id", "secret", "http://localhost/cb", client)

		info, err := p.UserInfo(context.Background(), &domain.OAuthToken{AccessToken: "at"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Email != "primary@example.com" {
			t.Errorf("Email = %s, want primary@example.com", info.Email)
		}
		if !info.EmailVerified {
			t.Error("EmailVerified = false, want true")
		}
		if info.ProviderID != "42" {
			t.Errorf("ProviderID = %s, want 42", info.ProviderID)
		}
	})

	t.Run("異常系: 検証済みのプライマリメールがない", func(t *testing.T) {
		client, _ := newStubClient(map[string]string{
			"https://api.github.com/user":        userBody,
			"https://api.github.com/user/emails": `[{"email":"primary@example.com","primary":true,"verified":false}]`,
		})
		p := infrastructure.NewGitHubOAuthProvider("id", "secret", "http://localhost/cb", client)

		if _, err := p.UserInfo(context.Background(), &domain.OAuthToken{AccessToken: "at"}); err == nil {
			t.Error("expected error when no verified primary email exists")
		}
	})

	t.Run("異常系: emails API がエラーを返す", func(t *testing.T) {
		client, _ := newStubClient(map[string]string{
			"https://api.github.com/user": userBody,
		})
		p := infrastructure.NewGitHubOAuthProvider("id", "secret", "http://localhost/cb", client)

		if _, err := p.UserInfo(context.Background(), &domain.OAuthToken{AccessToken: "at"}); err == nil {
			t.Error("expected error when emails API fails")
		}
	})
}

func TestMicrosoftOAuthProvider(t *testing.T) {
	const graphBody = `{"id":"oid-1","displayName":"A","mail":"victim@example.com","userPrincipalName":"a@contoso.onmicrosoft.com"}`

	tests := []struct {
		name         string
		tenantID     string
		wantTenant   string
		wantVerified bool
	}{
		{"テナント未設定は common で未検証", "", "common", false},
		{"common は未検証", "common", "common", false},
		{"organizations は未検証", "Organizations", "organizations", false},
		{"consumers は未検証", "consumers", "consumers", false},
		{"テナント ID を指定すると検証済み", "11111111-2222-3333-4444-555555555555", "11111111-2222-3333-4444-555555555555", true},
		{"テナントのドメインを指定すると検証済み", " contoso.onmicrosoft.com ", "contoso.onmicrosoft.com", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenURL := "https://login.microsoftonline.com/" + tt.wantTenant + "/oauth2/v2.0/token"
			client, transport := newStubClient(map[string]string{
				tokenURL:                              `{"access_token":"at"}`,
				"https://graph.microsoft.com/v1.0/me": graphBody,
			})
			p := infrastructure.NewMicrosoftOAuthProvider("id", "secret", "http://localhost/cb", tt.tenantID, client)

			wantAuthPrefix := "https://login.microsoftonline.com/" + tt.wantTenant + "/oauth2/v2.0/authorize?"
			if got := p.AuthURL("state"); !strings.HasPrefix(got, wantAuthPrefix) {
				t.Errorf("AuthURL = %s, want prefix %s", got, wantAuthPrefix)
			}
			if p.RestrictsTenant() != tt.wantVerified {
				t.Errorf("RestrictsTenant = %v, want %v", p.RestrictsTenant(), tt.wantVerified)
			}

			token, err := p.Exchange(context.Background(), "code")
			if err != nil {
				t.Fatalf("unexpected exchange error: %v", err)
			}
			if transport.requests[0] != tokenURL {
				t.Errorf("token request = %s, want %s", transport.requests[0], tokenURL)
			}

			info, err := p.UserInfo(context.Background(), token)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info.EmailVerified != tt.wantVerified {
				t.Errorf("EmailVerified = %v, want %v", info.EmailVerified, tt.wantVerified)
			}
			if info.Email != "victim@example.com" {
				t.Errorf("Email = %s, want victim@example.com", info.Email)
			}
		})
	}
}
