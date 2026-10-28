package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker/v2"

	"github.com/porebric/healthstep-libs/logger"
)

const (
	defaultTTL         = 5 * time.Minute
	breakerName        = "redis-task-list"
	breakerFailures    = 3
	breakerOpenTimeout = 10 * time.Second
)

// ListKey identifies one page of a team's task list: every filter is part of the key.
type ListKey struct {
	TeamID     int64
	Status     string
	AssigneeID int64
	Limit      int
	Offset     int
	Cursor     int64
}

// TaskList caches pages of task lists in Redis. Entries live under a per-team generation:
// invalidation replaces the generation, so every page of the team becomes unreachable at
// once and expires on its own TTL. Redis failures never fail a request, and a circuit
// breaker keeps the service from waiting on a Redis that is down.
type TaskList struct {
	rdb     *redis.Client
	ttl     time.Duration
	breaker *gobreaker.CircuitBreaker[[]byte]
}

func NewTaskList(rdb *redis.Client, ttl time.Duration) *TaskList {
	if ttl <= 0 {
		ttl = defaultTTL
	}
	return &TaskList{
		rdb: rdb,
		ttl: ttl,
		breaker: gobreaker.NewCircuitBreaker[[]byte](gobreaker.Settings{
			Name:    breakerName,
			Timeout: breakerOpenTimeout,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= breakerFailures
			},
		}),
	}
}

// Get loads the page for key into dst. The returned token names the generation observed
// before the lookup and must be handed to Set, so a page loaded from the database is never
// stored under a generation newer than the data it holds.
func (c *TaskList) Get(ctx context.Context, key ListKey, dst any) (token string, hit bool) {
	if c == nil {
		return "", false
	}
	var generation string
	raw, err := c.breaker.Execute(func() ([]byte, error) {
		var err error
		if generation, err = c.generation(ctx, key.TeamID); err != nil {
			return nil, err
		}
		b, err := c.rdb.Get(ctx, entryKey(key, generation)).Bytes()
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return b, err
	})
	if err != nil {
		logger.Warn(ctx, "cache: get failed", "error", err.Error())
		return "", false
	}
	if raw == nil {
		return generation, false
	}
	if err = json.Unmarshal(raw, dst); err != nil {
		logger.Warn(ctx, "cache: corrupted entry ignored", "key", entryKey(key, generation), "error", err.Error())
		return generation, false
	}
	return generation, true
}

func (c *TaskList) Set(ctx context.Context, key ListKey, token string, value any) {
	if c == nil || token == "" {
		return
	}
	raw, err := json.Marshal(value)
	if err != nil {
		logger.Warn(ctx, "cache: encode failed", "error", err.Error())
		return
	}
	if _, err = c.breaker.Execute(func() ([]byte, error) {
		return nil, c.rdb.Set(ctx, entryKey(key, token), raw, c.ttl).Err()
	}); err != nil {
		logger.Warn(ctx, "cache: set failed", "error", err.Error())
	}
}

// Invalidate starts a new generation for the team; pages of older generations are never read again.
func (c *TaskList) Invalidate(ctx context.Context, teamID int64) {
	if c == nil {
		return
	}
	if _, err := c.breaker.Execute(func() ([]byte, error) {
		return nil, c.rdb.Set(ctx, generationKey(teamID), newGeneration(), 0).Err()
	}); err != nil {
		logger.Warn(ctx, "cache: invalidate failed", "team_id", teamID, "error", err.Error())
	}
}

func (c *TaskList) generation(ctx context.Context, teamID int64) (string, error) {
	key := generationKey(teamID)
	current, err := c.rdb.Get(ctx, key).Result()
	if err == nil {
		return current, nil
	}
	if !errors.Is(err, redis.Nil) {
		return "", err
	}
	fresh := newGeneration()
	created, err := c.rdb.SetNX(ctx, key, fresh, 0).Result()
	if err != nil {
		return "", err
	}
	if created {
		return fresh, nil
	}
	return c.rdb.Get(ctx, key).Result()
}

func newGeneration() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func generationKey(teamID int64) string {
	return fmt.Sprintf("tasks:team:%d:generation", teamID)
}

func entryKey(key ListKey, generation string) string {
	return fmt.Sprintf("tasks:team:%d:%s:status=%s:assignee=%d:limit=%d:offset=%d:cursor=%d",
		key.TeamID, generation, key.Status, key.AssigneeID, key.Limit, key.Offset, key.Cursor)
}
