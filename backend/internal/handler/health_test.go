package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/webapp-template/internal/handler"
)

func TestHealth(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		wantStatusCode int
		wantStatus     string
	}{
		{
			name:           "正常系: GETリクエストで200を返す",
			method:         http.MethodGet,
			wantStatusCode: http.StatusOK,
			wantStatus:     "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()

			handler.Health(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("StatusCode = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}

			if resp.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Content-Type = %s, want application/json", resp.Header.Get("Content-Type"))
			}

			var body map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if body["status"] != tt.wantStatus {
				t.Errorf("status = %v, want %s", body["status"], tt.wantStatus)
			}

			if _, ok := body["timestamp"]; !ok {
				t.Error("timestamp field is missing")
			}
		})
	}
}
