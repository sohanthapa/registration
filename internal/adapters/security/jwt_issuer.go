package security

import (
	"context"
	"time"

	"github.com/example/registration/internal/app"
	"github.com/example/registration/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

type JWTIssuer struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

func NewJWTIssuer(secret string, issuer string, ttl time.Duration) *JWTIssuer {
	return &JWTIssuer{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
	}
}

type accessTokenClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func (j *JWTIssuer) Issue(ctx context.Context, user domain.User) (app.Token, error) {
	if err := ctx.Err(); err != nil {
		return app.Token{}, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(j.ttl)

	claims := accessTokenClaims{
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(j.secret)
	if err != nil {
		return app.Token{}, err
	}

	return app.Token{
		AccessToken: signedToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(j.ttl.Seconds()),
		ExpiresAt:   expiresAt,
	}, nil
}
