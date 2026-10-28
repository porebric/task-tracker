package handler

import (
	"context"

	httpapi "github.com/porebric/healthstep-libs/api"

	"github.com/porebric/task-tracker/internal/domain"
	"github.com/porebric/task-tracker/internal/service"
	"github.com/porebric/task-tracker/pkg/api"
)

func (h *Handler) CreateTeam(ctx context.Context, in api.CreateTeamRequest) (api.CreateTeamResponse, error) {
	return withActor(ctx, func(actor int64) (api.CreateTeamResponse, error) {
		team, err := h.teams.Create(ctx, actor, in.Name)
		if err != nil {
			return api.CreateTeamResponse{}, err
		}
		return api.CreateTeamResponse{Team: toTeam(team)}, nil
	})
}

func (h *Handler) ListTeams(ctx context.Context, _ httpapi.Empty) (api.ListTeamsResponse, error) {
	return withActor(ctx, func(actor int64) (api.ListTeamsResponse, error) {
		memberships, err := h.teams.ListForUser(ctx, actor)
		if err != nil {
			return api.ListTeamsResponse{}, err
		}
		resp := api.ListTeamsResponse{Teams: make([]api.TeamMembership, 0, len(memberships))}
		for _, m := range memberships {
			resp.Teams = append(resp.Teams, toMembership(m))
		}
		return resp, nil
	})
}

func (h *Handler) Invite(ctx context.Context, in api.InviteRequest) (api.InviteResponse, error) {
	return withActor(ctx, func(actor int64) (api.InviteResponse, error) {
		member, err := h.teams.Invite(ctx, actor, in.TeamID, service.InviteInput{
			UserID: in.UserID,
			Email:  in.Email,
			Role:   domain.Role(in.Role),
		})
		if err != nil {
			return api.InviteResponse{}, err
		}
		return api.InviteResponse{TeamID: member.TeamID, UserID: member.UserID, Role: string(member.Role)}, nil
	})
}

func (h *Handler) TeamStats(ctx context.Context, in api.TeamStatsRequest) (api.TeamStatsResponse, error) {
	return withActor(ctx, func(actor int64) (api.TeamStatsResponse, error) {
		stats, err := h.stats.Team(ctx, actor, in.TeamID)
		if err != nil {
			return api.TeamStatsResponse{}, err
		}
		return api.TeamStatsResponse{TeamStats: toStats(stats)}, nil
	})
}
