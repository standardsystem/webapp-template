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
	githubAuthURL     = "https://github.com/login/oauth/authorize"
	githubTokenURL    = "https://github.com/login/oauth/access_token"
	githubUserInfoURL = "https://api.github.com/user"
	githubEmailURL    = "https://api.github.com/user/emails"
)

// GitHubOAuthProvider は GitHub の OAuth2 プロバイダ実装です。
type GitHubOAuthProvider struct {
	clientID     string
	clientSecret string
	redirectURL  string
	httpClient   *http.Client
}

// NewGitHubOAuthProvider は GitHubOAuthProvider を生成します。
func NewGitHubOAuthProvider(clientID, clientSecret, redirectURL string, httpClient *http.Client) *GitHubOAuthProvider {
	return &GitHubOAuthProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		httpClient:   httpClient,
	}
}

func (g *GitHubOAuthProvider) Name() string {
	return "github"
}

func (g *GitHubOAuthProvider) AuthURL(state string) string {
	params := url.Values{
		"client_id":    {g.clientID},
		"redirect_uri": {g.redirectURL},
		"scope":        {"read:user user:email"},
		"state":        {state},
	}
	return githubAuthURL + "?" + params.Encode()
}

func (g *GitHubOAuthProvider) Exchange(ctx context.Context, code string) (*domain.OAuthToken, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {g.clientID},
		"client_secret": {g.clientSecret},
		"redirect_uri":  {g.redirectURL},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed (status %d): %s", resp.StatusCode, body)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}
	if result.Error != "" {
		return nil, fmt.Errorf("token exchange error: %s", result.Error)
	}

	return &domain.OAuthToken{
		AccessToken: result.AccessToken,
	}, nil
}

func (g *GitHubOAuthProvider) UserInfo(ctx context.Context, token *domain.OAuthToken) (*domain.OAuthUserInfo, error) {
	// ユーザー基本情報を取得
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubUserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo request failed (status %d): %s", resp.StatusCode, body)
	}

	var info struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
		Email     string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	name := info.Name
	if name == "" {
		name = info.Login
	}

	email := info.Email
	if email == "" {
		// GitHub ではメールが非公開の場合があるため、emails API で取得
		email, err = g.fetchPrimaryEmail(ctx, token.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch email: %w", err)
		}
	}

	return &domain.OAuthUserInfo{
		ProviderID: fmt.Sprintf("%d", info.ID),
		Email:      email,
		Name:       name,
		AvatarURL:  info.AvatarURL,
	}, nil
}

// fetchPrimaryEmail は GitHub の emails API からプライマリメールを取得します。
func (g *GitHubOAuthProvider) fetchPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubEmailURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create email request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get emails: %w", err)
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("failed to decode emails: %w", err)
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	return "", fmt.Errorf("no verified primary email found")
}
