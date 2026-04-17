package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/your-org/webapp-template/internal/domain"
)

const (
	microsoftAuthURL     = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	microsoftTokenURL    = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	microsoftUserInfoURL = "https://graph.microsoft.com/v1.0/me"
)

// MicrosoftOAuthProvider は Microsoft の OAuth2/OIDC プロバイダ実装です。
type MicrosoftOAuthProvider struct {
	clientID     string
	clientSecret string
	redirectURL  string
	httpClient   *http.Client
}

// NewMicrosoftOAuthProvider は MicrosoftOAuthProvider を生成します。
func NewMicrosoftOAuthProvider(clientID, clientSecret, redirectURL string, httpClient *http.Client) *MicrosoftOAuthProvider {
	return &MicrosoftOAuthProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		httpClient:   httpClient,
	}
}

func (m *MicrosoftOAuthProvider) Name() string {
	return "microsoft"
}

func (m *MicrosoftOAuthProvider) AuthURL(state string) string {
	params := url.Values{
		"client_id":     {m.clientID},
		"redirect_uri":  {m.redirectURL},
		"response_type": {"code"},
		"scope":         {"openid email profile User.Read"},
		"state":         {state},
		"response_mode": {"query"},
	}
	return microsoftAuthURL + "?" + params.Encode()
}

func (m *MicrosoftOAuthProvider) Exchange(ctx context.Context, code string) (*domain.OAuthToken, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {m.clientID},
		"client_secret": {m.clientSecret},
		"redirect_uri":  {m.redirectURL},
		"grant_type":    {"authorization_code"},
		"scope":         {"openid email profile User.Read"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, microsoftTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed (status %d): %s", resp.StatusCode, body)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &domain.OAuthToken{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

func (m *MicrosoftOAuthProvider) UserInfo(ctx context.Context, token *domain.OAuthToken) (*domain.OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, microsoftUserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo request failed (status %d): %s", resp.StatusCode, body)
	}

	var info struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Mail        string `json:"mail"`
		UPN         string `json:"userPrincipalName"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	email := info.Mail
	if email == "" {
		email = info.UPN
	}

	return &domain.OAuthUserInfo{
		ProviderID: info.ID,
		Email:      email,
		Name:       info.DisplayName,
	}, nil
}
