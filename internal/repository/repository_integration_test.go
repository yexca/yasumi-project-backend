package repository_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yasumi/yasumi-project-backend/internal/config"
	"github.com/yasumi/yasumi-project-backend/internal/migrations"
	"github.com/yasumi/yasumi-project-backend/internal/repository"
)

const (
	userA  = "00000000-0000-4000-8000-000000000001"
	userB  = "00000000-0000-4000-8000-000000000002"
	itemA  = "10000000-0000-4000-8000-000000000001"
	itemB  = "10000000-0000-4000-8000-000000000002"
	tplA   = "20000000-0000-4000-8000-000000000001"
	opA    = "30000000-0000-4000-8000-000000000001"
	device = "device-alpha"
)

func TestMigrationsEnforceMinimumConstraints(t *testing.T) {
	pool := newTestPool(t)
	ctx := context.Background()

	validItemSQL := `
		insert into items (
			id, user_id, item_type, title, status, created_at, updated_at,
			client_updated_at, server_updated_at, created_by_device_id,
			updated_by_device_id, revision
		) values (
			$1, $2, 'inbox', 'Capture', 'active', now(), now(),
			now(), now(), $3, $3, 1
		)
	`

	if _, err := pool.Exec(ctx, validItemSQL, itemA, userA, device); err != nil {
		t.Fatalf("insert valid item: %v", err)
	}

	_, err := pool.Exec(ctx, `
		insert into items (
			id, user_id, item_type, title, status, created_at, updated_at,
			client_updated_at, server_updated_at, created_by_device_id,
			updated_by_device_id, revision
		) values (
			'10000000-0000-4000-8000-000000000099', $1, 'task', 'Bad type', 'active',
			now(), now(), now(), now(), $2, $2, 1
		)
	`, userA, device)
	assertCheckViolation(t, err)

	_, err = pool.Exec(ctx, `
		insert into items (
			id, user_id, item_type, title, status, created_at, updated_at,
			client_updated_at, server_updated_at, created_by_device_id,
			updated_by_device_id, revision
		) values (
			'10000000-0000-4000-8000-000000000098', $1, 'inbox', '   ', 'active',
			now(), now(), now(), now(), $2, $2, 1
		)
	`, userA, device)
	assertCheckViolation(t, err)

	if _, err := pool.Exec(ctx, `
		insert into items (
			id, user_id, item_type, title, status, deadline_date, deadline_time,
			deadline_at, deadline_time_zone_mode, created_at, updated_at,
			client_updated_at, server_updated_at, created_by_device_id,
			updated_by_device_id, revision
		) values (
			'10000000-0000-4000-8000-000000000003', $1, 'deadline_task', 'Fixed deadline',
			'active', null, null, now(), 'fixed', now(), now(), now(), now(), $2, $2, 1
		)
	`, userA, device); err != nil {
		t.Fatalf("insert valid fixed deadline: %v", err)
	}

	_, err = pool.Exec(ctx, `
		insert into items (
			id, user_id, item_type, title, status, deadline_date, deadline_time,
			deadline_at, deadline_time_zone_mode, created_at, updated_at,
			client_updated_at, server_updated_at, created_by_device_id,
			updated_by_device_id, revision
		) values (
			'10000000-0000-4000-8000-000000000004', $1, 'deadline_task', 'Mixed deadline',
			'active', current_date, '09:00', now(), 'fixed', now(), now(), now(), now(), $2, $2, 1
		)
	`, userA, device)
	assertCheckViolation(t, err)
}

func TestMigrationsEnforceRecurrenceAndOperationUniqueness(t *testing.T) {
	pool := newTestPool(t)
	ctx := context.Background()

	insertTemplate(t, pool, tplA, userA)
	insertRecurringItem(t, pool, itemA, userA, tplA, 1)

	var uniqueErr *pgconn.PgError
	_, err := pool.Exec(ctx, `
		insert into items (
			id, user_id, item_type, title, status, recurring_template_id, recurring_sequence,
			created_at, updated_at, client_updated_at, server_updated_at,
			created_by_device_id, updated_by_device_id, revision
		) values (
			'10000000-0000-4000-8000-000000000005', $1, 'date_task', 'Duplicate',
			'active', $2, 1, now(), now(), now(), now(), $3, $3, 1
		)
	`, userA, tplA, device)
	if !errors.As(err, &uniqueErr) || uniqueErr.Code != "23505" {
		t.Fatalf("duplicate recurring instance error = %v, want unique violation", err)
	}

	insertOperation(t, pool, opA, userA, "same-key")
	_, err = pool.Exec(ctx, `
		insert into operation_history (
			id, user_id, event_type, idempotency_key, created_at, created_by_device_id
		) values (
			'30000000-0000-4000-8000-000000000002', $1, 'completed', 'same-key', now(), $2
		)
	`, userA, device)
	if !errors.As(err, &uniqueErr) || uniqueErr.Code != "23505" {
		t.Fatalf("duplicate idempotency error = %v, want unique violation", err)
	}

	_, err = pool.Exec(ctx, `
		insert into recurring_task_templates (
			id, user_id, title, frequency, interval, recurrence_basis, start_date, end_type,
			completed_count, next_sequence, status, created_at, updated_at,
			client_updated_at, server_updated_at, created_by_device_id, updated_by_device_id, revision
		) values (
			'20000000-0000-4000-8000-000000000009', $1, 'Bad recurrence', 'weekly',
			0, 'scheduled_date', current_date, 'never', 0, 1, 'active',
			now(), now(), now(), now(), $2, $2, 1
		)
	`, userA, device)
	assertCheckViolation(t, err)
}

func TestOperationHistoryIsAppendOnly(t *testing.T) {
	pool := newTestPool(t)
	ctx := context.Background()

	insertOperation(t, pool, opA, userA, "append-only")

	_, err := pool.Exec(ctx, "update operation_history set reason = 'changed' where id = $1", opA)
	if err == nil {
		t.Fatal("update operation_history error = nil, want append-only rejection")
	}

	_, err = pool.Exec(ctx, "delete from operation_history where id = $1", opA)
	if err == nil {
		t.Fatal("delete operation_history error = nil, want append-only rejection")
	}
}

func TestRepositoryScopesItemQueriesByUser(t *testing.T) {
	pool := newTestPool(t)
	ctx := context.Background()
	repo := repository.New(pool)

	insertBaseItem(t, pool, itemA, userA, "A item")
	insertBaseItem(t, pool, itemB, userB, "B item")

	if err := repo.InTx(ctx, func(ctx context.Context, tx *repository.Tx) error {
		row, err := tx.GetItemForUpdateByUser(ctx, userA, itemA)
		if err != nil {
			return err
		}
		if row.UserID != userA || row.ID != itemA {
			t.Fatalf("scoped row = %+v", row)
		}

		_, err = tx.GetItemForUpdateByUser(ctx, userA, itemB)
		if !errors.Is(err, repository.ErrNotFound) {
			t.Fatalf("cross-user lookup error = %v, want ErrNotFound", err)
		}
		return nil
	}); err != nil {
		t.Fatalf("repository transaction: %v", err)
	}
}

func TestMigrationsRejectCrossUserReferences(t *testing.T) {
	pool := newTestPool(t)
	ctx := context.Background()

	areaID := "40000000-0000-4000-8000-000000000001"
	insertArea(t, pool, areaID, userA)

	_, err := pool.Exec(ctx, `
		insert into items (
			id, user_id, item_type, title, status, area_id, created_at, updated_at,
			client_updated_at, server_updated_at, created_by_device_id,
			updated_by_device_id, revision
		) values (
			'10000000-0000-4000-8000-000000000010', $1, 'inbox', 'Cross-user area',
			'active', $2, now(), now(), now(), now(), $3, $3, 1
		)
	`, userB, areaID, device)
	assertForeignKeyViolation(t, err)

	insertTemplate(t, pool, tplA, userA)
	_, err = pool.Exec(ctx, `
		insert into items (
			id, user_id, item_type, title, status, scheduled_date,
			recurring_template_id, recurring_sequence, created_at, updated_at,
			client_updated_at, server_updated_at, created_by_device_id,
			updated_by_device_id, revision
		) values (
			'10000000-0000-4000-8000-000000000011', $1, 'date_task', 'Cross-user recurrence',
			'active', current_date, $2, 1, now(), now(), now(), now(), $3, $3, 1
		)
	`, userB, tplA, device)
	assertForeignKeyViolation(t, err)
}

func TestRepositoryOperationIdempotencyLookupAndSettingsUpsert(t *testing.T) {
	pool := newTestPool(t)
	ctx := context.Background()
	repo := repository.New(pool)
	now := time.Now().UTC().Truncate(time.Second)
	idempotencyKey := "action:00000000-0000-4000-8000-000000000001:device-alpha:test"

	if err := repo.InTx(ctx, func(ctx context.Context, tx *repository.Tx) error {
		if err := tx.InsertOperationHistory(ctx, repository.OperationHistoryRecord{
			ID:                opA,
			UserID:            userA,
			EventType:         "completed",
			IdempotencyKey:    &idempotencyKey,
			CreatedAt:         now,
			CreatedByDeviceID: device,
		}); err != nil {
			return err
		}
		found, err := tx.FindOperationByIdempotencyKey(ctx, userA, idempotencyKey)
		if err != nil {
			return err
		}
		if found.ID != opA {
			t.Fatalf("operation id = %q, want %q", found.ID, opA)
		}

		return tx.UpsertUserSettings(ctx, repository.UserSettingsRecord{
			UserID:                    userA,
			Language:                  "ja",
			Locale:                    "ja-JP",
			WeekStartDay:              "monday",
			TimeZone:                  "Asia/Tokyo",
			DateDisplayFormat:         "yyyy-MM-dd",
			TimeDisplayFormat:         "24h",
			DefaultTimeZoneMode:       "floating",
			TodayPrimaryLookaheadDays: 3,
			DeadlineAwarenessDays:     14,
			CreatedAt:                 now,
			UpdatedAt:                 now,
			ClientUpdatedAt:           now,
			ServerUpdatedAt:           now,
			CreatedByDeviceID:         device,
			UpdatedByDeviceID:         device,
			Revision:                  1,
		})
	}); err != nil {
		t.Fatalf("repository transaction: %v", err)
	}
}

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	dsn := os.Getenv("YASUMI_TEST_DATABASE_URL")
	if dsn == "" {
		cfg := config.MustLoad()
		dsn = cfg.Postgres.DSN()
	}

	adminPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("open postgres pool: %v", err)
	}
	if err := adminPool.Ping(ctx); err != nil {
		adminPool.Close()
		t.Skipf("PostgreSQL integration tests skipped: %v", err)
	}

	schemaName := "test_" + strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
	schemaName = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, schemaName)

	quotedSchema := pgxIdentifier(schemaName)
	if _, err := adminPool.Exec(ctx, "drop schema if exists "+quotedSchema+" cascade"); err != nil {
		t.Fatalf("drop test schema: %v", err)
	}
	if _, err := adminPool.Exec(ctx, "create schema "+quotedSchema); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		_, _ = adminPool.Exec(context.Background(), "drop schema if exists "+quotedSchema+" cascade")
		adminPool.Close()
	})

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse test postgres config: %v", err)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schemaName

	testPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("open test postgres pool: %v", err)
	}
	t.Cleanup(testPool.Close)
	if err := migrations.Apply(ctx, testPool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	return testPool
}

func pgxIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func assertCheckViolation(t *testing.T, err error) {
	t.Helper()
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23514" {
		t.Fatalf("error = %v, want check violation", err)
	}
}

func assertForeignKeyViolation(t *testing.T, err error) {
	t.Helper()
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		t.Fatalf("error = %v, want foreign key violation", err)
	}
}

func insertBaseItem(t *testing.T, pool *pgxpool.Pool, id, userID, title string) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		insert into items (
			id, user_id, item_type, title, status, created_at, updated_at,
			client_updated_at, server_updated_at, created_by_device_id,
			updated_by_device_id, revision
		) values (
			$1, $2, 'inbox', $3, 'active', now(), now(), now(), now(), $4, $4, 1
		)
	`, id, userID, title, device)
	if err != nil {
		t.Fatalf("insert base item %s: %v", id, err)
	}
}

func insertArea(t *testing.T, pool *pgxpool.Pool, id, userID string) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		insert into areas (
			id, user_id, name, sort_order, created_at, updated_at, client_updated_at,
			server_updated_at, created_by_device_id, updated_by_device_id, revision
		) values ($1, $2, 'Personal', 1, now(), now(), now(), now(), $3, $3, 1)
	`, id, userID, device)
	if err != nil {
		t.Fatalf("insert area %s: %v", id, err)
	}
}

func insertRecurringItem(t *testing.T, pool *pgxpool.Pool, id, userID, templateID string, sequence int) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		insert into items (
			id, user_id, item_type, title, status, scheduled_date,
			recurring_template_id, recurring_sequence, created_at, updated_at,
			client_updated_at, server_updated_at, created_by_device_id,
			updated_by_device_id, revision
		) values (
			$1, $2, 'date_task', $3, 'active', current_date,
			$4, $5, now(), now(), now(), now(), $6, $6, 1
		)
	`, id, userID, fmt.Sprintf("Generated %d", sequence), templateID, sequence, device)
	if err != nil {
		t.Fatalf("insert recurring item %s: %v", id, err)
	}
}

func insertTemplate(t *testing.T, pool *pgxpool.Pool, id, userID string) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		insert into recurring_task_templates (
			id, user_id, title, frequency, interval, recurrence_basis, start_date,
			end_type, completed_count, next_sequence, status, created_at, updated_at,
			client_updated_at, server_updated_at, created_by_device_id, updated_by_device_id, revision
		) values (
			$1, $2, 'Stretch', 'daily', 1, 'scheduled_date', current_date,
			'never', 0, 1, 'active', now(), now(), now(), now(), $3, $3, 1
		)
	`, id, userID, device)
	if err != nil {
		t.Fatalf("insert template %s: %v", id, err)
	}
}

func insertOperation(t *testing.T, pool *pgxpool.Pool, id, userID, idempotencyKey string) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		insert into operation_history (
			id, user_id, event_type, idempotency_key, created_at, created_by_device_id
		) values ($1, $2, 'completed', $3, now(), $4)
	`, id, userID, idempotencyKey, device)
	if err != nil {
		t.Fatalf("insert operation %s: %v", id, err)
	}
}
