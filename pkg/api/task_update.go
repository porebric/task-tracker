package api

import (
	"context"
	"net/http"
	"unicode/utf8"

	httpapi "github.com/porebric/healthstep-libs/api"

	"github.com/porebric/task-tracker/internal/domain"
)

// UpdateTaskRequest changes only the fields present in the body. An assignee_id of 0 clears
// the assignee; version, when sent, must match the current one or the update is rejected.
type UpdateTaskRequest struct {
	ID          int64   `path:"id" json:"-"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	AssigneeID  *int64  `json:"assignee_id,omitempty"`
	Version     *int64  `json:"version,omitempty"`
}

func NewUpdateTaskRequest(ctx context.Context, r *http.Request) (context.Context, UpdateTaskRequest, error) {
	var req UpdateTaskRequest
	id, err := pathID(r, paramID)
	if err != nil {
		return ctx, req, err
	}
	if err := decodeBody(r, &req); err != nil {
		return ctx, req, err
	}
	req.ID = id
	return ctx, req, nil
}

func (r UpdateTaskRequest) Validate() error {
	if r.Title != nil {
		if err := requireText("title", *r.Title, maxTitleLen); err != nil {
			return err
		}
	}
	if r.Description != nil && utf8.RuneCountInString(*r.Description) > maxTextLen {
		return httpapi.Invalid("description", msgTooLong(maxTextLen))
	}
	if r.Status != nil {
		if _, ok := domain.ParseStatus(*r.Status); !ok {
			return httpapi.Invalid("status", msgInvalidStatus)
		}
	}
	if r.AssigneeID != nil && *r.AssigneeID < 0 {
		return httpapi.Invalid("assignee_id", "must be zero or a positive integer")
	}
	if r.Version != nil && *r.Version <= 0 {
		return httpapi.Invalid("version", msgPositiveInteger)
	}
	return nil
}

type UpdateTaskResponse struct {
	httpapi.OK
	Task
}
