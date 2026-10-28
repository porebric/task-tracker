package api

import (
	"context"
	"net/http"

	httpapi "github.com/porebric/healthstep-libs/api"
)

type TeamStatsRequest struct {
	TeamID int64 `path:"team_id" json:"-"`
}

func NewTeamStatsRequest(ctx context.Context, r *http.Request) (context.Context, TeamStatsRequest, error) {
	id, err := pathID(r, paramTeamID)
	return ctx, TeamStatsRequest{TeamID: id}, err
}

func (TeamStatsRequest) Validate() error {
	return nil
}

type TeamStatsResponse struct {
	httpapi.OK
	TeamStats
}
