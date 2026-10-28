package api

import (
	"context"
	"net/http"

	httpapi "github.com/porebric/healthstep-libs/api"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func NewRegisterRequest(ctx context.Context, r *http.Request) (context.Context, RegisterRequest, error) {
	var req RegisterRequest
	err := decodeBody(r, &req)
	return ctx, req, err
}

func (r RegisterRequest) Validate() error {
	if err := validateEmail(r.Email); err != nil {
		return err
	}
	if n := len(r.Password); n < minPasswordLen || n > maxPasswordLen {
		return httpapi.Invalid("password", "must be 8 to 72 characters")
	}
	return requireText("name", r.Name, maxNameLen)
}

type RegisterResponse struct {
	httpapi.Created
	User
}
