package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/fx"

	"github.com/porebric/healthstep-libs/app"
	"github.com/porebric/healthstep-libs/configs"
	"github.com/porebric/healthstep-libs/logger"
)

const (
	connectAttempts = 15
	connectDelay    = 2 * time.Second
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(newFromConfigs),
		fx.Invoke(start),
	)
}

func newFromConfigs(c *app.Configs, l *logger.Logger) (*DB, error) {
	ctx := logger.ToContext(c.Ctx, l)
	cfg := Config{
		DSN:             configs.Value(ctx, "db_dsn").String(),
		MaxOpenConns:    configs.Value(ctx, "db_max_open_conns").Int(),
		MaxIdleConns:    configs.Value(ctx, "db_max_idle_conns").Int(),
		ConnMaxLifetime: configs.Value(ctx, "db_conn_max_lifetime").Duration(),
		ConnMaxIdleTime: configs.Value(ctx, "db_conn_max_idle_time").Duration(),
		SlowThreshold:   configs.Value(ctx, "db_slow_threshold").Duration(),
		MigrationsPath:  configs.Value(ctx, "db_migrations_path").String(),
	}
	if cfg.DSN == "" {
		return nil, errors.New("mysql: db_dsn is not set")
	}
	return openWithRetry(ctx, cfg)
}

// openWithRetry tolerates a database that is still starting next to the service.
func openWithRetry(ctx context.Context, cfg Config) (*DB, error) {
	var lastErr error
	for attempt := 1; attempt <= connectAttempts; attempt++ {
		d, err := New(cfg)
		if err == nil {
			return d, nil
		}
		lastErr = err
		logger.Warn(ctx, "mysql: connect failed, retrying", "attempt", attempt, "error", err.Error())
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(connectDelay):
		}
	}
	return nil, fmt.Errorf("mysql: connect: %w", lastErr)
}

func start(lc fx.Lifecycle, d *DB, l *logger.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ctx = logger.ToContext(ctx, l)
			if err := d.Ping(ctx); err != nil {
				return err
			}
			if d.migrationsPath == "" {
				return nil
			}
			return d.Migrate(ctx, d.migrationsPath)
		},
		OnStop: func(context.Context) error { return d.Close() },
	})
}
