package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/usecase"
)

// --- インメモリモック ---

type mockUserRepository struct {
	users  map[string]*domain.User
	saveErr error
	findErr error
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{users: make(map[string]*domain.User)}
}

func (m *mockUserRepository) FindByID(_ context.Context, id string) (*domain.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	u, ok := m.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepository) FindAll(_ context.Context) ([]*domain.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	result := make([]*domain.User, 0, len(m.users))
	for _, u := range m.users {
		result = append(result, u)
	}
	return result, nil
}

func (m *mockUserRepository) Save(_ context.Context, user *domain.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockUserRepository) Delete(_ context.Context, id string) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepository) Count(_ context.Context) (int64, error) {
	return int64(len(m.users)), nil
}

// --- テスト ---

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
			repo := newMockUserRepository()
			repo.saveErr = tt.saveErr
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
	repo := newMockUserRepository()
	repo.users["user-1"] = &domain.User{ID: "user-1", Name: "テスト", Email: "test@example.com"}
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
