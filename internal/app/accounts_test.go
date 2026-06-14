package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/registration/internal/domain"
)

// TestSignUpCreatesUserWithNormalizedEmailAndHashedPassword checks signup stores a normalized email and hashed password.
func TestSignUpCreatesUserWithNormalizedEmailAndHashedPassword(t *testing.T) {
	repo := newFakeUserRepository()
	service := NewService(repo, fakePasswordHasher{}, fakeTokenIssuer{})

	user, err := service.SignUp(context.Background(), Credentials{
		Email:    "  USER@example.COM ",
		Password: "supersecret123",
	})
	if err != nil {
		t.Fatalf("SignUp returned error: %v", err)
	}

	if user.Email != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", user.Email)
	}

	stored := repo.usersByEmail["user@example.com"]
	if stored.PasswordHash != "hash:supersecret123" {
		t.Fatalf("expected hashed password, got %q", stored.PasswordHash)
	}
}

// TestSignUpRejectsDuplicateEmail checks signup fails when the email is already taken.
func TestSignUpRejectsDuplicateEmail(t *testing.T) {
	repo := newFakeUserRepository()
	service := NewService(repo, fakePasswordHasher{}, fakeTokenIssuer{})

	_, err := service.SignUp(context.Background(), Credentials{
		Email:    "user@example.com",
		Password: "supersecret123",
	})
	if err != nil {
		t.Fatalf("first SignUp returned error: %v", err)
	}

	_, err = service.SignUp(context.Background(), Credentials{
		Email:    "USER@example.com",
		Password: "supersecret123",
	})
	if !errors.Is(err, domain.ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

// TestSignInReturnsTokenForValidCredentials checks login returns a token on success.
func TestSignInReturnsTokenForValidCredentials(t *testing.T) {
	repo := newFakeUserRepository()
	service := NewService(repo, fakePasswordHasher{}, fakeTokenIssuer{})

	_, err := service.SignUp(context.Background(), Credentials{
		Email:    "user@example.com",
		Password: "supersecret123",
	})
	if err != nil {
		t.Fatalf("SignUp returned error: %v", err)
	}

	result, err := service.SignIn(context.Background(), Credentials{
		Email:    "user@example.com",
		Password: "supersecret123",
	})
	if err != nil {
		t.Fatalf("SignIn returned error: %v", err)
	}

	if result.Token.AccessToken != "test-token" {
		t.Fatalf("expected test token, got %q", result.Token.AccessToken)
	}
}

// TestSignInRejectsInvalidPassword checks login fails with a wrong password.
func TestSignInRejectsInvalidPassword(t *testing.T) {
	repo := newFakeUserRepository()
	service := NewService(repo, fakePasswordHasher{}, fakeTokenIssuer{})

	_, err := service.SignUp(context.Background(), Credentials{
		Email:    "user@example.com",
		Password: "supersecret123",
	})
	if err != nil {
		t.Fatalf("SignUp returned error: %v", err)
	}

	_, err = service.SignIn(context.Background(), Credentials{
		Email:    "user@example.com",
		Password: "wrong-password",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

type fakeUserRepository struct {
	usersByEmail map[string]domain.User
	nextID       int
}

// newFakeUserRepository creates an in-memory user store for tests.
func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		usersByEmail: make(map[string]domain.User),
		nextID:       1,
	}
}

// Create stores a user in memory and rejects duplicate emails.
func (r *fakeUserRepository) Create(ctx context.Context, email string, passwordHash string) (domain.User, error) {
	if _, exists := r.usersByEmail[email]; exists {
		return domain.User{}, domain.ErrUserAlreadyExists
	}

	user := domain.User{
		ID:           "user-1",
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}

	r.usersByEmail[email] = user
	r.nextID++

	return user, nil
}

// FindByEmail looks up a user by email in memory.
func (r *fakeUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, exists := r.usersByEmail[email]
	if !exists {
		return domain.User{}, domain.ErrUserNotFound
	}

	return user, nil
}

type fakePasswordHasher struct{}

// Hash returns a predictable fake hash for tests.
func (fakePasswordHasher) Hash(ctx context.Context, password string) (string, error) {
	return "hash:" + password, nil
}

// Compare checks against the fake hash format used in tests.
func (fakePasswordHasher) Compare(passwordHash string, plainPassword string) bool {
	return passwordHash == "hash:"+plainPassword
}

type fakeTokenIssuer struct{}

// Issue returns a fixed test token.
func (fakeTokenIssuer) Issue(ctx context.Context, user domain.User) (Token, error) {
	expiresAt := time.Now().UTC().Add(time.Hour)

	return Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   expiresAt,
	}, nil
}
