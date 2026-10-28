package domain

import (
	"errors"
	"strings"
)

var (
	ErrEmailTaken         = errors.New("email is already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrTeamNotFound       = errors.New("team not found")
	ErrNotTeamMember      = errors.New("user is not a member of the team")
	ErrRoleNotAllowed     = errors.New("role does not allow this action")
	ErrAlreadyMember      = errors.New("user is already a member of the team")
	ErrTaskNotFound       = errors.New("task not found")
	ErrAssigneeNotMember  = errors.New("assignee is not a member of the team")
	ErrVersionConflict    = errors.New("task was modified concurrently, refresh and retry")
)

// ForbiddenFieldsError reports task fields the actor is not allowed to change.
type ForbiddenFieldsError struct {
	Fields []string
}

func (e *ForbiddenFieldsError) Error() string {
	return "fields are not editable by this user: " + strings.Join(e.Fields, ", ")
}
