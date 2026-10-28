package handler

import (
	"github.com/porebric/task-tracker/internal/domain"
	"github.com/porebric/task-tracker/pkg/api"
)

func toUser(u *domain.User) api.User {
	return api.User{ID: u.ID, Email: u.Email, Name: u.Name, CreatedAt: u.CreatedAt}
}

func toTeam(t *domain.Team) api.Team {
	return api.Team{ID: t.ID, Name: t.Name, CreatedBy: t.CreatedBy, CreatedAt: t.CreatedAt}
}

func toMembership(m domain.TeamMembership) api.TeamMembership {
	return api.TeamMembership{
		Team: api.Team{ID: m.ID, Name: m.Name, CreatedBy: m.CreatedBy, CreatedAt: m.CreatedAt},
		Role: string(m.Role),
	}
}

func toTask(t *domain.Task) api.Task {
	return api.Task{
		ID:          t.ID,
		TeamID:      t.TeamID,
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		CreatedBy:   t.CreatedBy,
		AssigneeID:  t.AssigneeID,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		ClosedAt:    t.ClosedAt,
		Version:     t.Version,
	}
}

func toTasks(tasks []domain.Task) []api.Task {
	out := make([]api.Task, 0, len(tasks))
	for i := range tasks {
		out = append(out, toTask(&tasks[i]))
	}
	return out
}

func toHistory(entries []domain.TaskHistory) []api.HistoryEntry {
	out := make([]api.HistoryEntry, 0, len(entries))
	for _, e := range entries {
		changes := make(map[string]api.FieldChange, len(e.Changes))
		for field, change := range e.Changes {
			changes[field] = api.FieldChange{Old: change.Old, New: change.New}
		}
		out = append(out, api.HistoryEntry{
			ID:        e.ID,
			TaskID:    e.TaskID,
			ChangedBy: e.ChangedBy,
			Changes:   changes,
			CreatedAt: e.CreatedAt,
		})
	}
	return out
}

func toComment(c *domain.TaskComment) api.Comment {
	return api.Comment{ID: c.ID, TaskID: c.TaskID, UserID: c.UserID, Content: c.Content, CreatedAt: c.CreatedAt}
}

func toComments(comments []domain.TaskComment) []api.Comment {
	out := make([]api.Comment, 0, len(comments))
	for i := range comments {
		out = append(out, toComment(&comments[i]))
	}
	return out
}

func toStats(s *domain.TeamStats) api.TeamStats {
	byStatus := make(map[string]int64, len(s.TasksByStatus))
	for status, count := range s.TasksByStatus {
		byStatus[string(status)] = count
	}
	closers := make([]api.Closer, 0, len(s.TopClosers))
	for _, c := range s.TopClosers {
		closers = append(closers, api.Closer{UserID: c.UserID, Name: c.Name, ClosedTasks: c.ClosedTasks})
	}
	byTask := make([]api.TaskComments, 0, len(s.CommentsByTask))
	for _, c := range s.CommentsByTask {
		byTask = append(byTask, api.TaskComments{TaskID: c.TaskID, Title: c.Title, Comments: c.Comments})
	}
	return api.TeamStats{
		TeamID:          s.TeamID,
		TasksByStatus:   byStatus,
		TopClosers:      closers,
		AvgCloseSeconds: s.AvgCloseSeconds,
		Comments:        api.CommentsStats{Total: s.CommentsTotal, ByTask: byTask},
	}
}
