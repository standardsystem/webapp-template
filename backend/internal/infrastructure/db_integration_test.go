//go:build integration

package infrastructure_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/your-org/webapp-template/internal/infrastructure"
)

func TestNewDB(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("正常系: 接続して ping が通る", func(t *testing.T) {
		pool, err := infrastructure.NewDB(ctx, dsn)
		if err != nil {
			t.Fatalf("NewDB failed: %v", err)
		}
		defer pool.Close()

		var one int
		if err := pool.QueryRow(ctx, "SELECT 1").Scan(&one); err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if one != 1 {
			t.Errorf("SELECT 1 = %d", one)
		}
	})

	t.Run("異常系: 接続文字列が不正", func(t *testing.T) {
		if _, err := infrastructure.NewDB(ctx, "not-a-url://"); err == nil {
			t.Error("expected error for invalid URL")
		}
	})

	t.Run("異常系: 接続先に到達できない", func(t *testing.T) {
		// ポート 1 では PostgreSQL が待ち受けていない
		if _, err := infrastructure.NewDB(ctx, "postgres://postgres:password@127.0.0.1:1/webapp?sslmode=disable&connect_timeout=2"); err == nil {
			t.Error("expected error for unreachable host")
		}
	})
}
