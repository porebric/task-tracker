package api

import (
	"context"
	"net/http"

	httpapi "github.com/porebric/healthstep-libs/api"
)

type GetTaskRequest struct {
	ID int64 `path:"id" json:"-"`
}

func NewGetTaskRequest(ctx context.Context, r *http.Request) (context.Context, GetTaskRequest, error) {
	id, err := pathID(r, paramID)
	return ctx, GetTaskRequest{ID: id}, err
}

func (GetTaskRequest) Validate() error {
	return nil
}

type GetTaskResponse struct {
	httpapi.OK
	Task
}
