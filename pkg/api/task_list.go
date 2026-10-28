package api

import (
	"context"
	"fmt"
	"net/http"

	httpapi "github.com/porebric/healthstep-libs/api"

	"github.com/porebric/task-tracker/internal/domain"
)

// ListTasksRequest pages through the tasks of one team. Offset pagination is the default;
// a cursor (the id of the last task seen) switches to keyset pagination and ignores offset.
type ListTasksRequest struct {
	TeamID     int64  `query:"team_id" json:"-"`
	Status     string `query:"status" json:"-"`
	AssigneeID int64  `query:"assignee_id" json:"-"`
	Limit      int    `query:"limit" json:"-"`
	Offset     int    `query:"offset" json:"-"`
	Cursor     int64  `query:"cursor" json:"-"`
}

func NewListTasksRequest(ctx context.Context, r *http.Request) (context.Context, ListTasksRequest, error) {
	var (
		req ListTasksRequest
		err error
	)
	if req.TeamID, err = queryInt64(r, paramTeamID); err != nil {
		return ctx, req, err
	}
	if req.AssigneeID, err = queryInt64(r, "assignee_id"); err != nil {
		return ctx, req, err
	}
	if req.Cursor, err = queryInt64(r, "cursor"); err != nil {
		return ctx, req, err
	}
	if req.Limit, err = queryInt(r, "limit", DefaultPageLimit); err != nil {
		return ctx, req, err
	}
	if req.Offset, err = queryInt(r, "offset", 0); err != nil {
		return ctx, req, err
	}
	req.Status = r.URL.Query().Get("status")
	return ctx, req, nil
}

func (r ListTasksRequest) Validate() error {
	if r.TeamID <= 0 {
		return httpapi.Invalid(paramTeamID, msgRequired)
	}
	if r.Status != "" {
		if _, ok := domain.ParseStatus(r.Status); !ok {
			return httpapi.Invalid("status", msgInvalidStatus)
		}
	}
	if r.Limit < 1 || r.Limit > MaxPageLimit {
		return httpapi.Invalid("limit", fmt.Sprintf("must be between 1 and %d", MaxPageLimit))
	}
	if r.Offset < 0 {
		return httpapi.Invalid("offset", "must not be negative")
	}
	return nil
}

type ListTasksResponse struct {
	httpapi.OK
	Tasks      []Task `json:"tasks"`
	Total      int64  `json:"total"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
	NextCursor int64  `json:"next_cursor,omitempty"`
}
