package handler

import (
	"context"
	"net/http"

	"github.com/your-org/webapp-template/internal/domain"
)

type contextKey string

const (
	ctxUserID contextKey = "userID"
	ctxEmail  contextKey = "email"
	ctxRole   contextKey = "role"
)

// AuthMiddleware は JWT セッションを検証する認証ミドルウェアです。
type AuthMiddleware struct {
	sessionSvc domain.SessionService
}

// NewAuthMiddleware は AuthMiddleware を生成します。
func NewAuthMiddleware(sessionSvc domain.SessionService) *AuthMiddleware {
	return &AuthMiddleware{sessionSvc: sessionSvc}
}

// Handler は認証ミドルウェアを返します。
func (m *AuthMiddleware) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
				return
			}

			claims, err := m.sessionSvc.ValidateToken(cookie.Value)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired session"})
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, ctxUserID, claims.UserID)
			ctx = context.WithValue(ctx, ctxEmail, claims.Email)
			ctx = context.WithValue(ctx, ctxRole, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole は指定ロール以上の権限を要求するミドルウェアです。
func RequireRole(required domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := RoleFromContext(r.Context())

			if required == domain.RoleAdmin && role != domain.RoleAdmin {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin access required"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// UserIDFromContext はコンテキストからユーザー ID を取得します。
func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxUserID).(string)
	return v
}

// RoleFromContext はコンテキストからロールを取得します。
func RoleFromContext(ctx context.Context) domain.Role {
	v, _ := ctx.Value(ctxRole).(domain.Role)
	return v
}

// ContextWithUser はテスト用にユーザー情報をコンテキストに設定するヘルパーです。
func ContextWithUser(ctx context.Context, userID, email string, role domain.Role) context.Context {
	ctx = context.WithValue(ctx, ctxUserID, userID)
	ctx = context.WithValue(ctx, ctxEmail, email)
	ctx = context.WithValue(ctx, ctxRole, role)
	return ctx
}
