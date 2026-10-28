package domain

import "time"

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

func ParseStatus(raw string) (Status, bool) {
	switch s := Status(raw); s {
	case StatusTodo, StatusInProgress, StatusDone:
		return s, true
	default:
		return "", false
	}
}

func Statuses() []Status {
	return []Status{StatusTodo, StatusInProgress, StatusDone}
}

type Task struct {
	ID          int64
	TeamID      int64
	Title       string
	Description string
	Status      Status
	CreatedBy   int64
	AssigneeID  *int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ClosedAt    *time.Time
	Version     int64
}

func (Task) TableName() string { return "tasks" }

func (t *Task) IsAssignee(userID int64) bool {
	return t.AssigneeID != nil && *t.AssigneeID == userID
}

type TaskComment struct {
	ID        int64
	TaskID    int64
	UserID    int64
	Content   string
	CreatedAt time.Time
}

func (TaskComment) TableName() string { return "task_comments" }
