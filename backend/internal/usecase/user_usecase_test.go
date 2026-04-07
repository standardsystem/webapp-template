package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/mock"
	"github.com/your-org/webapp-template/internal/usecase"
)

func TestUserUsecase_CreateUser(t *testing.T) {
	tests := []struct {
		name    string
		input   usecase.CreateUserInput
		saveErr error
		wantErr bool
	}{
		{
			name:  "正常系: ユーザーが作成される",
			input: usecase.CreateUserInput{Name: "加藤一由樹", Email: "kazuyuki@example.com"},
		},
		{
			name:    "異常系: 名前が空",
			input:   usecase.CreateUserInput{Name: "", Email: "test@example.com"},
			wantErr: true,
		},
		{
			name:    "異常系: メールが空",
			input:   usecase.CreateUserInput{Name: "テスト", Email: ""},
			wantErr: true,
		},
		{
			name:    "異常系: リポジトリエラー",
			input:   usecase.CreateUserInput{Name: "テスト", Email: "test@example.com"},
			saveErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mock.NewUserRepository()
			repo.SaveErr = tt.saveErr
			uc := usecase.NewUserUsecase(repo)

			got, err := uc.CreateUser(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Name != tt.input.Name {
				t.Errorf("Name = %s, want %s", got.Name, tt.input.Name)
			}
			if got.Email != tt.input.Email {
				t.Errorf("Email = %s, want %s", got.Email, tt.input.Email)
			}
		})
	}
}

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
