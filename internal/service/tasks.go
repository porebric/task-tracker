package service

import (
	"context"
	"time"

	"github.com/porebric/task-tracker/internal/cache"
	"github.com/porebric/task-tracker/internal/domain"
	"github.com/porebric/task-tracker/internal/repo"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

type Tasks struct {
	store *repo.Store
	lists *cache.TaskList
	now   func() time.Time
}

// NewTasks accepts a nil cache, in which case every list is served from the database.
func NewTasks(store *repo.Store, lists *cache.TaskList) *Tasks {
	return &Tasks{store: store, lists: lists, now: utcNow}
}

type CreateTaskInput struct {
	TeamID      int64
	Title       string
	Description string
	Status      domain.Status
	AssigneeID  *int64
}

// UpdateTaskInput carries only the fields present in the request.
type UpdateTaskInput struct {
	Title       *string
	Description *string
	Status      *domain.Status
	// AssigneeID of zero clears the assignee.
	AssigneeID *int64
	// Version, when given, must match the stored one.
	Version *int64
}

type TaskPage struct {
	Tasks []domain.Task `json:"tasks"`
	Total int64         `json:"total"`
	// NextCursor is set when the page is full; pass it as the cursor of the next request.
	NextCursor int64 `json:"next_cursor"`
}

func (s *Tasks) Create(ctx context.Context, actorID int64, in CreateTaskInput) (*domain.Task, error) {
	if _, err := memberRole(ctx, s.store, in.TeamID, actorID); err != nil {
		return nil, err
	}
	if err := checkAssignee(ctx, s.store, in.TeamID, in.AssigneeID); err != nil {
		return nil, err
	}

	now := s.now()
	task := &domain.Task{
		TeamID:      in.TeamID,
		Title:       in.Title,
		Description: in.Description,
		Status:      in.Status,
		CreatedBy:   actorID,
		AssigneeID:  in.AssigneeID,
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     1,
	}
	if task.Status == "" {
		task.Status = domain.StatusTodo
	}
	if task.Status == domain.StatusDone {
		task.ClosedAt = &now
	}

	err := s.store.Transaction(ctx, func(tx *repo.Store) error {
		if err := tx.CreateTask(ctx, task); err != nil {
			return err
		}
		return tx.AddHistory(ctx, &domain.TaskHistory{
			TaskID:    task.ID,
			ChangedBy: actorID,
			Changes:   creationChanges(task),
			CreatedAt: now,
		})
	})
	if err != nil {
		return nil, err
	}
	s.lists.Invalidate(ctx, task.TeamID)
	return task, nil
}

func (s *Tasks) Get(ctx context.Context, actorID, taskID int64) (*domain.Task, error) {
	task, err := accessTask(ctx, s.store, actorID, taskID)
	return task, err
}

func (s *Tasks) List(ctx context.Context, actorID int64, f repo.TaskFilter) (*TaskPage, error) {
	if _, err := memberRole(ctx, s.store, f.TeamID, actorID); err != nil {
		return nil, err
	}
	if f.Limit <= 0 {
		f.Limit = defaultListLimit
	}
	if f.Limit > maxListLimit {
		f.Limit = maxListLimit
	}

	key := cache.ListKey{
		TeamID:     f.TeamID,
		Status:     string(f.Status),
		AssigneeID: f.AssigneeID,
		Limit:      f.Limit,
		Offset:     f.Offset,
		Cursor:     f.Cursor,
	}
	var page TaskPage
	token, hit := s.lists.Get(ctx, key, &page)
	if hit {
		return &page, nil
	}

	tasks, total, err := s.store.ListTasks(ctx, f)
	if err != nil {
		return nil, err
	}
	if tasks == nil {
		tasks = []domain.Task{}
	}
	page = TaskPage{Tasks: tasks, Total: total}
	if len(tasks) > 0 && len(tasks) == f.Limit {
		page.NextCursor = tasks[len(tasks)-1].ID
	}
	s.lists.Set(ctx, key, token, &page)
	return &page, nil
}

// Update applies the changed fields under a row lock, so concurrent editors are serialized
// and every history entry is computed from the state the previous editor left.
func (s *Tasks) Update(ctx context.Context, actorID, taskID int64, in UpdateTaskInput) (*domain.Task, error) {
	var (
		result  *domain.Task
		changed bool
	)
	err := s.store.Transaction(ctx, func(tx *repo.Store) error {
		task, err := tx.TaskByIDForUpdate(ctx, taskID)
		if err != nil {
			return err
		}
		role, err := tx.MemberRole(ctx, task.TeamID, actorID)
		if err != nil {
			return err
		}
		if in.Version != nil && *in.Version != task.Version {
			return domain.ErrVersionConflict
		}

		changes := diffTask(task, in)
		if len(changes) == 0 {
			result = task
			return nil
		}
		if err := CheckEditable(role, task, actorID, changes); err != nil {
			return err
		}
		if _, ok := changes[FieldAssignee]; ok {
			if err := checkAssignee(ctx, tx, task.TeamID, assigneeOf(*in.AssigneeID)); err != nil {
				return err
			}
		}

		now := s.now()
		applyChanges(task, in, now)
		if err := tx.UpdateTask(ctx, task); err != nil {
			return err
		}
		if err := tx.AddHistory(ctx, &domain.TaskHistory{
			TaskID:    task.ID,
			ChangedBy: actorID,
			Changes:   changes,
			CreatedAt: now,
		}); err != nil {
			return err
		}
		result, changed = task, true
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Bumping after the commit is enough: a reader that fetched the generation earlier stores
	// its page under that older generation, which nobody reads once the bump lands.
	if changed {
		s.lists.Invalidate(ctx, result.TeamID)
	}
	return result, nil
}

func (s *Tasks) History(ctx context.Context, actorID, taskID int64) ([]domain.TaskHistory, error) {
	if _, err := accessTask(ctx, s.store, actorID, taskID); err != nil {
		return nil, err
	}
	return s.store.HistoryByTask(ctx, taskID)
}

func (s *Tasks) AddComment(ctx context.Context, actorID, taskID int64, content string) (*domain.TaskComment, error) {
	if _, err := accessTask(ctx, s.store, actorID, taskID); err != nil {
		return nil, err
	}
	comment := &domain.TaskComment{TaskID: taskID, UserID: actorID, Content: content, CreatedAt: s.now()}
	if err := s.store.AddComment(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *Tasks) Comments(ctx context.Context, actorID, taskID int64) ([]domain.TaskComment, error) {
	if _, err := accessTask(ctx, s.store, actorID, taskID); err != nil {
		return nil, err
	}
	return s.store.CommentsByTask(ctx, taskID)
}

// accessTask loads a task and makes sure the actor belongs to its team.
func accessTask(ctx context.Context, st *repo.Store, actorID, taskID int64) (*domain.Task, error) {
	task, err := st.TaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if _, err = st.MemberRole(ctx, task.TeamID, actorID); err != nil {
		return nil, err
	}
	return task, nil
}

func checkAssignee(ctx context.Context, st *repo.Store, teamID int64, assignee *int64) error {
	if assignee == nil {
		return nil
	}
	ok, err := st.IsMember(ctx, teamID, *assignee)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrAssigneeNotMember
	}
	return nil
}

func assigneeOf(id int64) *int64 {
	if id <= 0 {
		return nil
	}
	return &id
}

func diffTask(task *domain.Task, in UpdateTaskInput) domain.Changes {
	changes := domain.Changes{}
	if in.Title != nil && *in.Title != task.Title {
		changes[FieldTitle] = domain.FieldChange{Old: task.Title, New: *in.Title}
	}
	if in.Description != nil && *in.Description != task.Description {
		changes[FieldDescription] = domain.FieldChange{Old: task.Description, New: *in.Description}
	}
	if in.Status != nil && *in.Status != task.Status {
		changes[FieldStatus] = domain.FieldChange{Old: task.Status, New: *in.Status}
	}
	if in.AssigneeID != nil {
		next := assigneeOf(*in.AssigneeID)
		if !sameAssignee(task.AssigneeID, next) {
			changes[FieldAssignee] = domain.FieldChange{Old: idValue(task.AssigneeID), New: idValue(next)}
		}
	}
	return changes
}

func applyChanges(task *domain.Task, in UpdateTaskInput, now time.Time) {
	if in.Title != nil {
		task.Title = *in.Title
	}
	if in.Description != nil {
		task.Description = *in.Description
	}
	if in.AssigneeID != nil {
		task.AssigneeID = assigneeOf(*in.AssigneeID)
	}
	if in.Status != nil && *in.Status != task.Status {
		task.Status = *in.Status
		task.ClosedAt = nil
		if task.Status == domain.StatusDone {
			task.ClosedAt = &now
		}
	}
	task.UpdatedAt = now
}

func creationChanges(task *domain.Task) domain.Changes {
	return domain.Changes{
		FieldTitle:       {Old: nil, New: task.Title},
		FieldDescription: {Old: nil, New: task.Description},
		FieldStatus:      {Old: nil, New: task.Status},
		FieldAssignee:    {Old: nil, New: idValue(task.AssigneeID)},
	}
}

func sameAssignee(a, b *int64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func idValue(id *int64) any {
	if id == nil {
		return nil
	}
	return *id
}
