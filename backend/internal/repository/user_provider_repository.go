package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/your-org/webapp-template/internal/domain"
)

// PostgresUserProviderRepository は PostgreSQL による UserProviderRepository 実装です。
type PostgresUserProviderRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresUserProviderRepository は PostgresUserProviderRepository を生成します。
func NewPostgresUserProviderRepository(pool *pgxpool.Pool) *PostgresUserProviderRepository {
	return &PostgresUserProviderRepository{pool: pool}
}

func (r *PostgresUserProviderRepository) FindByProviderAndProviderID(ctx context.Context, provider, providerID string) (*domain.UserProvider, error) {
	var up domain.UserProvider
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, provider, provider_id
		 FROM user_providers WHERE provider = $1 AND provider_id = $2`,
		provider, providerID,
	).Scan(&up.ID, &up.UserID, &up.Provider, &up.ProviderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user_provider: %w", err)
	}
	return &up, nil
}

func (r *PostgresUserProviderRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.UserProvider, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, provider, provider_id
		 FROM user_providers WHERE user_id = $1`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query user_providers: %w", err)
	}
	defer rows.Close()

	var providers []*domain.UserProvider
	for rows.Next() {
		var up domain.UserProvider
		if err := rows.Scan(&up.ID, &up.UserID, &up.Provider, &up.ProviderID); err != nil {
			return nil, fmt.Errorf("failed to scan user_provider: %w", err)
		}
		providers = append(providers, &up)
	}
	return providers, rows.Err()
}

func (r *PostgresUserProviderRepository) Save(ctx context.Context, up *domain.UserProvider) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_providers (id, user_id, provider, provider_id)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (provider, provider_id) DO NOTHING`,
		up.ID, up.UserID, up.Provider, up.ProviderID,
	)
	if err != nil {
		return fmt.Errorf("failed to save user_provider: %w", err)
	}
	return nil
}

func (r *PostgresUserProviderRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_providers WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete user_provider: %w", err)
	}
	return nil
}
