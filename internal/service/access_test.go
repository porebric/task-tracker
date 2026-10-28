package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/porebric/task-tracker/internal/domain"
	"github.com/porebric/task-tracker/internal/service"
)

func TestEditableFields(t *testing.T) {
	assignee := int64(7)
	task := &domain.Task{CreatedBy: 3, AssigneeID: &assignee}
	all := []string{service.FieldTitle, service.FieldDescription, service.FieldStatus, service.FieldAssignee}

	tests := []struct {
		name string
		role domain.Role
		user int64
		want []string
	}{
		{name: "owner edits everything", role: domain.RoleOwner, user: 99, want: all},
		{name: "admin edits everything", role: domain.RoleAdmin, user: 99, want: all},
		{name: "creator edits own task", role: domain.RoleMember, user: 3, want: all},
		{name: "assignee changes status only", role: domain.RoleMember, user: 7, want: []string{service.FieldStatus}},
		{name: "other member edits nothing", role: domain.RoleMember, user: 42, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, service.EditableFields(tt.role, task, tt.user))
		})
	}
}

func TestCheckEditable(t *testing.T) {
	assignee := int64(7)
	task := &domain.Task{CreatedBy: 3, AssigneeID: &assignee}
	changes := domain.Changes{
		service.FieldTitle:    {Old: "a", New: "b"},
		service.FieldStatus:   {Old: "todo", New: "done"},
		service.FieldAssignee: {Old: 7, New: 3},
	}

	require.NoError(t, service.CheckEditable(domain.RoleAdmin, task, 99, changes))
	require.NoError(t, service.CheckEditable(domain.RoleOwner, task, 99, changes))
	require.NoError(t, service.CheckEditable(domain.RoleMember, task, 3, changes))

	var forbidden *domain.ForbiddenFieldsError
	err := service.CheckEditable(domain.RoleMember, task, 7, changes)
	require.ErrorAs(t, err, &forbidden)
	assert.Equal(t, []string{service.FieldTitle, service.FieldAssignee}, forbidden.Fields)

	statusOnly := domain.Changes{service.FieldStatus: {Old: "todo", New: "done"}}
	require.NoError(t, service.CheckEditable(domain.RoleMember, task, 7, statusOnly))
	err = service.CheckEditable(domain.RoleMember, task, 42, statusOnly)
	require.ErrorAs(t, err, &forbidden)
	assert.Equal(t, []string{service.FieldStatus}, forbidden.Fields)
}

func TestCanInvite(t *testing.T) {
	tests := []struct {
		actor, target domain.Role
		want          bool
	}{
		{domain.RoleOwner, domain.RoleAdmin, true},
		{domain.RoleOwner, domain.RoleMember, true},
		{domain.RoleOwner, domain.RoleOwner, false},
		{domain.RoleAdmin, domain.RoleMember, true},
		{domain.RoleAdmin, domain.RoleAdmin, false},
		{domain.RoleAdmin, domain.RoleOwner, false},
		{domain.RoleMember, domain.RoleMember, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.actor)+" invites "+string(tt.target), func(t *testing.T) {
			assert.Equal(t, tt.want, service.CanInvite(tt.actor, tt.target))
		})
	}
}

func TestCanViewStats(t *testing.T) {
	assert.True(t, service.CanViewStats(domain.RoleOwner))
	assert.True(t, service.CanViewStats(domain.RoleAdmin))
	assert.False(t, service.CanViewStats(domain.RoleMember))
}
