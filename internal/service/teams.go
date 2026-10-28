package service

import (
	"context"
	"time"

	"github.com/porebric/task-tracker/internal/domain"
	"github.com/porebric/task-tracker/internal/repo"
)

type Teams struct {
	store *repo.Store
	now   func() time.Time
}

func NewTeams(store *repo.Store) *Teams {
	return &Teams{store: store, now: utcNow}
}

// Create inserts the team and its owner membership in one transaction.
func (s *Teams) Create(ctx context.Context, ownerID int64, name string) (*domain.Team, error) {
	now := s.now()
	team := &domain.Team{Name: name, CreatedBy: ownerID, CreatedAt: now}
	err := s.store.Transaction(ctx, func(tx *repo.Store) error {
		if err := tx.CreateTeam(ctx, team); err != nil {
			return err
		}
		return tx.AddMember(ctx, &domain.TeamMember{
			TeamID:    team.ID,
			UserID:    ownerID,
			Role:      domain.RoleOwner,
			CreatedAt: now,
		})
	})
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (s *Teams) ListForUser(ctx context.Context, userID int64) ([]domain.TeamMembership, error) {
	return s.store.TeamsByUser(ctx, userID)
}

// InviteInput names the invited user by id or by email and the role to grant.
type InviteInput struct {
	UserID int64
	Email  string
	Role   domain.Role
}

func (s *Teams) Invite(ctx context.Context, actorID, teamID int64, in InviteInput) (*domain.TeamMember, error) {
	actorRole, err := memberRole(ctx, s.store, teamID, actorID)
	if err != nil {
		return nil, err
	}
	role := in.Role
	if role == "" {
		role = domain.RoleMember
	}
	if !CanInvite(actorRole, role) {
		return nil, domain.ErrRoleNotAllowed
	}

	user, err := s.resolveUser(ctx, in)
	if err != nil {
		return nil, err
	}
	member := &domain.TeamMember{TeamID: teamID, UserID: user.ID, Role: role, CreatedAt: s.now()}
	if err := s.store.AddMember(ctx, member); err != nil {
		return nil, err
	}
	return member, nil
}

func (s *Teams) resolveUser(ctx context.Context, in InviteInput) (*domain.User, error) {
	if in.UserID > 0 {
		return s.store.UserByID(ctx, in.UserID)
	}
	return s.store.UserByEmail(ctx, normalizeEmail(in.Email))
}
