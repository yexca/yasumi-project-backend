package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (tx *Tx) GetItemForUpdateByUser(ctx context.Context, userID, itemID string) (ItemRecord, error) {
	const query = `
		select id::text, user_id::text, item_type, title, note, status,
			area_id::text, scheduled_date::text, scheduled_time::text, planned_work_date::text,
			deadline_date::text, deadline_time::text, deadline_at, deadline_timezone,
			review_date::text, reminder_date::text, reminder_time::text, reminder_at,
			reminder_intent, scheduled_time_zone_mode, deadline_time_zone_mode,
			reminder_time_zone_mode, recurring_template_id::text, recurring_sequence,
			recurring_anchor_date::text, generated_from_item_id::text, importance,
			estimated_effort, pressure_metadata, quick_add_source_text, quick_add_parse_result,
			created_at, updated_at, deleted_at, archived_at, hidden_reason,
			client_updated_at, server_updated_at, created_by_device_id,
			updated_by_device_id, revision
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
		&row.Note,
		&row.Status,
		&row.AreaID,
		&row.ScheduledDate,
		&row.ScheduledTime,
		&row.PlannedWorkDate,
		&row.DeadlineDate,
		&row.DeadlineTime,
		&row.DeadlineAt,
		&row.DeadlineTimezone,
		&row.ReviewDate,
		&row.ReminderDate,
		&row.ReminderTime,
		&row.ReminderAt,
		&row.ReminderIntent,
		&row.ScheduledTimeZoneMode,
		&row.DeadlineTimeZoneMode,
		&row.ReminderTimeZoneMode,
		&row.RecurringID,
		&row.RecurringSeq,
		&row.RecurringAnchorDate,
		&row.GeneratedFromItemID,
		&row.Importance,
		&row.EstimatedEffort,
		&row.PressureMetadata,
		&row.QuickAddSourceText,
		&row.QuickAddParseResult,
		&row.CreatedAt,
		&row.UpdatedAt,
		&row.DeletedAt,
		&row.ArchivedAt,
		&row.HiddenReason,
		&row.ClientUpdatedAt,
		&row.ServerUpdatedAt,
		&row.CreatedByDeviceID,
		&row.UpdatedByDeviceID,
		&row.Revision,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ItemRecord{}, ErrNotFound
		}
		return ItemRecord{}, fmt.Errorf("get item for update by user: %w", err)
	}
	return row, nil
}

func (tx *Tx) UpsertItem(ctx context.Context, row ItemRecord) error {
	const query = `
		insert into items (
			id, user_id, item_type, title, note, status, area_id,
			scheduled_date, scheduled_time, planned_work_date, deadline_date,
			deadline_time, deadline_at, deadline_timezone, review_date,
			reminder_date, reminder_time, reminder_at, reminder_intent,
			scheduled_time_zone_mode, deadline_time_zone_mode, reminder_time_zone_mode,
			recurring_template_id, recurring_sequence, recurring_anchor_date,
			generated_from_item_id, importance, estimated_effort, pressure_metadata,
			quick_add_source_text, quick_add_parse_result, created_at, updated_at,
			deleted_at, archived_at, hidden_reason, client_updated_at, server_updated_at,
			created_by_device_id, updated_by_device_id, revision
		) values (
			$1, $2, $3, $4, $5, $6, $7,
			$8::date, $9::time, $10::date, $11::date,
			$12::time, $13, $14, $15::date,
			$16::date, $17::time, $18, $19,
			$20, $21, $22,
			$23, $24, $25::date,
			$26, $27, $28, coalesce($29, '{}'::jsonb),
			$30, $31, $32, $33,
			$34, $35, $36, $37, $38,
			$39, $40, $41
		)
		on conflict (id) do update set
			item_type = excluded.item_type,
			title = excluded.title,
			note = excluded.note,
			status = excluded.status,
			area_id = excluded.area_id,
			scheduled_date = excluded.scheduled_date,
			scheduled_time = excluded.scheduled_time,
			planned_work_date = excluded.planned_work_date,
			deadline_date = excluded.deadline_date,
			deadline_time = excluded.deadline_time,
			deadline_at = excluded.deadline_at,
			deadline_timezone = excluded.deadline_timezone,
			review_date = excluded.review_date,
			reminder_date = excluded.reminder_date,
			reminder_time = excluded.reminder_time,
			reminder_at = excluded.reminder_at,
			reminder_intent = excluded.reminder_intent,
			scheduled_time_zone_mode = excluded.scheduled_time_zone_mode,
			deadline_time_zone_mode = excluded.deadline_time_zone_mode,
			reminder_time_zone_mode = excluded.reminder_time_zone_mode,
			recurring_template_id = excluded.recurring_template_id,
			recurring_sequence = excluded.recurring_sequence,
			recurring_anchor_date = excluded.recurring_anchor_date,
			generated_from_item_id = excluded.generated_from_item_id,
			importance = excluded.importance,
			estimated_effort = excluded.estimated_effort,
			pressure_metadata = excluded.pressure_metadata,
			quick_add_source_text = excluded.quick_add_source_text,
			quick_add_parse_result = excluded.quick_add_parse_result,
			updated_at = excluded.updated_at,
			deleted_at = excluded.deleted_at,
			archived_at = excluded.archived_at,
			hidden_reason = excluded.hidden_reason,
			client_updated_at = excluded.client_updated_at,
			server_updated_at = excluded.server_updated_at,
			updated_by_device_id = excluded.updated_by_device_id,
			revision = excluded.revision
		where items.user_id = excluded.user_id
	`
	if len(row.PressureMetadata) == 0 {
		row.PressureMetadata = []byte("{}")
	}
	tag, err := tx.tx.Exec(ctx, query,
		row.ID,
		row.UserID,
		row.ItemType,
		row.Title,
		row.Note,
		row.Status,
		row.AreaID,
		row.ScheduledDate,
		row.ScheduledTime,
		row.PlannedWorkDate,
		row.DeadlineDate,
		row.DeadlineTime,
		row.DeadlineAt,
		row.DeadlineTimezone,
		row.ReviewDate,
		row.ReminderDate,
		row.ReminderTime,
		row.ReminderAt,
		row.ReminderIntent,
		row.ScheduledTimeZoneMode,
		row.DeadlineTimeZoneMode,
		row.ReminderTimeZoneMode,
		row.RecurringID,
		row.RecurringSeq,
		row.RecurringAnchorDate,
		row.GeneratedFromItemID,
		row.Importance,
		row.EstimatedEffort,
		row.PressureMetadata,
		row.QuickAddSourceText,
		row.QuickAddParseResult,
		row.CreatedAt,
		row.UpdatedAt,
		row.DeletedAt,
		row.ArchivedAt,
		row.HiddenReason,
		row.ClientUpdatedAt,
		row.ServerUpdatedAt,
		row.CreatedByDeviceID,
		row.UpdatedByDeviceID,
		row.Revision,
	)
	if err != nil {
		return fmt.Errorf("upsert item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
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
