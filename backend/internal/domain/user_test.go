package domain_test

import (
	"errors"
	"testing"

	"github.com/your-org/webapp-template/internal/domain"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    domain.User
		wantErr bool
	}{
		{"正常系: member", domain.User{Name: "テスト", Email: "test@example.com", Role: domain.RoleMember}, false},
		{"正常系: admin", domain.User{Name: "テスト", Email: "test@example.com", Role: domain.RoleAdmin}, false},
		{"異常系: 名前が空", domain.User{Name: "", Email: "test@example.com", Role: domain.RoleMember}, true},
		{"異常系: メールが空", domain.User{Name: "テスト", Email: "", Role: domain.RoleMember}, true},
		{"異常系: ロールが不正", domain.User{Name: "テスト", Email: "test@example.com", Role: "owner"}, true},
		{"異常系: ロールが空", domain.User{Name: "テスト", Email: "test@example.com"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()

			if !tt.wantErr {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if !errors.Is(err, domain.ErrInvalidInput) {
				t.Errorf("error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name string
		role domain.Role
		want bool
	}{
		{"admin", domain.RoleAdmin, true},
		{"member", domain.RoleMember, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := domain.User{Role: tt.role}
			if got := u.IsAdmin(); got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}
