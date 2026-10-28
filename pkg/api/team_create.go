package api

import (
	"context"
	"net/http"

	httpapi "github.com/porebric/healthstep-libs/api"
)

type CreateTeamRequest struct {
	Name string `json:"name"`
}

func NewCreateTeamRequest(ctx context.Context, r *http.Request) (context.Context, CreateTeamRequest, error) {
	var req CreateTeamRequest
	err := decodeBody(r, &req)
	return ctx, req, err
}

func (r CreateTeamRequest) Validate() error {
	return requireText("name", r.Name, maxNameLen)
}

type CreateTeamResponse struct {
	httpapi.Created
	Team
}
