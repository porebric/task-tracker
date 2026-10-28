package api

import (
	"context"
	"net/http"

	httpapi "github.com/porebric/healthstep-libs/api"

	"github.com/porebric/task-tracker/internal/domain"
)

// InviteRequest adds a user to the team by id or by email; the role defaults to member.
type InviteRequest struct {
	TeamID int64  `path:"id" json:"-"`
	UserID int64  `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
	Role   string `json:"role,omitempty"`
}

func NewInviteRequest(ctx context.Context, r *http.Request) (context.Context, InviteRequest, error) {
	var req InviteRequest
	id, err := pathID(r, paramID)
	if err != nil {
		return ctx, req, err
	}
	if err := decodeBody(r, &req); err != nil {
		return ctx, req, err
	}
	req.TeamID = id
	return ctx, req, nil
}

func (r InviteRequest) Validate() error {
	switch {
	case r.UserID < 0:
		return httpapi.Invalid("user_id", msgPositiveInteger)
	case r.UserID == 0 && r.Email == "":
		return httpapi.Invalid("user_id", "user_id or email is required")
	case r.UserID > 0 && r.Email != "":
		return httpapi.Invalid("email", "only one of user_id and email can be set")
	}
	if r.Email != "" {
		if err := validateEmail(r.Email); err != nil {
			return err
		}
	}
	switch r.Role {
	case "", string(domain.RoleMember), string(domain.RoleAdmin):
		return nil
	default:
		return httpapi.Invalid("role", "must be admin or member")
	}
}

type InviteResponse struct {
	httpapi.Created
	TeamID int64  `json:"team_id"`
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
}
