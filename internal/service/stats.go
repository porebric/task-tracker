package service

import (
	"context"

	"github.com/porebric/task-tracker/internal/domain"
	"github.com/porebric/task-tracker/internal/repo"
)

type Stats struct {
	store *repo.Store
}

func NewStats(store *repo.Store) *Stats {
	return &Stats{store: store}
}

func (s *Stats) Team(ctx context.Context, actorID, teamID int64) (*domain.TeamStats, error) {
	role, err := memberRole(ctx, s.store, teamID, actorID)
	if err != nil {
		return nil, err
	}
	if !CanViewStats(role) {
		return nil, domain.ErrRoleNotAllowed
	}
	return s.store.TeamStats(ctx, teamID)
}
