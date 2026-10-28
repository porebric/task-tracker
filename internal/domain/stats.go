package domain

type Closer struct {
	UserID      int64  `json:"user_id"`
	Name        string `json:"name"`
	ClosedTasks int64  `json:"closed_tasks"`
}

type TaskCommentsCount struct {
	TaskID   int64  `json:"task_id"`
	Title    string `json:"title"`
	Comments int64  `json:"comments"`
}

type TeamStats struct {
	TeamID          int64
	TasksByStatus   map[Status]int64
	TopClosers      []Closer
	AvgCloseSeconds *float64
	CommentsTotal   int64
	CommentsByTask  []TaskCommentsCount
}
