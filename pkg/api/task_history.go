package api

import (
	"context"
	"net/http"

	httpapi "github.com/porebric/healthstep-libs/api"
)

type TaskHistoryRequest struct {
	ID int64 `path:"id" json:"-"`
}

func NewTaskHistoryRequest(ctx context.Context, r *http.Request) (context.Context, TaskHistoryRequest, error) {
	return idRequest(ctx, r, func(id int64) TaskHistoryRequest { return TaskHistoryRequest{ID: id} })
}

func (TaskHistoryRequest) Validate() error {
	return nil
}

type TaskHistoryResponse struct {
	httpapi.OK
	History []HistoryEntry `json:"history"`
}
