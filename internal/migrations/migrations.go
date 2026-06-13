package migrations

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

//go:embed sql/*.up.sql
var migrationFiles embed.FS

type Conn interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

func Apply(ctx context.Context, conn Conn) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migrations: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		create table if not exists schema_migrations (
			version text primary key,
			applied_at timestamptz not null default now()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := migrationFiles.ReadDir("sql")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		version := strings.TrimSuffix(name, ".up.sql")
		var alreadyApplied bool
		if err := tx.QueryRow(ctx, "select exists(select 1 from schema_migrations where version = $1)", version).Scan(&alreadyApplied); err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if alreadyApplied {
			continue
		}

		sqlBytes, err := migrationFiles.ReadFile("sql/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("apply migration %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx, "insert into schema_migrations(version) values ($1)", version); err != nil {
			return fmt.Errorf("record migration %s: %w", version, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migrations: %w", err)
	}
	return nil
}
