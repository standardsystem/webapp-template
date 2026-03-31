package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

type healthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// Health は Cloud Run のヘルスチェックエンドポイント
func Health(w http.ResponseWriter, r *http.Request) {
	resp := healthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
