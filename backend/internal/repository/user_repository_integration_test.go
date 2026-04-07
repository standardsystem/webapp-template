//go:build integration

package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/repository"
)

// testPool はテスト用の DB コネクションプールを返します。
// 環境変数 TEST_DATABASE_URL が必要です。
//
// 使い方:
//
//	TEST_DATABASE_URL=postgres://user:pass@localhost:5432/testdb go test -tags=integration ./internal/repository/...
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	t.Cleanup(func() { pool.Close() })
	return pool
}

// cleanUsers はテスト前後に users テーブルをクリーンアップします。
func cleanUsers(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	// user_providers は ON DELETE CASCADE なので users だけ削除すれば十分
	if _, err := pool.Exec(ctx, "DELETE FROM users"); err != nil {
		t.Fatalf("failed to clean users table: %v", err)
	}
}

func newTestUser(id, name, email string) *domain.User {
	now := time.Now().Truncate(time.Microsecond) // PostgreSQL の精度に合わせる
	return &domain.User{
		ID:        id,
		Name:      name,
		Email:     email,
		AvatarURL: fmt.Sprintf("https://example.com/%s.png", id),
		Role:      domain.RoleMember,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestPostgresUserRepository_Save_and_FindByID(t *testing.T) {
	pool := testPool(t)
	cleanUsers(t, pool)
	t.Cleanup(func() { cleanUsers(t, pool) })

	repo := repository.NewPostgresUserRepository(pool)
	ctx := context.Background()

	user := newTestUser("test-id-1", "Alice", "alice@example.com")

	if err := repo.Save(ctx, user); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := repo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if got.ID != user.ID || got.Name != user.Name || got.Email != user.Email {
		t.Errorf("got %+v, want %+v", got, user)
	}
	if got.Role != domain.RoleMember {
		t.Errorf("role = %s, want member", got.Role)
	}
}

func TestPostgresUserRepository_FindByEmail(t *testing.T) {
	pool := testPool(t)
	cleanUsers(t, pool)
	t.Cleanup(func() { cleanUsers(t, pool) })

	repo := repository.NewPostgresUserRepository(pool)
	ctx := context.Background()

	user := newTestUser("test-id-2", "Bob", "bob@example.com")
	if err := repo.Save(ctx, user); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := repo.FindByEmail(ctx, "bob@example.com")
	if err != nil {
		t.Fatalf("FindByEmail failed: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("ID = %s, want %s", got.ID, user.ID)
	}
}

func TestPostgresUserRepository_FindByID_NotFound(t *testing.T) {
	pool := testPool(t)
	cleanUsers(t, pool)
	t.Cleanup(func() { cleanUsers(t, pool) })

	repo := repository.NewPostgresUserRepository(pool)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "nonexistent-id")
	if err != domain.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestPostgresUserRepository_FindAll(t *testing.T) {
	pool := testPool(t)
	cleanUsers(t, pool)
	t.Cleanup(func() { cleanUsers(t, pool) })

	repo := repository.NewPostgresUserRepository(pool)
	ctx := context.Background()

	users := []*domain.User{
		newTestUser("test-id-a", "Alice", "alice@example.com"),
		newTestUser("test-id-b", "Bob", "bob@example.com"),
	}
	for _, u := range users {
		if err := repo.Save(ctx, u); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	got, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestPostgresUserRepository_Save_Upsert(t *testing.T) {
	pool := testPool(t)
	cleanUsers(t, pool)
	t.Cleanup(func() { cleanUsers(t, pool) })

	repo := repository.NewPostgresUserRepository(pool)
	ctx := context.Background()

	user := newTestUser("test-id-upsert", "Original", "upsert@example.com")
	if err := repo.Save(ctx, user); err != nil {
		t.Fatalf("initial Save failed: %v", err)
	}

	// 同じ ID で更新
	user.Name = "Updated"
	user.UpdatedAt = time.Now().Truncate(time.Microsecond)
	if err := repo.Save(ctx, user); err != nil {
		t.Fatalf("upsert Save failed: %v", err)
	}

	got, err := repo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if got.Name != "Updated" {
		t.Errorf("name = %s, want Updated", got.Name)
	}

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1 (upsert should not duplicate)", count)
	}
}

func TestPostgresUserRepository_Delete(t *testing.T) {
	pool := testPool(t)
	cleanUsers(t, pool)
	t.Cleanup(func() { cleanUsers(t, pool) })

	repo := repository.NewPostgresUserRepository(pool)
	ctx := context.Background()

	user := newTestUser("test-id-del", "ToDelete", "delete@example.com")
	if err := repo.Save(ctx, user); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if err := repo.Delete(ctx, user.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := repo.FindByID(ctx, user.ID)
	if err != domain.ErrNotFound {
		t.Errorf("after delete: err = %v, want ErrNotFound", err)
	}
}

func TestPostgresUserRepository_Count(t *testing.T) {
	pool := testPool(t)
	cleanUsers(t, pool)
	t.Cleanup(func() { cleanUsers(t, pool) })

	repo := repository.NewPostgresUserRepository(pool)
	ctx := context.Background()

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Errorf("initial count = %d, want 0", count)
	}

	if err := repo.Save(ctx, newTestUser("c1", "A", "a@e.com")); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if err := repo.Save(ctx, newTestUser("c2", "B", "b@e.com")); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	count, err = repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}
