package handler

import (
	"context"

	httpapi "github.com/porebric/healthstep-libs/api"

	"github.com/porebric/task-tracker/internal/auth"
	"github.com/porebric/task-tracker/internal/service"
)

type Handler struct {
	auth  *service.Auth
	teams *service.Teams
	tasks *service.Tasks
	stats *service.Stats
}

func New(a *service.Auth, teams *service.Teams, tasks *service.Tasks, stats *service.Stats) *Handler {
	return &Handler{auth: a, teams: teams, tasks: tasks, stats: stats}
}

// withActor runs fn for the authenticated user and translates service errors into API ones.
func withActor[Out any](ctx context.Context, fn func(actor int64) (Out, error)) (Out, error) {
	var zero Out
	actor, ok := auth.UserID(ctx)
	if !ok {
		return zero, httpapi.Unauthorized(httpapi.CodeUnauthorized, "missing access token")
	}
	out, err := fn(actor)
	if err != nil {
		return zero, toAPIError(err)
	}
	return out, nil
}
