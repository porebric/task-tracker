package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewListTasksRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?team_id=5&status=todo&assignee_id=2", http.NoBody)
	_, req, err := NewListTasksRequest(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, ListTasksRequest{TeamID: 5, Status: "todo", AssigneeID: 2, Limit: DefaultPageLimit}, req)
	require.NoError(t, req.Validate())

	r = httptest.NewRequest(http.MethodGet, "/api/v1/tasks?team_id=5&limit=500", http.NoBody)
	_, req, err = NewListTasksRequest(context.Background(), r)
	require.NoError(t, err)
	require.Error(t, req.Validate())

	r = httptest.NewRequest(http.MethodGet, "/api/v1/tasks?team_id=5&status=unknown", http.NoBody)
	_, req, err = NewListTasksRequest(context.Background(), r)
	require.NoError(t, err)
	require.Error(t, req.Validate())

	r = httptest.NewRequest(http.MethodGet, "/api/v1/tasks?team_id=abc", http.NoBody)
	_, _, err = NewListTasksRequest(context.Background(), r)
	require.Error(t, err)

	r = httptest.NewRequest(http.MethodGet, "/api/v1/tasks", http.NoBody)
	_, req, err = NewListTasksRequest(context.Background(), r)
	require.NoError(t, err)
	require.Error(t, req.Validate(), "team_id is required")
}

func TestInviteRequest_Validate(t *testing.T) {
	tests := []struct {
		name string
		req  InviteRequest
		ok   bool
	}{
		{name: "by id", req: InviteRequest{UserID: 1}, ok: true},
		{name: "by email as admin", req: InviteRequest{Email: "a@b.io", Role: "admin"}, ok: true},
		{name: "nothing", req: InviteRequest{}},
		{name: "both", req: InviteRequest{UserID: 1, Email: "a@b.io"}},
		{name: "owner role", req: InviteRequest{UserID: 1, Role: "owner"}},
		{name: "bad email", req: InviteRequest{Email: "nope"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.ok {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestRegisterRequest_Validate(t *testing.T) {
	require.NoError(t, RegisterRequest{Email: "a@b.io", Password: "12345678", Name: "Ann"}.Validate())
	require.Error(t, RegisterRequest{Email: "Ann <a@b.io>", Password: "12345678", Name: "Ann"}.Validate())
	require.Error(t, RegisterRequest{Email: "a@b.io", Password: "short", Name: "Ann"}.Validate())
	require.Error(t, RegisterRequest{Email: "a@b.io", Password: "12345678", Name: "  "}.Validate())
}

func TestUpdateTaskRequest_Validate(t *testing.T) {
	empty, bad, long := "", "unknown", strings.Repeat("x", maxTitleLen+1)
	zero, negative := int64(0), int64(-1)
	require.NoError(t, UpdateTaskRequest{AssigneeID: &zero}.Validate())
	require.Error(t, UpdateTaskRequest{Title: &empty}.Validate())
	require.Error(t, UpdateTaskRequest{Title: &long}.Validate())
	require.Error(t, UpdateTaskRequest{Status: &bad}.Validate())
	require.Error(t, UpdateTaskRequest{AssigneeID: &negative}.Validate())
	require.Error(t, UpdateTaskRequest{Version: &zero}.Validate())
}

func TestDecodeBody(t *testing.T) {
	var dst CreateTeamRequest
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"core"}`))
	require.NoError(t, decodeBody(r, &dst))
	assert.Equal(t, "core", dst.Name)

	r = httptest.NewRequest(http.MethodPost, "/", http.NoBody)
	require.Error(t, decodeBody(r, &dst))
	r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":`))
	require.Error(t, decodeBody(r, &dst))
}
