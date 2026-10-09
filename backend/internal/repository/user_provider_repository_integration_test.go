//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/repository"
)

func TestPostgresUserProviderRepository(t *testing.T) {
	pool := testPool(t)
	cleanUsers(t, pool)
	t.Cleanup(func() { cleanUsers(t, pool) })

	userRepo := repository.NewPostgresUserRepository(pool)
	repo := repository.NewPostgresUserProviderRepository(pool)
	ctx := context.Background()

	user := newTestUser(uuid.New().String(), "Alice", "alice@example.com")
	if err := userRepo.Save(ctx, user); err != nil {
		t.Fatalf("Save user failed: %v", err)
	}
	google := &domain.UserProvider{ID: uuid.New().String(), UserID: user.ID, Provider: "google", ProviderID: "google-1"}
	github := &domain.UserProvider{ID: uuid.New().String(), UserID: user.ID, Provider: "github", ProviderID: "github-1"}

	t.Run("正常系: 保存した紐付けをプロバイダと ID で取得", func(t *testing.T) {
		for _, up := range []*domain.UserProvider{google, github} {
			if err := repo.Save(ctx, up); err != nil {
				t.Fatalf("Save failed: %v", err)
			}
		}

		got, err := repo.FindByProviderAndProviderID(ctx, "google", "google-1")
		if err != nil {
			t.Fatalf("FindByProviderAndProviderID failed: %v", err)
		}
		if got.ID != google.ID || got.UserID != user.ID {
			t.Errorf("got %+v, want %+v", got, google)
		}
	})

	t.Run("異常系: 紐付けがなければ ErrNotFound", func(t *testing.T) {
		_, err := repo.FindByProviderAndProviderID(ctx, "google", "unknown")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("正常系: ユーザーの紐付けをすべて取得", func(t *testing.T) {
		got, err := repo.FindByUserID(ctx, user.ID)
		if err != nil {
			t.Fatalf("FindByUserID failed: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("len = %d, want 2", len(got))
		}
	})

	t.Run("正常系: 同じ (provider, provider_id) の保存は既存の紐付けを変えない", func(t *testing.T) {
		other := newTestUser(uuid.New().String(), "Bob", "bob@example.com")
		if err := userRepo.Save(ctx, other); err != nil {
			t.Fatalf("Save user failed: %v", err)
		}
		dup := &domain.UserProvider{ID: uuid.New().String(), UserID: other.ID, Provider: "google", ProviderID: "google-1"}
		if err := repo.Save(ctx, dup); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		got, err := repo.FindByProviderAndProviderID(ctx, "google", "google-1")
		if err != nil {
			t.Fatalf("FindByProviderAndProviderID failed: %v", err)
		}
		if got.UserID != user.ID {
			t.Errorf("UserID = %s, want %s (link must not move to another user)", got.UserID, user.ID)
		}
	})

	t.Run("異常系: 存在しないユーザーへの紐付けは保存できない", func(t *testing.T) {
		orphan := &domain.UserProvider{ID: uuid.New().String(), UserID: uuid.New().String(), Provider: "google", ProviderID: "google-orphan"}
		if err := repo.Save(ctx, orphan); err == nil {
			t.Error("expected foreign key error but got nil")
		}
	})

	t.Run("正常系: 紐付けを削除", func(t *testing.T) {
		if err := repo.Delete(ctx, github.ID); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		if _, err := repo.FindByProviderAndProviderID(ctx, "github", "github-1"); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("after delete: err = %v, want ErrNotFound", err)
		}
	})

	t.Run("正常系: ユーザーを削除すると紐付けも消える", func(t *testing.T) {
		if err := userRepo.Delete(ctx, user.ID); err != nil {
			t.Fatalf("Delete user failed: %v", err)
		}
		got, err := repo.FindByUserID(ctx, user.ID)
		if err != nil {
			t.Fatalf("FindByUserID failed: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("len = %d, want 0 (ON DELETE CASCADE)", len(got))
		}
	})
}
