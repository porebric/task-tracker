package repo

import (
	"context"

	"github.com/porebric/task-tracker/internal/domain"
)

func (s *Store) AddComment(ctx context.Context, c *domain.TaskComment) error {
	return s.create(ctx, c, "repo: add comment")
}

func (s *Store) CommentsByTask(ctx context.Context, taskID int64) ([]domain.TaskComment, error) {
	return listByTask[domain.TaskComment](ctx, s, taskID, "repo: comments by task")
}
