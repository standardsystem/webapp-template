package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/your-org/webapp-template/internal/domain"
)

// UserHandler はユーザー関連の HTTP ハンドラです。
type UserHandler struct {
	userRepo domain.UserRepository
}

// NewUserHandler は UserHandler を生成します。
func NewUserHandler(userRepo domain.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

// Router はユーザー関連のルーターを返します（認証必須エンドポイント用）。
func (h *UserHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.handleList)
	r.Get("/{id}", h.handleGet)
	return r
}

// handleList は全ユーザーの一覧を返します。
func (h *UserHandler) handleList(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.FindAll(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list users"})
		return
	}

	resp := make([]userListItem, 0, len(users))
	for _, u := range users {
		resp = append(resp, userListItem{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleGet は指定 ID のユーザーを返します。
func (h *UserHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.userRepo.FindByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, userListItem{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

type userListItem struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
