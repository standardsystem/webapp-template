package domain

import "context"

// --- OAuth プロバイダ抽象 ---

// OAuthUserInfo は外部プロバイダから取得したユーザー情報です。
type OAuthUserInfo struct {
	ProviderID string // プロバイダ側の一意な ID
	Email      string
	Name       string
	AvatarURL  string
}

// OAuthToken はプロバイダから取得したトークンです。
type OAuthToken struct {
	AccessToken  string
	RefreshToken string
}

// OAuthProvider は外部認証プロバイダの抽象インターフェースです。
// 新しいプロバイダを追加するには、このインターフェースを実装します。
type OAuthProvider interface {
	// Name はプロバイダ名を返します（"google", "github", "microsoft"）。
	Name() string
	// AuthURL は認可エンドポイントの URL を生成します。state は CSRF 防止トークンです。
	AuthURL(state string) string
	// Exchange は認可コードをトークンに交換します。
	Exchange(ctx context.Context, code string) (*OAuthToken, error)
	// UserInfo はトークンを使ってユーザー情報を取得します。
	UserInfo(ctx context.Context, token *OAuthToken) (*OAuthUserInfo, error)
}

// --- プロバイダ紐付けエンティティ ---

// UserProvider はユーザーと外部プロバイダの紐付けです。
type UserProvider struct {
	ID         string
	UserID     string
	Provider   string // "google", "github", "microsoft"
	ProviderID string // プロバイダ側のユーザー ID
}

// --- リポジトリインターフェース ---

// UserProviderRepository はユーザー・プロバイダ紐付けのデータアクセスインターフェースです。
type UserProviderRepository interface {
	FindByProviderAndProviderID(ctx context.Context, provider, providerID string) (*UserProvider, error)
	FindByUserID(ctx context.Context, userID string) ([]*UserProvider, error)
	Save(ctx context.Context, up *UserProvider) error
	Delete(ctx context.Context, id string) error
}

// --- セッションサービス抽象 ---

// SessionClaims は認証済みセッションのクレーム情報です。
type SessionClaims struct {
	UserID string
	Email  string
	Role   Role
}

// SessionService はセッションの発行と検証を担当するインターフェースです。
type SessionService interface {
	// IssueToken はセッション JWT を生成します。
	IssueToken(claims *SessionClaims) (string, error)
	// ValidateToken はセッション JWT を検証し、クレームを返します。
	ValidateToken(token string) (*SessionClaims, error)
}
