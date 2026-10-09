package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/your-org/webapp-template/internal/domain"
)

// PostgresUserRepository は PostgreSQL による UserRepository 実装です。
type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresUserRepository は PostgresUserRepository を生成します。
func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, avatar_url, role, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.AvatarURL, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		// id が UUID の形式でないときも「存在しない」として扱う。
		// URL のパスから渡された不正な ID を、DB 障害（500）と区別するため。
		if errors.Is(err, pgx.ErrNoRows) || isInvalidTextRepresentation(err) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	return &u, nil
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, avatar_url, role, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.AvatarURL, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return &u, nil
}

func (r *PostgresUserRepository) FindAll(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, email, avatar_url, role, created_at, updated_at
		 FROM users ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.AvatarURL, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}

func (r *PostgresUserRepository) Save(ctx context.Context, user *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, name, email, avatar_url, role, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (id) DO UPDATE SET
		   name = EXCLUDED.name,
		   email = EXCLUDED.email,
		   avatar_url = EXCLUDED.avatar_url,
		   role = EXCLUDED.role,
		   updated_at = EXCLUDED.updated_at`,
		user.ID, user.Name, user.Email, user.AvatarURL, user.Role, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) CreateWithProvider(ctx context.Context, user *domain.User, up *domain.UserProvider) error {
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`INSERT INTO users (id, name, email, avatar_url, role, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			user.ID, user.Name, user.Email, user.AvatarURL, user.Role, user.CreatedAt, user.UpdatedAt,
		); err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}
		// 紐付けが重複したときに黙って成功させると、紐付けのないユーザーが残る。
		// ON CONFLICT を付けず、エラーにしてユーザーの作成ごとロールバックする。
		if _, err := tx.Exec(ctx,
			`INSERT INTO user_providers (id, user_id, provider, provider_id)
			 VALUES ($1, $2, $3, $4)`,
			up.ID, up.UserID, up.Provider, up.ProviderID,
		); err != nil {
			return fmt.Errorf("failed to insert user_provider: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create user with provider: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

// isInvalidTextRepresentation は、値を列の型（UUID など）に変換できなかったエラーかどうかを返します。
// SQLSTATE 22P02 (invalid_text_representation)。
func isInvalidTextRepresentation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "22P02"
}
