package security

import (
	"context"

	"time"

	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct {
	cost  int
	slots chan struct{}
}

func NewBcryptHasher() *BcryptHasher {

	// TODO:  we can make this configurable when adding in real production env.
	maxConcurrency := 4

	return &BcryptHasher{
		cost:  bcrypt.DefaultCost,
		slots: make(chan struct{}, maxConcurrency),
	}
}

func (h *BcryptHasher) Hash(ctx context.Context, password string) (string, error) {

	// Wait briefly for a bcrypt slot; if none is free before the timeout, fail fast instead of waiting forever.
	// TODO: we can make this timeout configurable when putting it real production env.
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	// adding bcrypt concurrency to protect high CPU usage
	//Without a cap:
	//500 signups -> 500 bcrypt hashes at once -> CPU spike -> everything gets slow
	//With a cap:
	//500 signups -> 4 bcrypt hashes at once -> rest wait -> server stays responsive
	select {
	case h.slots <- struct{}{}:
		defer func() { <-h.slots }()
	case <-ctx.Done():
		return "", ctxWithTimeout.Err()
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func (h *BcryptHasher) Compare(passwordHash string, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(plainPassword),
	)

	return err == nil
}
