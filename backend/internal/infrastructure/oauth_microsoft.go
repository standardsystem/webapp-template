package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/your-org/webapp-template/internal/domain"
)

const (
	microsoftLoginBaseURL = "https://login.microsoftonline.com/"
	microsoftUserInfoURL  = "https://graph.microsoft.com/v1.0/me"
	microsoftCommonTenant = "common"
)

// microsoftMultiTenantAliases は、複数のテナントからのサインインを受け付けるエンドポイントの別名です。
var microsoftMultiTenantAliases = map[string]bool{
	microsoftCommonTenant: true,
	"organizations":       true,
	"consumers":           true,
}

// MicrosoftOAuthProvider は Microsoft の OAuth2/OIDC プロバイダ実装です。
type MicrosoftOAuthProvider struct {
	clientID     string
	clientSecret string
	redirectURL  string
	tenantID     string
	httpClient   *http.Client
}

// NewMicrosoftOAuthProvider は MicrosoftOAuthProvider を生成します。
//
// tenantID にはサインインを許可する Entra テナントの ID かドメインを渡します。
// Microsoft Graph の mail と userPrincipalName はテナントの管理者が任意の値に設定できるため、
// 自組織のテナントに限定したときだけ EmailVerified を真にして、新規ユーザーの作成を許可します。
// 空文字や common・organizations・consumers を渡すと EmailVerified は偽になります。
// ゲストユーザーの mail は古いままのことがあるため、真でも現在の所有者は保証されません。
func NewMicrosoftOAuthProvider(clientID, clientSecret, redirectURL, tenantID string, httpClient *http.Client) *MicrosoftOAuthProvider {
	tenantID = strings.ToLower(strings.TrimSpace(tenantID))
	if tenantID == "" {
		tenantID = microsoftCommonTenant
	}
	return &MicrosoftOAuthProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		tenantID:     tenantID,
		httpClient:   httpClient,
	}
}

// RestrictsTenant は、サインインを単一のテナントに限定しているかどうかを返します。
func (m *MicrosoftOAuthProvider) RestrictsTenant() bool {
	return !microsoftMultiTenantAliases[m.tenantID]
}

func (m *MicrosoftOAuthProvider) endpoint(name string) string {
	return microsoftLoginBaseURL + url.PathEscape(m.tenantID) + "/oauth2/v2.0/" + name
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
	return m.endpoint("authorize") + "?" + params.Encode()
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.endpoint("token"), strings.NewReader(data.Encode()))
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
		return nil, responseError("token exchange", resp)
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
		return nil, responseError("userinfo request", resp)
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
		ProviderID:    info.ID,
		Email:         email,
		EmailVerified: m.RestrictsTenant(),
		Name:          info.DisplayName,
	}, nil
}
