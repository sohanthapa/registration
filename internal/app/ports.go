package app

import (
	"context"
	"time"

	"github.com/example/registration/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, email string, passwordHash string) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
}

type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Compare(passwordHash string, plainPassword string) bool
}

type TokenIssuer interface {
	Issue(ctx context.Context, user domain.User) (Token, error)
}

type Token struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int64
	ExpiresAt   time.Time
}
