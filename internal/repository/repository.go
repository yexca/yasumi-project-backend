package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("repository: not found")

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

type Tx struct {
	tx pgx.Tx
}

func (r *Repository) InTx(ctx context.Context, fn func(context.Context, *Tx) error) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := fn(ctx, &Tx{tx: tx}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

type ItemRecord struct {
	ID              string
	UserID          string
	ItemType        string
	Title           string
	Status          string
	RecurringID     *string
	RecurringSeq    *int
	UpdatedAt       time.Time
	ServerUpdatedAt time.Time
	Revision        int64
}

type OperationHistoryRecord struct {
	ID                string
	UserID            string
	ItemID            *string
	RecurringTemplate *string
	EventType         string
	PreviousValue     []byte
	NewValue          []byte
	Reason            *string
	IdempotencyKey    *string
	CreatedAt         time.Time
	CreatedByDeviceID string
}

type UserSettingsRecord struct {
	UserID                    string
	Language                  string
	Locale                    string
	WeekStartDay              string
	TimeZone                  string
	DateDisplayFormat         string
	TimeDisplayFormat         string
	DefaultTimeZoneMode       string
	TodayPrimaryLookaheadDays int
	DeadlineAwarenessDays     int
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	ClientUpdatedAt           time.Time
	ServerUpdatedAt           time.Time
	CreatedByDeviceID         string
	UpdatedByDeviceID         string
	Revision                  int64
}

func (tx *Tx) GetItemForUpdateByUser(ctx context.Context, userID, itemID string) (ItemRecord, error) {
	const query = `
		select id::text, user_id::text, item_type, title, status,
			recurring_template_id::text, recurring_sequence,
			updated_at, server_updated_at, revision
		from items
		where user_id = $1 and id = $2
		for update
	`
	var row ItemRecord
	if err := tx.tx.QueryRow(ctx, query, userID, itemID).Scan(
		&row.ID,
		&row.UserID,
		&row.ItemType,
		&row.Title,
		&row.Status,
		&row.RecurringID,
		&row.RecurringSeq,
		&row.UpdatedAt,
		&row.ServerUpdatedAt,
		&row.Revision,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ItemRecord{}, ErrNotFound
		}
		return ItemRecord{}, fmt.Errorf("get item for update by user: %w", err)
	}
	return row, nil
}

func (tx *Tx) UpdateItemServerMetadata(ctx context.Context, userID, itemID string, updatedAt time.Time, revision int64, deviceID string) error {
	const query = `
		update items
		set updated_at = $3,
			server_updated_at = $3,
			updated_by_device_id = $5,
			revision = $4
		where user_id = $1 and id = $2
	`
	tag, err := tx.tx.Exec(ctx, query, userID, itemID, updatedAt, revision, deviceID)
	if err != nil {
		return fmt.Errorf("update item server metadata: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (tx *Tx) InsertOperationHistory(ctx context.Context, row OperationHistoryRecord) error {
	const query = `
		insert into operation_history (
			id,
			user_id,
			item_id,
			recurring_template_id,
			event_type,
			previous_value,
			new_value,
			reason,
			idempotency_key,
			created_at,
			created_by_device_id
		) values ($1, $2, $3, $4, $5, coalesce($6, '{}'::jsonb), coalesce($7, '{}'::jsonb), $8, $9, $10, $11)
	`
	if len(row.PreviousValue) == 0 {
		row.PreviousValue = []byte("{}")
	}
	if len(row.NewValue) == 0 {
		row.NewValue = []byte("{}")
	}
	if _, err := tx.tx.Exec(ctx, query,
		row.ID,
		row.UserID,
		row.ItemID,
		row.RecurringTemplate,
		row.EventType,
		row.PreviousValue,
		row.NewValue,
		row.Reason,
		row.IdempotencyKey,
		row.CreatedAt,
		row.CreatedByDeviceID,
	); err != nil {
		return fmt.Errorf("insert operation history: %w", err)
	}
	return nil
}

func (tx *Tx) FindOperationByIdempotencyKey(ctx context.Context, userID, idempotencyKey string) (OperationHistoryRecord, error) {
	const query = `
		select id::text, user_id::text, item_id::text, recurring_template_id::text,
			event_type, previous_value, new_value, reason, idempotency_key,
			created_at, created_by_device_id
		from operation_history
		where user_id = $1 and idempotency_key = $2
	`
	var row OperationHistoryRecord
	if err := tx.tx.QueryRow(ctx, query, userID, idempotencyKey).Scan(
		&row.ID,
		&row.UserID,
		&row.ItemID,
		&row.RecurringTemplate,
		&row.EventType,
		&row.PreviousValue,
		&row.NewValue,
		&row.Reason,
		&row.IdempotencyKey,
		&row.CreatedAt,
		&row.CreatedByDeviceID,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OperationHistoryRecord{}, ErrNotFound
		}
		return OperationHistoryRecord{}, fmt.Errorf("find operation by idempotency key: %w", err)
	}
	return row, nil
}

func (tx *Tx) UpsertUserSettings(ctx context.Context, row UserSettingsRecord) error {
	const query = `
		insert into user_settings (
			user_id,
			language,
			locale,
			week_start_day,
			time_zone,
			date_display_format,
			time_display_format,
			default_time_zone_mode,
			today_primary_lookahead_days,
			deadline_awareness_days,
			created_at,
			updated_at,
			client_updated_at,
			server_updated_at,
			created_by_device_id,
			updated_by_device_id,
			revision
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		on conflict (user_id) do update set
			language = excluded.language,
			locale = excluded.locale,
			week_start_day = excluded.week_start_day,
			time_zone = excluded.time_zone,
			date_display_format = excluded.date_display_format,
			time_display_format = excluded.time_display_format,
			default_time_zone_mode = excluded.default_time_zone_mode,
			today_primary_lookahead_days = excluded.today_primary_lookahead_days,
			deadline_awareness_days = excluded.deadline_awareness_days,
			updated_at = excluded.updated_at,
			client_updated_at = excluded.client_updated_at,
			server_updated_at = excluded.server_updated_at,
			updated_by_device_id = excluded.updated_by_device_id,
			revision = excluded.revision
	`
	if _, err := tx.tx.Exec(ctx, query,
		row.UserID,
		row.Language,
		row.Locale,
		row.WeekStartDay,
		row.TimeZone,
		row.DateDisplayFormat,
		row.TimeDisplayFormat,
		row.DefaultTimeZoneMode,
		row.TodayPrimaryLookaheadDays,
		row.DeadlineAwarenessDays,
		row.CreatedAt,
		row.UpdatedAt,
		row.ClientUpdatedAt,
		row.ServerUpdatedAt,
		row.CreatedByDeviceID,
		row.UpdatedByDeviceID,
		row.Revision,
	); err != nil {
		return fmt.Errorf("upsert user settings: %w", err)
	}
	return nil
}
