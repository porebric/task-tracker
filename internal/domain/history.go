package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type FieldChange struct {
	Old any `json:"old"`
	New any `json:"new"`
}

// Changes maps a task field name to its value before and after the change.
type Changes map[string]FieldChange

// Value stores the changes as a JSON document. A string is returned on purpose: MySQL
// refuses binary strings for JSON columns.
func (c Changes) Value() (driver.Value, error) {
	raw, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("history: encode changes: %w", err)
	}
	return string(raw), nil
}

func (c *Changes) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*c = nil
		return nil
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	default:
		return fmt.Errorf("history: cannot scan %T into Changes", src)
	}
}

type TaskHistory struct {
	ID        int64
	TaskID    int64
	ChangedBy int64
	Changes   Changes
	CreatedAt time.Time
}

func (TaskHistory) TableName() string { return "task_history" }
