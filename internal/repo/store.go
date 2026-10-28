package repo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/porebric/task-tracker/internal/mysql"
)

// Store is the data access layer. Methods are grouped by entity in the sibling files; a Store
// bound to a transaction is obtained through Transaction.
type Store struct {
	db *gorm.DB
}

func New(d *mysql.DB) *Store {
	return &Store{db: d.Gorm()}
}

func (s *Store) Transaction(ctx context.Context, fn func(tx *Store) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&Store{db: tx})
	})
}

func (s *Store) q(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx)
}

func (s *Store) create(ctx context.Context, row any, op string) error {
	if err := s.q(ctx).Create(row).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// listByTask loads the rows that belong to one task in insertion order.
func listByTask[T any](ctx context.Context, s *Store, taskID int64, op string) ([]T, error) {
	var rows []T
	if err := s.q(ctx).Where("task_id = ?", taskID).Order("id").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

// notFound maps gorm.ErrRecordNotFound to the domain error and wraps anything else.
func notFound(err, domainErr error, op string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainErr
	}
	return fmt.Errorf("%s: %w", op, err)
}
