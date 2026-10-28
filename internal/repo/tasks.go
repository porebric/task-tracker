package repo

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/porebric/task-tracker/internal/domain"
)

type TaskFilter struct {
	TeamID int64
	// Status narrows the list when set.
	Status domain.Status
	// AssigneeID narrows the list when positive.
	AssigneeID int64
	Limit      int
	Offset     int
	// Cursor switches to keyset pagination: only tasks with a smaller id are returned.
	Cursor int64
}

func (s *Store) CreateTask(ctx context.Context, t *domain.Task) error {
	return s.create(ctx, t, "repo: create task")
}

func (s *Store) TaskByID(ctx context.Context, id int64) (*domain.Task, error) {
	var t domain.Task
	if err := s.q(ctx).Take(&t, id).Error; err != nil {
		return nil, notFound(err, domain.ErrTaskNotFound, "repo: task by id")
	}
	return &t, nil
}

// TaskByIDForUpdate locks the row until the surrounding transaction ends, so concurrent
// editors are serialized and each of them works from the latest committed state.
func (s *Store) TaskByIDForUpdate(ctx context.Context, id int64) (*domain.Task, error) {
	var t domain.Task
	err := s.q(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Take(&t, id).Error
	if err != nil {
		return nil, notFound(err, domain.ErrTaskNotFound, "repo: task for update")
	}
	return &t, nil
}

// UpdateTask writes the mutable fields guarded by the version the caller read; on success
// the version in t is bumped to the stored one.
func (s *Store) UpdateTask(ctx context.Context, t *domain.Task) error {
	res := s.q(ctx).Model(&domain.Task{}).
		Where("id = ? AND version = ?", t.ID, t.Version).
		Updates(map[string]any{
			"title":       t.Title,
			"description": t.Description,
			"status":      t.Status,
			"assignee_id": t.AssigneeID,
			"closed_at":   t.ClosedAt,
			"updated_at":  t.UpdatedAt,
			"version":     gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return fmt.Errorf("repo: update task: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrVersionConflict
	}
	t.Version++
	return nil
}

func (s *Store) ListTasks(ctx context.Context, f TaskFilter) (tasks []domain.Task, total int64, err error) {
	base := s.q(ctx).Model(&domain.Task{}).Where("team_id = ?", f.TeamID)
	if f.Status != "" {
		base = base.Where("status = ?", f.Status)
	}
	if f.AssigneeID > 0 {
		base = base.Where("assignee_id = ?", f.AssigneeID)
	}
	base = base.Session(&gorm.Session{})

	if err = base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("repo: count tasks: %w", err)
	}

	page := base
	if f.Cursor > 0 {
		page = page.Where("id < ?", f.Cursor)
	} else {
		page = page.Offset(f.Offset)
	}
	if err = page.Order("id DESC").Limit(f.Limit).Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("repo: list tasks: %w", err)
	}
	return tasks, total, nil
}
