package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/porebric/healthstep-libs/app"
	"github.com/porebric/healthstep-libs/configs"
	"github.com/porebric/healthstep-libs/logger"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(newClient, newTaskList),
		fx.Invoke(start),
	)
}

func newClient(c *app.Configs) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     configs.Value(c.Ctx, "redis_addr").String(),
		Password: configs.Value(c.Ctx, "redis_password").String(),
		DB:       configs.Value(c.Ctx, "redis_db").Int(),
	})
}

func newTaskList(c *app.Configs, rdb *redis.Client) *TaskList {
	return NewTaskList(rdb, configs.Value(c.Ctx, "cache_task_list_ttl").Duration())
}

func start(lc fx.Lifecycle, rdb *redis.Client, l *logger.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := rdb.Ping(ctx).Err(); err != nil {
				return fmt.Errorf("cache: redis ping: %w", err)
			}
			logger.Info(logger.ToContext(ctx, l), "cache: redis connected", "addr", rdb.Options().Addr)
			return nil
		},
		OnStop: func(context.Context) error { return rdb.Close() },
	})
}
