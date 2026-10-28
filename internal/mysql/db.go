package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	gomysql "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 5
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 5 * time.Minute
	defaultSlowThreshold   = 200 * time.Millisecond
)

type Config struct {
	DSN string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	SlowThreshold   time.Duration

	MigrationsPath string
}

type DB struct {
	gorm           *gorm.DB
	migrationsPath string
}

func New(cfg Config) (*DB, error) {
	if cfg.DSN == "" {
		return nil, errors.New("mysql: dsn is required")
	}
	dsn, err := normalizeDSN(cfg.DSN)
	if err != nil {
		return nil, err
	}

	g, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:         &gormLogger{slow: durationOr(cfg.SlowThreshold, defaultSlowThreshold)},
		NowFunc:        func() time.Time { return time.Now().UTC() },
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("mysql: open: %w", err)
	}

	sqlDB, err := g.DB()
	if err != nil {
		return nil, fmt.Errorf("mysql: %w", err)
	}
	sqlDB.SetMaxOpenConns(intOr(cfg.MaxOpenConns, defaultMaxOpenConns))
	sqlDB.SetMaxIdleConns(intOr(cfg.MaxIdleConns, defaultMaxIdleConns))
	sqlDB.SetConnMaxLifetime(durationOr(cfg.ConnMaxLifetime, defaultConnMaxLifetime))
	sqlDB.SetConnMaxIdleTime(durationOr(cfg.ConnMaxIdleTime, defaultConnMaxIdleTime))

	return &DB{gorm: g, migrationsPath: cfg.MigrationsPath}, nil
}

// normalizeDSN forces the driver settings the service relies on: time.Time values in UTC,
// matched-rows semantics for optimistic locking and no multi-statement queries.
func normalizeDSN(raw string) (string, error) {
	cfg, err := gomysql.ParseDSN(raw)
	if err != nil {
		return "", fmt.Errorf("mysql: parse dsn: %w", err)
	}
	cfg.ParseTime = true
	cfg.Loc = time.UTC
	cfg.ClientFoundRows = true
	cfg.MultiStatements = false
	return cfg.FormatDSN(), nil
}

func (d *DB) Gorm() *gorm.DB {
	return d.gorm
}

func (d *DB) Ping(ctx context.Context) error {
	sqlDB, err := d.gorm.DB()
	if err != nil {
		return fmt.Errorf("mysql: %w", err)
	}
	if err = sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("mysql: ping: %w", err)
	}
	return nil
}

func (d *DB) Close() error {
	sqlDB, err := d.gorm.DB()
	if err != nil {
		return fmt.Errorf("mysql: %w", err)
	}
	return sqlDB.Close()
}

func intOr(v, def int) int {
	if v > 0 {
		return v
	}
	return def
}

func durationOr(v, def time.Duration) time.Duration {
	if v > 0 {
		return v
	}
	return def
}
