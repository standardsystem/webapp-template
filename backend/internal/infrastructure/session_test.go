package infrastructure_test

import (
	"testing"
	"time"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/infrastructure"
)

func TestJWTSessionService(t *testing.T) {
	svc := infrastructure.NewJWTSessionService("test-secret-key-at-least-32-bytes!", 1*time.Hour)

	t.Run("正常系: トークン発行と検証", func(t *testing.T) {
		claims := &domain.SessionClaims{
			UserID: "user-123",
			Email:  "test@example.com",
			Role:   domain.RoleMember,
		}

		token, err := svc.IssueToken(claims)
		if err != nil {
			t.Fatalf("IssueToken failed: %v", err)
		}
		if token == "" {
			t.Fatal("token should not be empty")
		}

		got, err := svc.ValidateToken(token)
		if err != nil {
			t.Fatalf("ValidateToken failed: %v", err)
		}
		if got.UserID != claims.UserID {
			t.Errorf("UserID = %s, want %s", got.UserID, claims.UserID)
		}
		if got.Email != claims.Email {
			t.Errorf("Email = %s, want %s", got.Email, claims.Email)
		}
		if got.Role != claims.Role {
			t.Errorf("Role = %s, want %s", got.Role, claims.Role)
		}
	})

	t.Run("異常系: 不正なトークン", func(t *testing.T) {
		_, err := svc.ValidateToken("invalid-token")
		if err == nil {
			t.Error("expected error for invalid token")
		}
	})

	t.Run("異常系: 期限切れトークン", func(t *testing.T) {
		expiredSvc := infrastructure.NewJWTSessionService("test-secret-key-at-least-32-bytes!", -1*time.Hour)

		claims := &domain.SessionClaims{
			UserID: "user-123",
			Email:  "test@example.com",
			Role:   domain.RoleMember,
		}

		token, err := expiredSvc.IssueToken(claims)
		if err != nil {
			t.Fatalf("IssueToken failed: %v", err)
		}

		_, err = expiredSvc.ValidateToken(token)
		if err == nil {
			t.Error("expected error for expired token")
		}
	})

	t.Run("異常系: 異なるシークレットで検証", func(t *testing.T) {
		otherSvc := infrastructure.NewJWTSessionService("different-secret-key-at-least-32!!", 1*time.Hour)

		claims := &domain.SessionClaims{
			UserID: "user-123",
			Email:  "test@example.com",
			Role:   domain.RoleMember,
		}

		token, err := svc.IssueToken(claims)
		if err != nil {
			t.Fatalf("IssueToken failed: %v", err)
		}

		_, err = otherSvc.ValidateToken(token)
		if err == nil {
			t.Error("expected error for different secret")
		}
	})
}
