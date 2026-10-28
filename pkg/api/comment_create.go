package api

import (
	"context"
	"net/http"
)

type CreateCommentRequest struct {
	ID      int64  `path:"id" json:"-"`
	Content string `json:"content"`
}

func NewCreateCommentRequest(ctx context.Context, r *http.Request) (context.Context, CreateCommentRequest, error) {
	var req CreateCommentRequest
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

func (r CreateCommentRequest) Validate() error {
	return requireText("content", r.Content, maxTextLen)
}

type CreateCommentResponse struct {
	Comment
}

func (CreateCommentResponse) StatusCode() int {
	return http.StatusCreated
}
