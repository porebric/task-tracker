package handler

import (
	"context"

	"github.com/porebric/task-tracker/pkg/api"
)

func (h *Handler) Register(ctx context.Context, in api.RegisterRequest) (api.RegisterResponse, error) {
	user, err := h.auth.Register(ctx, in.Email, in.Password, in.Name)
	if err != nil {
		return api.RegisterResponse{}, toAPIError(err)
	}
	return api.RegisterResponse{User: toUser(user)}, nil
}

func (h *Handler) Login(ctx context.Context, in api.LoginRequest) (api.LoginResponse, error) {
	token, err := h.auth.Login(ctx, in.Email, in.Password)
	if err != nil {
		return api.LoginResponse{}, toAPIError(err)
	}
	return api.LoginResponse{AccessToken: token.Value, TokenType: "Bearer", ExpiresAt: token.ExpiresAt}, nil
}
