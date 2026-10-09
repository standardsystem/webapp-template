package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/handler"
	"github.com/your-org/webapp-template/internal/mock"
)

func newUserRepoWithAlice() *mock.UserRepository {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	repo := mock.NewUserRepository()
	repo.Users["u1"] = &domain.User{
		ID: "u1", Name: "Alice", Email: "alice@example.com",
		Role: domain.RoleMember, CreatedAt: now, UpdatedAt: now,
	}
	return repo
}

func TestUserHandler_List(t *testing.T) {
	tests := []struct {
		name       string
		findErr    error
		wantStatus int
		wantCount  int
	}{
		{"正常系: 一覧を返す", nil, http.StatusOK, 1},
		{"異常系: DB 障害は 500", errors.New("connection refused"), http.StatusInternalServerError, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newUserRepoWithAlice()
			repo.FindErr = tt.findErr
			h := handler.NewUserHandler(repo)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			h.Router().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantStatus != http.StatusOK {
				return
			}
			var body []map[string]any
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode: %v", err)
			}
			if len(body) != tt.wantCount {
				t.Errorf("len = %d, want %d", len(body), tt.wantCount)
			}
		})
	}
}

func TestUserHandler_Get(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		findErr    error
		wantStatus int
	}{
		{"正常系: ユーザーを返す", "u1", nil, http.StatusOK},
		{"異常系: 存在しないユーザーは 404", "missing", nil, http.StatusNotFound},
		{"異常系: DB 障害は 404 ではなく 500", "u1", errors.New("connection refused"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newUserRepoWithAlice()
			repo.FindErr = tt.findErr
			h := handler.NewUserHandler(repo)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.id, nil)
			rec := httptest.NewRecorder()
			h.Router().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantStatus != http.StatusOK {
				return
			}
			var body map[string]any
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode: %v", err)
			}
			if body["email"] != "alice@example.com" {
				t.Errorf("email = %v, want alice@example.com", body["email"])
			}
		})
	}
}
