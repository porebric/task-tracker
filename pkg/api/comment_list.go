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
	id, err := pathID(r, paramID)
	return ctx, ListCommentsRequest{ID: id}, err
}

func (ListCommentsRequest) Validate() error {
	return nil
}

type ListCommentsResponse struct {
	httpapi.OK
	Comments []Comment `json:"comments"`
}
