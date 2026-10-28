package mysql

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"

	"github.com/porebric/healthstep-libs/logger"
)

const (
	migrationsTable = "schema_migrations"

	createMigrationsTable = `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    VARCHAR(255) NOT NULL,
		applied_at DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
		PRIMARY KEY (version)
	)`
)

// Migrate applies every *.sql file from dir that is not yet recorded in schema_migrations,
// in lexical order. A file is a sequence of statements separated by semicolons.
func (d *DB) Migrate(ctx context.Context, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("mysql: read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	if err = d.gorm.WithContext(ctx).Exec(createMigrationsTable).Error; err != nil {
		return fmt.Errorf("mysql: ensure migrations table: %w", err)
	}

	applied, err := d.appliedMigrations(ctx)
	if err != nil {
		return err
	}

	for _, name := range files {
		version := strings.TrimSuffix(name, ".sql")
		if _, ok := applied[version]; ok {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("mysql: read migration %s: %w", name, err)
		}
		if err = d.applyMigration(ctx, version, splitStatements(string(content))); err != nil {
			return fmt.Errorf("mysql: apply migration %s: %w", name, err)
		}
		logger.Info(ctx, "mysql: migration applied", "version", version)
	}
	return nil
}

func (d *DB) appliedMigrations(ctx context.Context) (map[string]struct{}, error) {
	var versions []string
	if err := d.gorm.WithContext(ctx).Table(migrationsTable).Pluck("version", &versions).Error; err != nil {
		return nil, fmt.Errorf("mysql: load applied migrations: %w", err)
	}
	applied := make(map[string]struct{}, len(versions))
	for _, v := range versions {
		applied[v] = struct{}{}
	}
	return applied, nil
}

// applyMigration runs the statements one by one. MySQL commits DDL implicitly, so the
// transaction only guards the version record of a migration made of DML statements.
func (d *DB) applyMigration(ctx context.Context, version string, statements []string) error {
	return d.gorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, stmt := range statements {
			if err := tx.Exec(stmt).Error; err != nil {
				return err
			}
		}
		return tx.Table(migrationsTable).Create(map[string]any{"version": version}).Error
	})
}

// splitStatements cuts a script into statements on semicolons outside of quoted literals,
// honouring backslash escapes inside them and dropping "--" comments and blank lines.
func splitStatements(script string) []string {
	sc := &statementScanner{}
	for line := range strings.SplitSeq(script, "\n") {
		sc.feed(line)
	}
	sc.flush()
	return sc.statements
}

type statementScanner struct {
	statements []string
	current    strings.Builder
	quote      rune
}

func (sc *statementScanner) flush() {
	if s := strings.TrimSpace(sc.current.String()); s != "" {
		sc.statements = append(sc.statements, s)
	}
	sc.current.Reset()
}

func (sc *statementScanner) feed(line string) {
	trimmed := strings.TrimSpace(line)
	if sc.quote == 0 && (trimmed == "" || strings.HasPrefix(trimmed, "--")) {
		return
	}
	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		switch {
		case sc.quote != 0:
			i = sc.quoted(runes, i)
		case ch == '-' && i+1 < len(runes) && runes[i+1] == '-':
			i = len(runes)
		case ch == '\'' || ch == '"' || ch == '`':
			sc.quote = ch
			sc.current.WriteRune(ch)
		case ch == ';':
			sc.flush()
		default:
			sc.current.WriteRune(ch)
		}
	}
	sc.current.WriteByte('\n')
}

// quoted copies the rune at i inside a quoted literal and returns the index of the last rune
// consumed: an escaped character takes two.
func (sc *statementScanner) quoted(runes []rune, i int) int {
	ch := runes[i]
	sc.current.WriteRune(ch)
	switch {
	case ch == '\\' && i+1 < len(runes):
		sc.current.WriteRune(runes[i+1])
		return i + 1
	case ch == sc.quote:
		sc.quote = 0
	}
	return i
}
