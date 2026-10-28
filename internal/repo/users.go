package repo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/porebric/task-tracker/internal/domain"
)

func (s *Store) CreateUser(ctx context.Context, u *domain.User) error {
	err := s.q(ctx).Create(u).Error
	switch {
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return domain.ErrEmailTaken
	case err != nil:
		return fmt.Errorf("repo: create user: %w", err)
	}
	return nil
}

func (s *Store) UserByID(ctx context.Context, id int64) (*domain.User, error) {
	var u domain.User
	if err := s.q(ctx).Take(&u, id).Error; err != nil {
		return nil, notFound(err, domain.ErrUserNotFound, "repo: user by id")
	}
	return &u, nil
}

func (s *Store) UserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	if err := s.q(ctx).Where("email = ?", email).Take(&u).Error; err != nil {
		return nil, notFound(err, domain.ErrUserNotFound, "repo: user by email")
	}
	return &u, nil
}
