package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	httpapi "github.com/porebric/healthstep-libs/api"
)

var errMalformedHeader = errors.New("auth: malformed authorization header")

type Middleware struct {
	issuer *Issuer
}

func NewMiddleware(issuer *Issuer) *Middleware {
	return &Middleware{issuer: issuer}
}

// Authenticate resolves the bearer token into the request context without rejecting
// anything: public routes ignore the outcome, protected ones apply Required.
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		if header == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		scheme, token, found := strings.Cut(header, " ")
		token = strings.TrimSpace(token)
		switch {
		case !found || !strings.EqualFold(scheme, "Bearer") || token == "":
			ctx = withAuthError(ctx, errMalformedHeader)
		default:
			if id, err := m.issuer.Parse(token); err != nil {
				ctx = withAuthError(ctx, err)
			} else {
				ctx = WithUserID(ctx, id)
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Required is the request-level guard of protected routes.
func Required(ctx context.Context, _ httpapi.Request) (context.Context, error) {
	if _, ok := UserID(ctx); ok {
		return ctx, nil
	}
	if authError(ctx) != nil {
		return ctx, httpapi.Unauthorized(httpapi.CodeUnauthorized, "invalid access token")
	}
	return ctx, httpapi.Unauthorized(httpapi.CodeUnauthorized, "missing access token")
}
