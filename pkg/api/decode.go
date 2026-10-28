package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	httpapi "github.com/porebric/healthstep-libs/api"
)

const (
	paramID     = "id"
	paramTeamID = "team_id"
)

func NewEmptyRequest(ctx context.Context, _ *http.Request) (context.Context, httpapi.Empty, error) {
	return ctx, httpapi.Empty{}, nil
}

func decodeBody(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return httpapi.BadRequest(httpapi.CodeInvalidRequest, "request body is required")
		}
		return httpapi.BadRequest(httpapi.CodeInvalidRequest, "invalid json body")
	}
	return nil
}

func pathID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, name), 10, 64)
	if err != nil || id <= 0 {
		return 0, httpapi.Invalid(name, msgPositiveInteger)
	}
	return id, nil
}

// queryInt64 reads an optional positive integer parameter; absence yields zero.
func queryInt64(r *http.Request, name string) (int64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return 0, nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || v <= 0 {
		return 0, httpapi.Invalid(name, msgPositiveInteger)
	}
	return v, nil
}

// queryInt reads an optional integer parameter, falling back to def when absent.
func queryInt(r *http.Request, name string, def int) (int, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return def, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, httpapi.Invalid(name, "must be an integer")
	}
	return v, nil
}
