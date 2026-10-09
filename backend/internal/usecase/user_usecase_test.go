package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/mock"
	"github.com/your-org/webapp-template/internal/usecase"
)

func TestUserUsecase_GetUser(t *testing.T) {
	repo := mock.NewUserRepository()
	repo.Users["user-1"] = &domain.User{ID: "user-1", Name: "テスト", Email: "test@example.com"}
	uc := usecase.NewUserUsecase(repo)

	t.Run("正常系: 存在するユーザーを取得", func(t *testing.T) {
		got, err := uc.GetUser(context.Background(), "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != "user-1" {
			t.Errorf("ID = %s, want user-1", got.ID)
		}
	})

	t.Run("異常系: 存在しないユーザー", func(t *testing.T) {
		_, err := uc.GetUser(context.Background(), "not-exist")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("error = %v, want ErrNotFound", err)
		}
	})
}

func TestUserUsecase_ListUsers(t *testing.T) {
	dbErr := errors.New("db error")

	tests := []struct {
		name    string
		users   []*domain.User
		findErr error
		wantLen int
		wantErr error
	}{
		{
			name:    "正常系: ユーザーがいない",
			wantLen: 0,
		},
		{
			name: "正常系: 全ユーザーを取得",
			users: []*domain.User{
				{ID: "user-1", Name: "A", Email: "a@example.com"},
				{ID: "user-2", Name: "B", Email: "b@example.com"},
			},
			wantLen: 2,
		},
		{
			name:    "異常系: リポジトリエラーを原因として返す",
			findErr: dbErr,
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mock.NewUserRepository()
			for _, u := range tt.users {
				repo.Users[u.ID] = u
			}
			repo.FindErr = tt.findErr
			uc := usecase.NewUserUsecase(repo)

			got, err := uc.ListUsers(context.Background())

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Errorf("len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}
