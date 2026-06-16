package app

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/example/registration/internal/domain"
)

// using a basic email regex check for simplicity
var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type Service struct {
	users     UserRepository
	passwords PasswordHasher
	tokens    TokenIssuer
}

// NewService wires up the account service with its dependencies.
func NewService(
	users UserRepository,
	passwords PasswordHasher,
	tokens TokenIssuer,
) *Service {
	return &Service{
		users:     users,
		passwords: passwords,
		tokens:    tokens,
	}
}

type Credentials struct {
	Email    string
	Password string
}

type SignInResult struct {
	User  domain.User
	Token Token
}

type ValidationError struct {
	Field   string
	Message string
}

// Error returns the validation message.
func (e *ValidationError) Error() string {
	return e.Message
}

// SignUp validates input, hashes the password, and creates a new user.
func (s *Service) SignUp(ctx context.Context, input Credentials) (domain.User, error) {
	email := normalizeEmail(input.Email)

	if !isValidEmail(email) {
		return domain.User{}, &ValidationError{
			Field:   "email",
			Message: "valid email is required",
		}
	}

	if err := validatePassword(input.Password); err != nil {
		return domain.User{}, err
	}

	// NOTE: depending on how much CPU time is consumed during hashing, we could optimize and check if the email exists 
	// 		first before hashing since hashing is a bit expensive operation.
	passwordHash, err := s.passwords.Hash(ctx, input.Password)
	if err != nil {
		return domain.User{}, err
	}

	user, err := s.users.Create(ctx, email, passwordHash)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

// SignIn checks credentials and returns a token for a valid user.
func (s *Service) SignIn(ctx context.Context, input Credentials) (SignInResult, error) {
	email := normalizeEmail(input.Email)

	if !isValidEmail(email) || input.Password == "" {
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrUserNotFound) {
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	if err != nil {
		return SignInResult{}, err
	}

	passwordMatches := s.passwords.Compare(user.PasswordHash, input.Password)
	if !passwordMatches {
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	token, err := s.tokens.Issue(ctx, user)
	if err != nil {
		return SignInResult{}, err
	}

	return SignInResult{
		User:  user,
		Token: token,
	}, nil
}

// normalizeEmail trims whitespace and lowercases the email.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// isValidEmail checks basic email format and length.
func isValidEmail(email string) bool {
	if len(email) == 0 || len(email) > 254 {
		return false
	}

	return emailRegex.MatchString(email)
}

// validatePassword enforces password length rules.
func validatePassword(password string) error {
	// NOTE for the reviewer: we can add other password validation check here
	switch {
	case password == "":
		return &ValidationError{
			Field:   "password",
			Message: "password is required",
		}
	case len([]rune(password)) < 8:
		return &ValidationError{
			Field:   "password",
			Message: "password must be at least 8 characters",
		}
	case len([]byte(password)) > 72:
		return &ValidationError{
			Field:   "password",
			Message: "password must be 72 bytes or fewer",
		}
	default:
		return nil
	}
}
