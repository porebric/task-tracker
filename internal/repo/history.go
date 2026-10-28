package repo

import (
	"context"

	"github.com/porebric/task-tracker/internal/domain"
)

func (s *Store) AddHistory(ctx context.Context, h *domain.TaskHistory) error {
	return s.create(ctx, h, "repo: add history")
}

func (s *Store) HistoryByTask(ctx context.Context, taskID int64) ([]domain.TaskHistory, error) {
	return listByTask[domain.TaskHistory](ctx, s, taskID, "repo: history by task")
}
