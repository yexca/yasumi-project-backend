package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (tx *Tx) GetAreaForUpdateByUser(ctx context.Context, userID, areaID string) (AreaRecord, error) {
	const query = `
		select id::text, user_id::text, name, sort_order, created_at, updated_at,
			deleted_at, archived_at, hidden_reason, client_updated_at, server_updated_at,
			created_by_device_id, updated_by_device_id, revision
		from areas
		where user_id = $1 and id = $2
		for update
	`
	var row AreaRecord
	if err := tx.tx.QueryRow(ctx, query, userID, areaID).Scan(
		&row.ID,
		&row.UserID,
		&row.Name,
		&row.SortOrder,
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
			return AreaRecord{}, ErrNotFound
		}
		return AreaRecord{}, fmt.Errorf("get area for update by user: %w", err)
	}
	return row, nil
}

func (tx *Tx) UpsertArea(ctx context.Context, row AreaRecord) error {
	const query = `
		insert into areas (
			id, user_id, name, sort_order, created_at, updated_at, deleted_at,
			archived_at, hidden_reason, client_updated_at, server_updated_at,
			created_by_device_id, updated_by_device_id, revision
		) values (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14
		)
		on conflict (id) do update set
			name = excluded.name,
			sort_order = excluded.sort_order,
			updated_at = excluded.updated_at,
			deleted_at = excluded.deleted_at,
			archived_at = excluded.archived_at,
			hidden_reason = excluded.hidden_reason,
			client_updated_at = excluded.client_updated_at,
			server_updated_at = excluded.server_updated_at,
			updated_by_device_id = excluded.updated_by_device_id,
			revision = excluded.revision
		where areas.user_id = excluded.user_id
	`
	tag, err := tx.tx.Exec(ctx, query,
		row.ID,
		row.UserID,
		row.Name,
		row.SortOrder,
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
		return fmt.Errorf("upsert area: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (tx *Tx) GetRecurringTemplateForUpdateByUser(ctx context.Context, userID, templateID string) (RecurringTemplateRecord, error) {
	const query = `
		select id::text, user_id::text, title, note, area_id::text, frequency, interval,
			weekdays, recurrence_basis, start_date::text, end_type, end_date::text,
			end_after_count, completed_count, next_sequence, scheduled_time::text,
			reminder_rule, generated_task_defaults, status, created_at, updated_at,
			deleted_at, archived_at, hidden_reason, client_updated_at, server_updated_at,
			created_by_device_id, updated_by_device_id, revision
		from recurring_task_templates
		where user_id = $1 and id = $2
		for update
	`
	var row RecurringTemplateRecord
	if err := tx.tx.QueryRow(ctx, query, userID, templateID).Scan(
		&row.ID,
		&row.UserID,
		&row.Title,
		&row.Note,
		&row.AreaID,
		&row.Frequency,
		&row.Interval,
		&row.Weekdays,
		&row.RecurrenceBasis,
		&row.StartDate,
		&row.EndType,
		&row.EndDate,
		&row.EndAfterCount,
		&row.CompletedCount,
		&row.NextSequence,
		&row.ScheduledTime,
		&row.ReminderRule,
		&row.GeneratedTaskDefaults,
		&row.Status,
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
			return RecurringTemplateRecord{}, ErrNotFound
		}
		return RecurringTemplateRecord{}, fmt.Errorf("get recurring template for update by user: %w", err)
	}
	return row, nil
}

func (tx *Tx) UpsertRecurringTemplate(ctx context.Context, row RecurringTemplateRecord) error {
	const query = `
		insert into recurring_task_templates (
			id, user_id, title, note, area_id, frequency, interval, weekdays,
			recurrence_basis, start_date, end_type, end_date, end_after_count,
			completed_count, next_sequence, scheduled_time, reminder_rule,
			generated_task_defaults, status, created_at, updated_at, deleted_at,
			archived_at, hidden_reason, client_updated_at, server_updated_at,
			created_by_device_id, updated_by_device_id, revision
		) values (
			$1, $2, $3, $4, $5, $6, $7, coalesce($8, '[]'::jsonb),
			$9, $10::date, $11, $12::date, $13,
			$14, $15, $16::time, coalesce($17, '{}'::jsonb),
			coalesce($18, '{}'::jsonb), $19, $20, $21, $22,
			$23, $24, $25, $26,
			$27, $28, $29
		)
		on conflict (id) do update set
			title = excluded.title,
			note = excluded.note,
			area_id = excluded.area_id,
			frequency = excluded.frequency,
			interval = excluded.interval,
			weekdays = excluded.weekdays,
			recurrence_basis = excluded.recurrence_basis,
			start_date = excluded.start_date,
			end_type = excluded.end_type,
			end_date = excluded.end_date,
			end_after_count = excluded.end_after_count,
			completed_count = excluded.completed_count,
			next_sequence = excluded.next_sequence,
			scheduled_time = excluded.scheduled_time,
			reminder_rule = excluded.reminder_rule,
			generated_task_defaults = excluded.generated_task_defaults,
			status = excluded.status,
			updated_at = excluded.updated_at,
			deleted_at = excluded.deleted_at,
			archived_at = excluded.archived_at,
			hidden_reason = excluded.hidden_reason,
			client_updated_at = excluded.client_updated_at,
			server_updated_at = excluded.server_updated_at,
			updated_by_device_id = excluded.updated_by_device_id,
			revision = excluded.revision
		where recurring_task_templates.user_id = excluded.user_id
	`
	if len(row.Weekdays) == 0 {
		row.Weekdays = []byte("[]")
	}
	if len(row.ReminderRule) == 0 {
		row.ReminderRule = []byte("{}")
	}
	if len(row.GeneratedTaskDefaults) == 0 {
		row.GeneratedTaskDefaults = []byte("{}")
	}
	tag, err := tx.tx.Exec(ctx, query,
		row.ID,
		row.UserID,
		row.Title,
		row.Note,
		row.AreaID,
		row.Frequency,
		row.Interval,
		row.Weekdays,
		row.RecurrenceBasis,
		row.StartDate,
		row.EndType,
		row.EndDate,
		row.EndAfterCount,
		row.CompletedCount,
		row.NextSequence,
		row.ScheduledTime,
		row.ReminderRule,
		row.GeneratedTaskDefaults,
		row.Status,
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
		return fmt.Errorf("upsert recurring template: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (tx *Tx) GetUserSettingsForUpdate(ctx context.Context, userID string) (UserSettingsRecord, error) {
	const query = `
		select user_id::text, language, locale, week_start_day, time_zone,
			date_display_format, time_display_format, default_time_zone_mode,
			today_primary_lookahead_days, deadline_awareness_days, weather_city, created_at, updated_at,
			client_updated_at, server_updated_at, created_by_device_id,
			updated_by_device_id, revision
		from user_settings
		where user_id = $1
		for update
	`
	var row UserSettingsRecord
	if err := tx.tx.QueryRow(ctx, query, userID).Scan(
		&row.UserID,
		&row.Language,
		&row.Locale,
		&row.WeekStartDay,
		&row.TimeZone,
		&row.DateDisplayFormat,
		&row.TimeDisplayFormat,
		&row.DefaultTimeZoneMode,
		&row.TodayPrimaryLookaheadDays,
		&row.DeadlineAwarenessDays,
		&row.WeatherCity,
		&row.CreatedAt,
		&row.UpdatedAt,
		&row.ClientUpdatedAt,
		&row.ServerUpdatedAt,
		&row.CreatedByDeviceID,
		&row.UpdatedByDeviceID,
		&row.Revision,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserSettingsRecord{}, ErrNotFound
		}
		return UserSettingsRecord{}, fmt.Errorf("get user settings for update: %w", err)
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
			weather_city,
			created_at,
			updated_at,
			client_updated_at,
			server_updated_at,
			created_by_device_id,
			updated_by_device_id,
			revision
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
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
			weather_city = excluded.weather_city,
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
		row.WeatherCity,
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
