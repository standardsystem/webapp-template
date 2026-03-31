package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/your-org/webapp-template/internal/domain"
)

// AuthUsecase は認証・認可のユースケースです。
type AuthUsecase struct {
	userRepo     domain.UserRepository
	providerRepo domain.UserProviderRepository
	sessionSvc   domain.SessionService
	providers    map[string]domain.OAuthProvider
}

// NewAuthUsecase は AuthUsecase を生成します。
func NewAuthUsecase(
	userRepo domain.UserRepository,
	providerRepo domain.UserProviderRepository,
	sessionSvc domain.SessionService,
	providers map[string]domain.OAuthProvider,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:     userRepo,
		providerRepo: providerRepo,
		sessionSvc:   sessionSvc,
		providers:    providers,
	}
}

// GetAuthURL は指定プロバイダの認可 URL と state を返します。
func (a *AuthUsecase) GetAuthURL(providerName string) (authURL, state string, err error) {
	provider, ok := a.providers[providerName]
	if !ok {
		return "", "", fmt.Errorf("%w: unknown provider: %s", domain.ErrInvalidInput, providerName)
	}

	state, err = generateState()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	return provider.AuthURL(state), state, nil
}

// AuthCallbackResult はコールバック処理の結果です。
type AuthCallbackResult struct {
	User       *domain.User
	SessionToken string
	IsNewUser  bool
}

// HandleCallback は OAuth コールバックを処理し、ユーザーの upsert とセッション発行を行います。
func (a *AuthUsecase) HandleCallback(ctx context.Context, providerName, code string) (*AuthCallbackResult, error) {
	provider, ok := a.providers[providerName]
	if !ok {
		return nil, fmt.Errorf("%w: unknown provider: %s", domain.ErrInvalidInput, providerName)
	}

	// 1. 認可コードをトークンに交換
	oauthToken, err := provider.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// 2. ユーザー情報を取得
	userInfo, err := provider.UserInfo(ctx, oauthToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// 3. プロバイダ紐付けを検索
	isNewUser := false
	up, err := a.providerRepo.FindByProviderAndProviderID(ctx, providerName, userInfo.ProviderID)

	var user *domain.User
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("failed to find provider link: %w", err)
		}

		// 4a. 新規ユーザー: メールで既存ユーザーを検索（別プロバイダで登録済みの可能性）
		user, err = a.userRepo.FindByEmail(ctx, userInfo.Email)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				return nil, fmt.Errorf("failed to find user by email: %w", err)
			}

			// 完全な新規ユーザーを作成
			role, err := a.determineRole(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to determine role: %w", err)
			}

			now := time.Now().UTC()
			user = &domain.User{
				ID:        uuid.New().String(),
				Name:      userInfo.Name,
				Email:     userInfo.Email,
				AvatarURL: userInfo.AvatarURL,
				Role:      role,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := a.userRepo.Save(ctx, user); err != nil {
				return nil, fmt.Errorf("failed to save new user: %w", err)
			}
			isNewUser = true
		}

		// プロバイダ紐付けを保存
		newUP := &domain.UserProvider{
			ID:         uuid.New().String(),
			UserID:     user.ID,
			Provider:   providerName,
			ProviderID: userInfo.ProviderID,
		}
		if err := a.providerRepo.Save(ctx, newUP); err != nil {
			return nil, fmt.Errorf("failed to save provider link: %w", err)
		}
	} else {
		// 4b. 既存ユーザー: 情報を更新
		user, err = a.userRepo.FindByID(ctx, up.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to find user: %w", err)
		}
		user.Name = userInfo.Name
		user.AvatarURL = userInfo.AvatarURL
		user.UpdatedAt = time.Now().UTC()
		if err := a.userRepo.Save(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	// 5. セッショントークンを発行
	sessionToken, err := a.sessionSvc.IssueToken(&domain.SessionClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to issue session: %w", err)
	}

	return &AuthCallbackResult{
		User:         user,
		SessionToken: sessionToken,
		IsNewUser:    isNewUser,
	}, nil
}

// GetCurrentUser は認証済みユーザーの情報を返します。
func (a *AuthUsecase) GetCurrentUser(ctx context.Context, userID string) (*domain.User, error) {
	user, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return user, nil
}

// UpdateUserRole はユーザーのロールを変更します（admin のみ実行可能）。
func (a *AuthUsecase) UpdateUserRole(ctx context.Context, targetUserID string, newRole domain.Role) error {
	if !newRole.IsValid() {
		return fmt.Errorf("%w: invalid role: %s", domain.ErrInvalidInput, newRole)
	}

	user, err := a.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	user.Role = newRole
	user.UpdatedAt = time.Now().UTC()
	if err := a.userRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	return nil
}

// determineRole は新規ユーザーのロールを決定します。
// 最初のユーザーは admin、以降は member。
func (a *AuthUsecase) determineRole(ctx context.Context) (domain.Role, error) {
	count, err := a.userRepo.Count(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to count users: %w", err)
	}
	if count == 0 {
		return domain.RoleAdmin, nil
	}
	return domain.RoleMember, nil
}

// generateState は CSRF 防止用のランダム state 文字列を生成します。
func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
