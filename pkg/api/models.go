package api

import "time"

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Team struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type TeamMembership struct {
	Team
	Role string `json:"role"`
}

type Task struct {
	ID          int64      `json:"id"`
	TeamID      int64      `json:"team_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedBy   int64      `json:"created_by"`
	AssigneeID  *int64     `json:"assignee_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ClosedAt    *time.Time `json:"closed_at"`
	Version     int64      `json:"version"`
}

type FieldChange struct {
	Old any `json:"old"`
	New any `json:"new"`
}

type HistoryEntry struct {
	ID        int64                  `json:"id"`
	TaskID    int64                  `json:"task_id"`
	ChangedBy int64                  `json:"changed_by"`
	Changes   map[string]FieldChange `json:"changes"`
	CreatedAt time.Time              `json:"created_at"`
}

type Comment struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Closer struct {
	UserID      int64  `json:"user_id"`
	Name        string `json:"name"`
	ClosedTasks int64  `json:"closed_tasks"`
}

type TaskComments struct {
	TaskID   int64  `json:"task_id"`
	Title    string `json:"title"`
	Comments int64  `json:"comments"`
}

type CommentsStats struct {
	Total  int64          `json:"total"`
	ByTask []TaskComments `json:"by_task"`
}

type TeamStats struct {
	TeamID          int64            `json:"team_id"`
	TasksByStatus   map[string]int64 `json:"tasks_by_status"`
	TopClosers      []Closer         `json:"top_closers"`
	AvgCloseSeconds *float64         `json:"avg_close_seconds"`
	Comments        CommentsStats    `json:"comments"`
}
