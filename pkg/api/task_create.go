package api

import (
	"context"
	"net/http"
	"unicode/utf8"

	httpapi "github.com/porebric/healthstep-libs/api"

	"github.com/porebric/task-tracker/internal/domain"
)

type CreateTaskRequest struct {
	TeamID      int64  `json:"team_id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	AssigneeID  *int64 `json:"assignee_id,omitempty"`
}

func NewCreateTaskRequest(ctx context.Context, r *http.Request) (context.Context, CreateTaskRequest, error) {
	var req CreateTaskRequest
	err := decodeBody(r, &req)
	return ctx, req, err
}

func (r CreateTaskRequest) Validate() error {
	if r.TeamID <= 0 {
		return httpapi.Invalid(paramTeamID, msgPositiveInteger)
	}
	if err := requireText("title", r.Title, maxTitleLen); err != nil {
		return err
	}
	if utf8.RuneCountInString(r.Description) > maxTextLen {
		return httpapi.Invalid("description", msgTooLong(maxTextLen))
	}
	if r.Status != "" {
		if _, ok := domain.ParseStatus(r.Status); !ok {
			return httpapi.Invalid("status", msgInvalidStatus)
		}
	}
	if r.AssigneeID != nil && *r.AssigneeID <= 0 {
		return httpapi.Invalid("assignee_id", msgPositiveInteger)
	}
	return nil
}

type CreateTaskResponse struct {
	httpapi.Created
	Task
}
