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
	return idRequest(ctx, r, func(id int64) GetTaskRequest { return GetTaskRequest{ID: id} })
}

func (GetTaskRequest) Validate() error {
	return nil
}

type GetTaskResponse struct {
	httpapi.OK
	Task
}
