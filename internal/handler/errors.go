package handler

import (
	"errors"
	"strings"

	httpapi "github.com/porebric/healthstep-libs/api"

	"github.com/porebric/task-tracker/internal/domain"
	"github.com/porebric/task-tracker/pkg/api"
)

type errorMapping struct {
	err     error
	status  func(code int, message string) *httpapi.Error
	code    int
	message string
}

var errorMappings = []errorMapping{
	{domain.ErrEmailTaken, httpapi.Conflict, api.CodeEmailTaken, "email is already registered"},
	{domain.ErrInvalidCredentials, httpapi.Unauthorized, api.CodeInvalidCredentials, "invalid email or password"},
	{domain.ErrUserNotFound, httpapi.NotFound, api.CodeUserNotFound, "user not found"},
	{domain.ErrTeamNotFound, httpapi.NotFound, api.CodeTeamNotFound, "team not found"},
	{domain.ErrNotTeamMember, httpapi.Forbidden, api.CodeNotTeamMember, "you are not a member of this team"},
	{domain.ErrRoleNotAllowed, httpapi.Forbidden, api.CodeRoleNotAllowed, "your role does not allow this action"},
	{domain.ErrAlreadyMember, httpapi.Conflict, api.CodeAlreadyMember, "user is already a member of the team"},
	{domain.ErrTaskNotFound, httpapi.NotFound, api.CodeTaskNotFound, "task not found"},
	{domain.ErrAssigneeNotMember, httpapi.BadRequest, api.CodeAssigneeNotMember, "assignee is not a member of the team"},
	{domain.ErrVersionConflict, httpapi.Conflict, api.CodeVersionConflict, "task was modified concurrently, refresh and retry"},
}

// toAPIError maps domain errors to responses; anything unknown stays an internal error.
func toAPIError(err error) error {
	var apiErr *httpapi.Error
	if errors.As(err, &apiErr) {
		return apiErr
	}
	var forbidden *domain.ForbiddenFieldsError
	if errors.As(err, &forbidden) {
		return httpapi.Forbidden(api.CodeFieldNotEditable, "fields are not editable by this user").
			WithValue("fields", strings.Join(forbidden.Fields, ", "))
	}
	for _, m := range errorMappings {
		if errors.Is(err, m.err) {
			return m.status(m.code, m.message)
		}
	}
	return err
}
