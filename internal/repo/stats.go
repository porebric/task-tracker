package repo

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/porebric/task-tracker/internal/domain"
)

// teamStatsQuery is the whole report in one statement: the team's tasks are selected once
// into a CTE and every metric is an aggregate over it, so the cost stays a handful of index
// scans whatever the number of tasks.
const teamStatsQuery = `
WITH team_tasks AS (
    SELECT id, title, status, assignee_id, created_at, closed_at
    FROM tasks
    WHERE team_id = ?
),
by_status AS (
    SELECT status, COUNT(*) AS cnt
    FROM team_tasks
    GROUP BY status
),
closers AS (
    SELECT t.assignee_id AS user_id, u.name, COUNT(*) AS closed_cnt
    FROM team_tasks t
    JOIN users u ON u.id = t.assignee_id
    WHERE t.status = 'done'
      AND t.closed_at >= UTC_TIMESTAMP(6) - INTERVAL 30 DAY
    GROUP BY t.assignee_id, u.name
    ORDER BY closed_cnt DESC, t.assignee_id
    LIMIT 3
),
comments AS (
    SELECT t.id AS task_id, t.title, COUNT(c.id) AS cnt
    FROM team_tasks t
    LEFT JOIN task_comments c ON c.task_id = t.id
    GROUP BY t.id, t.title
)
SELECT
    (SELECT JSON_OBJECTAGG(status, cnt) FROM by_status) AS tasks_by_status,
    (SELECT JSON_ARRAYAGG(JSON_OBJECT('user_id', user_id, 'name', name, 'closed_tasks', closed_cnt))
       FROM closers) AS top_closers,
    (SELECT AVG(TIMESTAMPDIFF(SECOND, created_at, closed_at))
       FROM team_tasks
      WHERE closed_at IS NOT NULL) AS avg_close_seconds,
    (SELECT CAST(COALESCE(SUM(cnt), 0) AS SIGNED) FROM comments) AS comments_total,
    (SELECT JSON_ARRAYAGG(JSON_OBJECT('task_id', task_id, 'title', title, 'comments', cnt))
       FROM comments) AS comments_by_task
`

type statsRow struct {
	TasksByStatus   []byte
	TopClosers      []byte
	AvgCloseSeconds sql.NullFloat64
	CommentsTotal   int64
	CommentsByTask  []byte
}

func (s *Store) TeamStats(ctx context.Context, teamID int64) (*domain.TeamStats, error) {
	var row statsRow
	if err := s.q(ctx).Raw(teamStatsQuery, teamID).Scan(&row).Error; err != nil {
		return nil, fmt.Errorf("repo: team stats: %w", err)
	}

	stats := &domain.TeamStats{
		TeamID:        teamID,
		TasksByStatus: make(map[domain.Status]int64, len(domain.Statuses())),
		CommentsTotal: row.CommentsTotal,
	}
	for _, st := range domain.Statuses() {
		stats.TasksByStatus[st] = 0
	}
	if err := decodeJSON(row.TasksByStatus, &stats.TasksByStatus); err != nil {
		return nil, fmt.Errorf("repo: team stats: tasks_by_status: %w", err)
	}
	if err := decodeJSON(row.TopClosers, &stats.TopClosers); err != nil {
		return nil, fmt.Errorf("repo: team stats: top_closers: %w", err)
	}
	if err := decodeJSON(row.CommentsByTask, &stats.CommentsByTask); err != nil {
		return nil, fmt.Errorf("repo: team stats: comments_by_task: %w", err)
	}
	if row.AvgCloseSeconds.Valid {
		avg := row.AvgCloseSeconds.Float64
		stats.AvgCloseSeconds = &avg
	}

	// JSON_ARRAYAGG does not promise to keep the order of the aggregated rows.
	slices.SortFunc(stats.TopClosers, func(a, b domain.Closer) int {
		if c := cmp.Compare(b.ClosedTasks, a.ClosedTasks); c != 0 {
			return c
		}
		return cmp.Compare(a.UserID, b.UserID)
	})
	slices.SortFunc(stats.CommentsByTask, func(a, b domain.TaskCommentsCount) int {
		return cmp.Compare(a.TaskID, b.TaskID)
	})
	return stats, nil
}

func decodeJSON(raw []byte, dst any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dst)
}
