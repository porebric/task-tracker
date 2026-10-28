package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/porebric/task-tracker/internal/auth"
	"github.com/porebric/task-tracker/internal/domain"
	"github.com/porebric/task-tracker/internal/repo"
)

type Auth struct {
	store  *repo.Store
	issuer *auth.Issuer
	hasher auth.PasswordHasher
	now    func() time.Time
}

func NewAuth(store *repo.Store, issuer *auth.Issuer, hasher auth.PasswordHasher) *Auth {
	return &Auth{store: store, issuer: issuer, hasher: hasher, now: utcNow}
}

func (s *Auth) Register(ctx context.Context, email, password, name string) (*domain.User, error) {
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}
	user := &domain.User{
		Email:        normalizeEmail(email),
		PasswordHash: hash,
		Name:         strings.TrimSpace(name),
		CreatedAt:    s.now(),
	}
	if err := s.store.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Login answers an unknown email and a wrong password with the same error.
func (s *Auth) Login(ctx context.Context, email, password string) (auth.Token, error) {
	user, err := s.store.UserByEmail(ctx, normalizeEmail(email))
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		s.hasher.Burn(password)
		return auth.Token{}, domain.ErrInvalidCredentials
	case err != nil:
		return auth.Token{}, err
	}
	if !s.hasher.Compare(user.PasswordHash, password) {
		return auth.Token{}, domain.ErrInvalidCredentials
	}
	return s.issuer.Issue(user.ID)
}
