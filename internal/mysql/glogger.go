package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/porebric/healthstep-libs/logger"
)

type gormLogger struct {
	slow time.Duration
}

func (l *gormLogger) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (*gormLogger) Info(ctx context.Context, msg string, args ...any) {
	logger.Info(ctx, fmt.Sprintf(msg, args...))
}

func (*gormLogger) Warn(ctx context.Context, msg string, args ...any) {
	logger.Warn(ctx, fmt.Sprintf(msg, args...))
}

func (*gormLogger) Error(ctx context.Context, msg string, args ...any) {
	logger.Error(ctx, fmt.Errorf(msg, args...), "mysql: gorm")
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		logger.Error(ctx, err, "mysql: query failed", "sql", sql, "rows", rows, "elapsed", elapsed.String())
	case l.slow > 0 && elapsed > l.slow:
		logger.Warn(ctx, "mysql: slow query", "sql", sql, "rows", rows, "elapsed", elapsed.String())
	default:
		logger.Debug(ctx, "mysql: query", "sql", sql, "rows", rows, "elapsed", elapsed.String())
	}
}
