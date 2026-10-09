package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/your-org/webapp-template/internal/domain"
)

// writeJSON は JSON レスポンスを書き込むヘルパーです。
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// ステータスコードは送信済みで、呼び出し元へ返しても対応できない。原因をログに残す。
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write response body", "err", err, "status", status)
	}
}

// errorResponses は domain のエラーと、利用者へ返すステータスコード・本文の対応です。
// 本文は固定の文言にして、内部のエラー文言を利用者へ見せません。
var errorResponses = []struct {
	target  error
	status  int
	message string
}{
	{domain.ErrNotFound, http.StatusNotFound, "not found"},
	{domain.ErrInvalidInput, http.StatusBadRequest, "invalid input"},
	{domain.ErrUnauthorized, http.StatusUnauthorized, "unauthorized"},
	{domain.ErrForbidden, http.StatusForbidden, "forbidden"},
	{domain.ErrEmailNotVerified, http.StatusForbidden, "email not verified by provider"},
	{domain.ErrEmailDomainNotAllowed, http.StatusForbidden, "email domain not allowed"},
	{domain.ErrAlreadyExists, http.StatusConflict, "already exists"},
	{domain.ErrAccountLinkRequired, http.StatusConflict, "email already registered with another provider"},
}

// writeError は、ユースケースが返したエラーを HTTP レスポンスに変換して書き込みます。
// domain のエラーに当てはまらないものは 500 として扱い、原因をログに残します。
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	for _, e := range errorResponses {
		if errors.Is(err, e.target) {
			writeJSON(w, e.status, map[string]string{"error": e.message})
			return
		}
	}

	slog.ErrorContext(r.Context(), "request failed",
		"err", err,
		"request_id", middleware.GetReqID(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
	)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}
