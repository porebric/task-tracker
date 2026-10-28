package repo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/porebric/task-tracker/internal/domain"
)

func (s *Store) CreateTeam(ctx context.Context, t *domain.Team) error {
	return s.create(ctx, t, "repo: create team")
}

func (s *Store) TeamByID(ctx context.Context, id int64) (*domain.Team, error) {
	var t domain.Team
	if err := s.q(ctx).Take(&t, id).Error; err != nil {
		return nil, notFound(err, domain.ErrTeamNotFound, "repo: team by id")
	}
	return &t, nil
}

func (s *Store) AddMember(ctx context.Context, m *domain.TeamMember) error {
	err := s.q(ctx).Create(m).Error
	switch {
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return domain.ErrAlreadyMember
	case err != nil:
		return fmt.Errorf("repo: add member: %w", err)
	}
	return nil
}

func (s *Store) MemberRole(ctx context.Context, teamID, userID int64) (domain.Role, error) {
	var m domain.TeamMember
	err := s.q(ctx).Where("team_id = ? AND user_id = ?", teamID, userID).Take(&m).Error
	if err != nil {
		return "", notFound(err, domain.ErrNotTeamMember, "repo: member role")
	}
	return m.Role, nil
}

func (s *Store) IsMember(ctx context.Context, teamID, userID int64) (bool, error) {
	_, err := s.MemberRole(ctx, teamID, userID)
	switch {
	case errors.Is(err, domain.ErrNotTeamMember):
		return false, nil
	case err != nil:
		return false, err
	}
	return true, nil
}

func (s *Store) TeamsByUser(ctx context.Context, userID int64) ([]domain.TeamMembership, error) {
	var rows []domain.TeamMembership
	err := s.q(ctx).
		Table("teams AS t").
		Select("t.id, t.name, t.created_by, t.created_at, m.role").
		Joins("JOIN team_members AS m ON m.team_id = t.id").
		Where("m.user_id = ?", userID).
		Order("t.id").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("repo: teams by user: %w", err)
	}
	return rows, nil
}
