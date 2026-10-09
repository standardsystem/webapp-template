package handler_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/handler"
	"github.com/your-org/webapp-template/internal/mock"
)

func testUserAlice() *domain.User {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return &domain.User{
		ID: "u1", Name: "Alice", Email: "alice@example.com",
		Role: domain.RoleMember, CreatedAt: now, UpdatedAt: now,
	}
}

func TestUserHandler_List(t *testing.T) {
	tests := []struct {
		name       string
		svc        *mock.UserService
		wantStatus int
		wantCount  int
	}{
		{"正常系: 一覧を返す", &mock.UserService{Users: []*domain.User{testUserAlice()}}, http.StatusOK, 1},
		{"正常系: ユーザーがいなければ空配列", &mock.UserService{}, http.StatusOK, 0},
		{"異常系: DB 障害は 500", &mock.UserService{ListErr: errors.New("connection refused")}, http.StatusInternalServerError, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handler.NewUserHandler(tt.svc)

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
			if body == nil {
				t.Error("body = null, want JSON array")
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
		svc        *mock.UserService
		wantStatus int
	}{
		{"正常系: ユーザーを返す", &mock.UserService{User: testUserAlice()}, http.StatusOK},
		{"異常系: 存在しないユーザーは 404", &mock.UserService{GetErr: fmt.Errorf("failed to find user: %w", domain.ErrNotFound)}, http.StatusNotFound},
		{"異常系: DB 障害は 404 ではなく 500", &mock.UserService{GetErr: errors.New("connection refused")}, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handler.NewUserHandler(tt.svc)

			req := httptest.NewRequest(http.MethodGet, "/u1", nil)
			rec := httptest.NewRecorder()
			h.Router().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if len(tt.svc.GotIDs) != 1 || tt.svc.GotIDs[0] != "u1" {
				t.Errorf("GetUser called with %v, want [u1]", tt.svc.GotIDs)
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
