package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/your-org/webapp-template/internal/domain"
)

// captureLogs は slog の既定ロガーを差し替え、テスト中のログ出力をためるバッファを返します。
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantMessage string
		wantLogged  bool
	}{
		{"ErrNotFound は 404", fmt.Errorf("failed to find user: %w", domain.ErrNotFound), http.StatusNotFound, "not found", false},
		{"ErrInvalidInput は 400", fmt.Errorf("%w: invalid role: x", domain.ErrInvalidInput), http.StatusBadRequest, "invalid input", false},
		{"ErrUnauthorized は 401", domain.ErrUnauthorized, http.StatusUnauthorized, "unauthorized", false},
		{"ErrForbidden は 403", domain.ErrForbidden, http.StatusForbidden, "forbidden", false},
		{"ErrEmailNotVerified は 403", fmt.Errorf("%w: provider google", domain.ErrEmailNotVerified), http.StatusForbidden, "email not verified by provider", false},
		{"ErrEmailDomainNotAllowed は 403", fmt.Errorf("%w: provider google", domain.ErrEmailDomainNotAllowed), http.StatusForbidden, "email domain not allowed", false},
		{"ErrAlreadyExists は 409", domain.ErrAlreadyExists, http.StatusConflict, "already exists", false},
		{"ErrAccountLinkRequired は 409", domain.ErrAccountLinkRequired, http.StatusConflict, "email already registered with another provider", false},
		{"それ以外は 500 でログに残す", errors.New("pq: connection refused"), http.StatusInternalServerError, "internal server error", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs := captureLogs(t)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/u1", nil)
			req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "req-123"))
			rec := httptest.NewRecorder()

			writeError(rec, req, tt.err)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			var body map[string]string
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode body: %v", err)
			}
			if body["error"] != tt.wantMessage {
				t.Errorf("error = %q, want %q", body["error"], tt.wantMessage)
			}

			logged := logs.String()
			if !tt.wantLogged {
				if logged != "" {
					t.Errorf("unexpected log output: %s", logged)
				}
				return
			}
			for _, want := range []string{`"level":"ERROR"`, "pq: connection refused", `"request_id":"req-123"`, `"path":"/api/v1/users/u1"`} {
				if !strings.Contains(logged, want) {
					t.Errorf("log output does not contain %s: %s", want, logged)
				}
			}
		})
	}
}
