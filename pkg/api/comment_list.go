package api

import (
	"context"
	"net/http"

	httpapi "github.com/porebric/healthstep-libs/api"
)

type ListCommentsRequest struct {
	ID int64 `path:"id" json:"-"`
}

func NewListCommentsRequest(ctx context.Context, r *http.Request) (context.Context, ListCommentsRequest, error) {
	return idRequest(ctx, r, func(id int64) ListCommentsRequest { return ListCommentsRequest{ID: id} })
}

func (ListCommentsRequest) Validate() error {
	return nil
}

type ListCommentsResponse struct {
	httpapi.OK
	Comments []Comment `json:"comments"`
}
