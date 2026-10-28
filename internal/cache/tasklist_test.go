package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntryKey_IncludesEveryFilter(t *testing.T) {
	base := ListKey{TeamID: 1, Status: "todo", AssigneeID: 2, Limit: 20}
	variants := []ListKey{
		{TeamID: 2, Status: "todo", AssigneeID: 2, Limit: 20},
		{TeamID: 1, Status: "done", AssigneeID: 2, Limit: 20},
		{TeamID: 1, Status: "todo", AssigneeID: 3, Limit: 20},
		{TeamID: 1, Status: "todo", AssigneeID: 2, Limit: 10},
		{TeamID: 1, Status: "todo", AssigneeID: 2, Limit: 20, Offset: 20},
		{TeamID: 1, Status: "todo", AssigneeID: 2, Limit: 20, Cursor: 5},
	}
	for _, v := range variants {
		assert.NotEqual(t, entryKey(base, "g"), entryKey(v, "g"))
	}
	assert.NotEqual(t, entryKey(base, "g1"), entryKey(base, "g2"))
}

func TestTaskList_NilIsNoop(t *testing.T) {
	var c *TaskList
	var page []string
	token, hit := c.Get(context.Background(), ListKey{TeamID: 1}, &page)
	assert.Empty(t, token)
	assert.False(t, hit)
	c.Set(context.Background(), ListKey{TeamID: 1}, "g", page)
	c.Invalidate(context.Background(), 1)
}

func TestTaskList_Generations(t *testing.T) {
	srv := miniredis.RunT(t)
	c := NewTaskList(redis.NewClient(&redis.Options{Addr: srv.Addr()}), time.Minute)
	ctx := context.Background()
	key := ListKey{TeamID: 1, Limit: 20}
	var got []string

	token, hit := c.Get(ctx, key, &got)
	require.NotEmpty(t, token)
	assert.False(t, hit)

	c.Set(ctx, key, token, []string{"a"})
	_, hit = c.Get(ctx, key, &got)
	assert.True(t, hit)
	assert.Equal(t, []string{"a"}, got)

	// A page loaded before an invalidation never becomes visible after it.
	otherKey := ListKey{TeamID: 1, Limit: 10}
	stale, _ := c.Get(ctx, otherKey, &got)
	c.Invalidate(ctx, 1)
	c.Set(ctx, otherKey, stale, []string{"old"})
	_, hit = c.Get(ctx, otherKey, &got)
	assert.False(t, hit)
	_, hit = c.Get(ctx, key, &got)
	assert.False(t, hit)

	// Other teams keep their pages.
	foreign := ListKey{TeamID: 2, Limit: 20}
	foreignToken, _ := c.Get(ctx, foreign, &got)
	c.Set(ctx, foreign, foreignToken, []string{"b"})
	c.Invalidate(ctx, 1)
	_, hit = c.Get(ctx, foreign, &got)
	assert.True(t, hit)
	assert.Equal(t, []string{"b"}, got)

	// Entries expire on their own.
	srv.FastForward(2 * time.Minute)
	_, hit = c.Get(ctx, foreign, &got)
	assert.False(t, hit)
}

func TestTaskList_FailsOpenWhenRedisIsDown(t *testing.T) {
	srv := miniredis.RunT(t)
	c := NewTaskList(redis.NewClient(&redis.Options{
		Addr:        srv.Addr(),
		MaxRetries:  -1,
		DialTimeout: 100 * time.Millisecond,
	}), time.Minute)
	srv.Close()

	var got []string
	for range breakerFailures {
		token, hit := c.Get(context.Background(), ListKey{TeamID: 1}, &got)
		assert.Empty(t, token)
		assert.False(t, hit)
	}
	assert.Equal(t, gobreaker.StateOpen, c.breaker.State())
	c.Set(context.Background(), ListKey{TeamID: 1}, "g", got)
	c.Invalidate(context.Background(), 1)
}

func TestTaskList_ReplaysInvalidationAfterOutage(t *testing.T) {
	srv := miniredis.RunT(t)
	c := NewTaskList(redis.NewClient(&redis.Options{
		Addr:        srv.Addr(),
		MaxRetries:  -1,
		DialTimeout: 100 * time.Millisecond,
	}), time.Minute)
	ctx := context.Background()
	key := ListKey{TeamID: 1, Limit: 20}
	var got []string

	token, _ := c.Get(ctx, key, &got)
	c.Set(ctx, key, token, []string{"before"})

	srv.Close()
	c.Invalidate(ctx, 1)
	require.NoError(t, srv.Restart())

	_, hit := c.Get(ctx, key, &got)
	assert.False(t, hit, "a page cached before the outage must not survive a lost invalidation")

	token, _ = c.Get(ctx, key, &got)
	c.Set(ctx, key, token, []string{"after"})
	_, hit = c.Get(ctx, key, &got)
	assert.True(t, hit)
	assert.Equal(t, []string{"after"}, got)
}
