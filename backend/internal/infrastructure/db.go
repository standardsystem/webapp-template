package infrastructure

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDB は PostgreSQL コネクションプールを作成します。
func NewDB(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return pool, nil
}

// MigrationFile はマイグレーションファイルを表します。
type MigrationFile struct {
	Name    string
	Content string
}

// RunMigrations はマイグレーションを実行します。
// migrations は名前順にソートされたマイグレーションファイルのリストです。
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrations []MigrationFile) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name < migrations[j].Name
	})

	for _, m := range migrations {
		if !strings.HasSuffix(m.Name, ".sql") {
			continue
		}

		var exists bool
		err := pool.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)",
			m.Name,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", m.Name, err)
		}
		if exists {
			continue
		}

		_, err = pool.Exec(ctx, m.Content)
		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", m.Name, err)
		}

		_, err = pool.Exec(ctx,
			"INSERT INTO schema_migrations (version) VALUES ($1)",
			m.Name,
		)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", m.Name, err)
		}
	}

	return nil
}
