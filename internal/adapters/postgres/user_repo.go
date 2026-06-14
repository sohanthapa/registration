package postgres

import (
	"context"
	"errors"

	"github.com/example/registration/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a Postgres-backed user store.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create inserts a user and returns ErrUserAlreadyExists on duplicate email.
func (r *UserRepository) Create(
	ctx context.Context,
	email string,
	passwordHash string,
) (domain.User, error) {
	var user domain.User

	query := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (email) DO NOTHING
		RETURNING id::text, email, password_hash, created_at
	`

	err := r.db.QueryRow(ctx, query, email, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	// returning no rows implies that nothing was inserted because there was a conflict with email.
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrUserAlreadyExists
	}

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

// FindByEmail loads a user by email.
func (r *UserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (domain.User, error) {
	var user domain.User

	query := `
		SELECT id::text, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}
