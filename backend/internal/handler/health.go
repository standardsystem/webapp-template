package handler

import (
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
	writeJSON(w, http.StatusOK, resp)
}
