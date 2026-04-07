package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/usecase"
)

// AuthService は AuthHandler が依存する認証ユースケースのインターフェースです。
type AuthService interface {
	GetAuthURL(providerName string) (authURL, state string, err error)
	HandleCallback(ctx context.Context, providerName, code string) (*usecase.AuthCallbackResult, error)
	GetCurrentUser(ctx context.Context, userID string) (*domain.User, error)
	UpdateUserRole(ctx context.Context, targetUserID string, newRole domain.Role) error
}

// AuthHandlerConfig は AuthHandler の設定です。
type AuthHandlerConfig struct {
	SecureCookie    bool
	FrontendOrigin  string
}

// AuthHandler は認証関連の HTTP ハンドラです。
type AuthHandler struct {
	authUsecase    AuthService
	secureCookie   bool
	frontendOrigin string
}

// NewAuthHandler は AuthHandler を生成します。
func NewAuthHandler(authUsecase AuthService, cfg AuthHandlerConfig) *AuthHandler {
	return &AuthHandler{
		authUsecase:    authUsecase,
		secureCookie:   cfg.SecureCookie,
		frontendOrigin: cfg.FrontendOrigin,
	}
}

// Router は認証関連のルーターを返します。
func (h *AuthHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/{provider}/login", h.handleLogin)
	r.Get("/{provider}/callback", h.handleCallback)
	return r
}

// handleLogin はプロバイダの認可画面にリダイレクトします。
func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")

	authURL, state, err := h.authUsecase.GetAuthURL(providerName)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// state を Cookie に保存（CSRF 防止）
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/api/v1/auth",
		MaxAge:   600, // 10 分
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// handleCallback は OAuth コールバックを処理します。
func (h *AuthHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")

	// state 検証（CSRF 防止）
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing state cookie"})
		return
	}
	if r.URL.Query().Get("state") != stateCookie.Value {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "state mismatch"})
		return
	}

	// state Cookie を削除
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	// OAuth エラーチェック
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "oauth error: " + errMsg})
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing code"})
		return
	}

	result, err := h.authUsecase.HandleCallback(r.Context(), providerName, code)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "authentication failed"})
		return
	}

	// セッション Cookie を設定
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    result.SessionToken,
		Path:     "/api",
		MaxAge:   86400, // 24 時間
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	// フロントエンドにリダイレクト
	http.Redirect(w, r, h.frontendOrigin, http.StatusTemporaryRedirect)
}

// handleLogout はログアウト処理を行います。
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/api",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

// HandleMe は認証済みユーザーの情報を返します（認証ミドルウェア適用後に使用）。
func (h *AuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	user, err := h.authUsecase.GetCurrentUser(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get user"})
		return
	}

	writeJSON(w, http.StatusOK, userResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		Role:      string(user.Role),
		CreatedAt: user.CreatedAt,
	})
}

// HandleUpdateRole はユーザーのロールを変更します（admin のみ）。
func (h *AuthHandler) HandleUpdateRole(w http.ResponseWriter, r *http.Request) {
	targetUserID := chi.URLParam(r, "id")

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.authUsecase.UpdateUserRole(r.Context(), targetUserID, domain.Role(req.Role)); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type userResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatarUrl"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}
