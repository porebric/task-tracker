package service

import (
	"context"
	"slices"

	"github.com/porebric/task-tracker/internal/domain"
	"github.com/porebric/task-tracker/internal/repo"
)

// Task fields that can change through the API; history entries use the same names.
const (
	FieldTitle       = "title"
	FieldDescription = "description"
	FieldStatus      = "status"
	FieldAssignee    = "assignee_id"
)

var allTaskFields = []string{FieldTitle, FieldDescription, FieldStatus, FieldAssignee}

// EditableFields lists the task fields the user may change given their role in the team.
// Owners and admins edit anything, the creator edits their own task, the assignee may only
// move the task between statuses, everybody else gets nothing.
func EditableFields(role domain.Role, task *domain.Task, userID int64) []string {
	switch {
	case role.CanManageTeam(), task.CreatedBy == userID:
		return allTaskFields
	case task.IsAssignee(userID):
		return []string{FieldStatus}
	default:
		return nil
	}
}

// CheckEditable rejects changes to fields outside EditableFields.
func CheckEditable(role domain.Role, task *domain.Task, userID int64, changes domain.Changes) error {
	allowed := EditableFields(role, task, userID)
	var forbidden []string
	for _, field := range allTaskFields {
		if _, changed := changes[field]; changed && !slices.Contains(allowed, field) {
			forbidden = append(forbidden, field)
		}
	}
	if len(forbidden) > 0 {
		return &domain.ForbiddenFieldsError{Fields: forbidden}
	}
	return nil
}

// CanInvite tells whether actor may add a member with the given role: owners hand out the
// admin and member roles, admins only member; the owner role is never granted this way.
func CanInvite(actor, target domain.Role) bool {
	switch actor {
	case domain.RoleOwner:
		return target == domain.RoleAdmin || target == domain.RoleMember
	case domain.RoleAdmin:
		return target == domain.RoleMember
	default:
		return false
	}
}

// CanViewStats limits team analytics to the owner and admins.
func CanViewStats(role domain.Role) bool {
	return role.CanManageTeam()
}

// memberRole resolves the actor's role in a team, telling a missing team apart from a user
// who simply is not in it.
func memberRole(ctx context.Context, st *repo.Store, teamID, userID int64) (domain.Role, error) {
	if _, err := st.TeamByID(ctx, teamID); err != nil {
		return "", err
	}
	return st.MemberRole(ctx, teamID, userID)
}
