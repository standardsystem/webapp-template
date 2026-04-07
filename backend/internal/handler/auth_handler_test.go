package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/handler"
	"github.com/your-org/webapp-template/internal/mock"
	"github.com/your-org/webapp-template/internal/usecase"
)

func newTestAuthHandler(svc *mock.AuthService) *handler.AuthHandler {
	return handler.NewAuthHandler(svc, handler.AuthHandlerConfig{
		SecureCookie:   false,
		FrontendOrigin: "http://localhost:5173",
	})
}

// --- handleLogin テスト ---

func TestAuthHandler_Login(t *testing.T) {
	t.Run("正常系: 認可URLへリダイレクト", func(t *testing.T) {
		svc := &mock.AuthService{
			AuthURL_: "https://accounts.google.com/auth?foo=bar",
			State_:   "random-state",
		}
		h := newTestAuthHandler(svc)

		r := chi.NewRouter()
		r.Mount("/auth", h.Router())

		req := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusTemporaryRedirect {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusTemporaryRedirect)
		}
		loc := rec.Header().Get("Location")
		if loc != "https://accounts.google.com/auth?foo=bar" {
			t.Errorf("Location = %s, want auth URL", loc)
		}
		// state cookie が設定されていることを確認
		cookies := rec.Result().Cookies()
		var found bool
		for _, c := range cookies {
			if c.Name == "oauth_state" && c.Value == "random-state" {
				found = true
				break
			}
		}
		if !found {
			t.Error("oauth_state cookie not set")
		}
	})

	t.Run("異常系: 未知のプロバイダ", func(t *testing.T) {
		svc := &mock.AuthService{
			GetAuthURLErr: errors.New("unknown provider"),
		}
		h := newTestAuthHandler(svc)

		r := chi.NewRouter()
		r.Mount("/auth", h.Router())

		req := httptest.NewRequest(http.MethodGet, "/auth/unknown/login", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

// --- handleCallback テスト ---

func TestAuthHandler_Callback(t *testing.T) {
	t.Run("正常系: コールバック成功でフロントエンドにリダイレクト", func(t *testing.T) {
		svc := &mock.AuthService{
			CallbackResult: &usecase.AuthCallbackResult{
				User:         &domain.User{ID: "u1", Name: "Test", Email: "t@e.com", Role: domain.RoleMember},
				SessionToken: "jwt-token-123",
				IsNewUser:    true,
			},
		}
		h := newTestAuthHandler(svc)

		r := chi.NewRouter()
		r.Mount("/auth", h.Router())

		req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=authcode&state=valid-state", nil)
		req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "valid-state"})
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusTemporaryRedirect {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusTemporaryRedirect)
		}
		loc := rec.Header().Get("Location")
		if loc != "http://localhost:5173" {
			t.Errorf("Location = %s, want frontend origin", loc)
		}
		// session_token cookie が設定されていることを確認
		var sessionCookie *http.Cookie
		for _, c := range rec.Result().Cookies() {
			if c.Name == "session_token" {
				sessionCookie = c
				break
			}
		}
		if sessionCookie == nil {
			t.Fatal("session_token cookie not set")
		}
		if sessionCookie.Value != "jwt-token-123" {
			t.Errorf("session_token = %s, want jwt-token-123", sessionCookie.Value)
		}
		if !sessionCookie.HttpOnly {
			t.Error("session_token should be HttpOnly")
		}
	})

	t.Run("異常系: state Cookie なし", func(t *testing.T) {
		h := newTestAuthHandler(&mock.AuthService{})

		r := chi.NewRouter()
		r.Mount("/auth", h.Router())

		req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=authcode&state=s", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: state 不一致", func(t *testing.T) {
		h := newTestAuthHandler(&mock.AuthService{})

		r := chi.NewRouter()
		r.Mount("/auth", h.Router())

		req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=authcode&state=wrong", nil)
		req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "expected"})
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: code パラメータなし", func(t *testing.T) {
		h := newTestAuthHandler(&mock.AuthService{})

		r := chi.NewRouter()
		r.Mount("/auth", h.Router())

		req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state=s", nil)
		req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "s"})
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: OAuth エラーレスポンス", func(t *testing.T) {
		h := newTestAuthHandler(&mock.AuthService{})

		r := chi.NewRouter()
		r.Mount("/auth", h.Router())

		req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?error=access_denied&state=s", nil)
		req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "s"})
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: ユースケースエラー", func(t *testing.T) {
		svc := &mock.AuthService{
			CallbackErr: errors.New("exchange failed"),
		}
		h := newTestAuthHandler(svc)

		r := chi.NewRouter()
		r.Mount("/auth", h.Router())

		req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=authcode&state=s", nil)
		req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "s"})
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})
}

// --- HandleLogout テスト ---

func TestAuthHandler_Logout(t *testing.T) {
	h := newTestAuthHandler(&mock.AuthService{})

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	rec := httptest.NewRecorder()
	h.HandleLogout(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	// session_token Cookie がクリアされていることを確認
	for _, c := range rec.Result().Cookies() {
		if c.Name == "session_token" && c.MaxAge != -1 {
			t.Errorf("session_token MaxAge = %d, want -1", c.MaxAge)
		}
	}
}

// --- HandleMe テスト ---

func TestAuthHandler_Me(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("正常系: ユーザー情報を取得", func(t *testing.T) {
		svc := &mock.AuthService{
			CurrentUser: &domain.User{
				ID: "u1", Name: "Test", Email: "t@e.com",
				AvatarURL: "https://example.com/a.png", Role: domain.RoleMember,
				CreatedAt: now, UpdatedAt: now,
			},
		}
		h := newTestAuthHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		ctx := handler.ContextWithUser(req.Context(), "u1", "t@e.com", domain.RoleMember)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		h.HandleMe(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		var body map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if body["name"] != "Test" {
			t.Errorf("name = %v, want Test", body["name"])
		}
		if body["role"] != "member" {
			t.Errorf("role = %v, want member", body["role"])
		}
	})

	t.Run("異常系: コンテキストに userID がない", func(t *testing.T) {
		h := newTestAuthHandler(&mock.AuthService{})

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		rec := httptest.NewRecorder()
		h.HandleMe(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("異常系: ユーザー取得失敗", func(t *testing.T) {
		svc := &mock.AuthService{
			CurrentUserErr: errors.New("not found"),
		}
		h := newTestAuthHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		ctx := handler.ContextWithUser(req.Context(), "u1", "t@e.com", domain.RoleMember)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		h.HandleMe(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})
}

// --- HandleUpdateRole テスト ---

func TestAuthHandler_UpdateRole(t *testing.T) {
	t.Run("正常系: ロール更新成功", func(t *testing.T) {
		svc := &mock.AuthService{}
		h := newTestAuthHandler(svc)

		r := chi.NewRouter()
		r.Put("/users/{id}/role", h.HandleUpdateRole)

		body := strings.NewReader(`{"role":"admin"}`)
		req := httptest.NewRequest(http.MethodPut, "/users/u1/role", body)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	})

	t.Run("異常系: 不正な JSON", func(t *testing.T) {
		h := newTestAuthHandler(&mock.AuthService{})

		r := chi.NewRouter()
		r.Put("/users/{id}/role", h.HandleUpdateRole)

		body := strings.NewReader(`invalid`)
		req := httptest.NewRequest(http.MethodPut, "/users/u1/role", body)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: ユースケースエラー", func(t *testing.T) {
		svc := &mock.AuthService{
			UpdateRoleErr: errors.New("invalid role"),
		}
		h := newTestAuthHandler(svc)

		r := chi.NewRouter()
		r.Put("/users/{id}/role", h.HandleUpdateRole)

		body := strings.NewReader(`{"role":"superadmin"}`)
		req := httptest.NewRequest(http.MethodPut, "/users/u1/role", body)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

