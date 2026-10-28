package api

import (
	"context"
	"net/http"
	"time"

	httpapi "github.com/porebric/healthstep-libs/api"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewLoginRequest(ctx context.Context, r *http.Request) (context.Context, LoginRequest, error) {
	var req LoginRequest
	err := decodeBody(r, &req)
	return ctx, req, err
}

func (r LoginRequest) Validate() error {
	if err := validateEmail(r.Email); err != nil {
		return err
	}
	if r.Password == "" {
		return httpapi.Invalid("password", msgRequired)
	}
	return nil
}

type LoginResponse struct {
	httpapi.OK
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}
