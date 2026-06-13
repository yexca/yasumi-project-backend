package repository

import (
	"context"
	"encoding/json"
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
	ID                    string          `json:"id"`
	UserID                string          `json:"user_id"`
	ItemType              string          `json:"item_type"`
	Title                 string          `json:"title"`
	Note                  *string         `json:"note"`
	Status                string          `json:"status"`
	AreaID                *string         `json:"area_id"`
	ScheduledDate         *string         `json:"scheduled_date"`
	ScheduledTime         *string         `json:"scheduled_time"`
	PlannedWorkDate       *string         `json:"planned_work_date"`
	DeadlineDate          *string         `json:"deadline_date"`
	DeadlineTime          *string         `json:"deadline_time"`
	DeadlineAt            *time.Time      `json:"deadline_at"`
	DeadlineTimezone      *string         `json:"deadline_timezone"`
	ReviewDate            *string         `json:"review_date"`
	ReminderDate          *string         `json:"reminder_date"`
	ReminderTime          *string         `json:"reminder_time"`
	ReminderAt            *time.Time      `json:"reminder_at"`
	ReminderIntent        *string         `json:"reminder_intent"`
	ScheduledTimeZoneMode *string         `json:"scheduled_time_zone_mode"`
	DeadlineTimeZoneMode  *string         `json:"deadline_time_zone_mode"`
	ReminderTimeZoneMode  *string         `json:"reminder_time_zone_mode"`
	RecurringID           *string         `json:"recurring_template_id"`
	RecurringSeq          *int            `json:"recurring_sequence"`
	RecurringAnchorDate   *string         `json:"recurring_anchor_date"`
	GeneratedFromItemID   *string         `json:"generated_from_item_id"`
	Importance            *int            `json:"importance"`
	EstimatedEffort       *int            `json:"estimated_effort"`
	PressureMetadata      json.RawMessage `json:"pressure_metadata"`
	QuickAddSourceText    *string         `json:"quick_add_source_text"`
	QuickAddParseResult   json.RawMessage `json:"quick_add_parse_result"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	DeletedAt             *time.Time      `json:"deleted_at"`
	ArchivedAt            *time.Time      `json:"archived_at"`
	HiddenReason          *string         `json:"hidden_reason"`
	ClientUpdatedAt       time.Time       `json:"client_updated_at"`
	ServerUpdatedAt       time.Time       `json:"server_updated_at"`
	CreatedByDeviceID     string          `json:"created_by_device_id"`
	UpdatedByDeviceID     string          `json:"updated_by_device_id"`
	Revision              int64           `json:"revision"`
}

type AreaRecord struct {
	ID                string     `json:"id"`
	UserID            string     `json:"user_id"`
	Name              string     `json:"name"`
	SortOrder         int        `json:"sort_order"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at"`
	ArchivedAt        *time.Time `json:"archived_at"`
	HiddenReason      *string    `json:"hidden_reason"`
	ClientUpdatedAt   time.Time  `json:"client_updated_at"`
	ServerUpdatedAt   time.Time  `json:"server_updated_at"`
	CreatedByDeviceID string     `json:"created_by_device_id"`
	UpdatedByDeviceID string     `json:"updated_by_device_id"`
	Revision          int64      `json:"revision"`
}

type RecurringTemplateRecord struct {
	ID                    string          `json:"id"`
	UserID                string          `json:"user_id"`
	Title                 string          `json:"title"`
	Note                  *string         `json:"note"`
	AreaID                *string         `json:"area_id"`
	Frequency             string          `json:"frequency"`
	Interval              int             `json:"interval"`
	Weekdays              json.RawMessage `json:"weekdays"`
	RecurrenceBasis       string          `json:"recurrence_basis"`
	StartDate             string          `json:"start_date"`
	EndType               string          `json:"end_type"`
	EndDate               *string         `json:"end_date"`
	EndAfterCount         *int            `json:"end_after_count"`
	CompletedCount        int             `json:"completed_count"`
	NextSequence          int             `json:"next_sequence"`
	ScheduledTime         *string         `json:"scheduled_time"`
	ReminderRule          json.RawMessage `json:"reminder_rule"`
	GeneratedTaskDefaults json.RawMessage `json:"generated_task_defaults"`
	Status                string          `json:"status"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	DeletedAt             *time.Time      `json:"deleted_at"`
	ArchivedAt            *time.Time      `json:"archived_at"`
	HiddenReason          *string         `json:"hidden_reason"`
	ClientUpdatedAt       time.Time       `json:"client_updated_at"`
	ServerUpdatedAt       time.Time       `json:"server_updated_at"`
	CreatedByDeviceID     string          `json:"created_by_device_id"`
	UpdatedByDeviceID     string          `json:"updated_by_device_id"`
	Revision              int64           `json:"revision"`
}

type OperationHistoryRecord struct {
	ID                string          `json:"id"`
	UserID            string          `json:"user_id"`
	ItemID            *string         `json:"item_id"`
	RecurringTemplate *string         `json:"recurring_template_id"`
	EventType         string          `json:"event_type"`
	PreviousValue     json.RawMessage `json:"previous_value"`
	NewValue          json.RawMessage `json:"new_value"`
	Reason            *string         `json:"reason"`
	IdempotencyKey    *string         `json:"idempotency_key"`
	CreatedAt         time.Time       `json:"created_at"`
	CreatedByDeviceID string          `json:"created_by_device_id"`
}

type UserSettingsRecord struct {
	UserID                    string    `json:"user_id"`
	Language                  string    `json:"language"`
	Locale                    string    `json:"locale"`
	WeekStartDay              string    `json:"week_start_day"`
	TimeZone                  string    `json:"time_zone"`
	DateDisplayFormat         string    `json:"date_display_format"`
	TimeDisplayFormat         string    `json:"time_display_format"`
	DefaultTimeZoneMode       string    `json:"default_time_zone_mode"`
	TodayPrimaryLookaheadDays int       `json:"today_primary_lookahead_days"`
	DeadlineAwarenessDays     int       `json:"deadline_awareness_days"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
	ClientUpdatedAt           time.Time `json:"client_updated_at"`
	ServerUpdatedAt           time.Time `json:"server_updated_at"`
	CreatedByDeviceID         string    `json:"created_by_device_id"`
	UpdatedByDeviceID         string    `json:"updated_by_device_id"`
	Revision                  int64     `json:"revision"`
}

type UserRecord struct {
	ID              string     `json:"id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	DisplayName     *string    `json:"display_name"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CredentialRecord struct {
	UserID                string          `json:"user_id"`
	PasswordHash          string          `json:"password_hash"`
	PasswordHashAlgorithm string          `json:"password_hash_algorithm"`
	PasswordHashParams    json.RawMessage `json:"password_hash_params"`
	PasswordChangedAt     time.Time       `json:"password_changed_at"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

type AccountWithCredentialRecord struct {
	User       UserRecord
	Credential CredentialRecord
}

type SessionRecord struct {
	ID                  string     `json:"id"`
	UserID              string     `json:"user_id"`
	RefreshTokenHash    string     `json:"refresh_token_hash"`
	CreatedAt           time.Time  `json:"created_at"`
	LastUsedAt          *time.Time `json:"last_used_at"`
	ExpiresAt           time.Time  `json:"expires_at"`
	RevokedAt           *time.Time `json:"revoked_at"`
	ReplacedBySessionID *string    `json:"replaced_by_session_id"`
	UserAgent           *string    `json:"user_agent"`
	IPHash              *string    `json:"ip_hash"`
}

type SessionWithUserRecord struct {
	Session SessionRecord
	User    UserRecord
}

func (tx *Tx) CreateAccount(ctx context.Context, user UserRecord, credential CredentialRecord, settings UserSettingsRecord) error {
	const userQuery = `
		insert into users (
			id, username, email, email_verified_at, display_name, status, created_at, updated_at
		) values ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if _, err := tx.tx.Exec(ctx, userQuery,
		user.ID,
		user.Username,
		user.Email,
		user.EmailVerifiedAt,
		user.DisplayName,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	const credentialQuery = `
		insert into user_credentials (
			user_id, password_hash, password_hash_algorithm, password_hash_params,
			password_changed_at, created_at, updated_at
		) values ($1, $2, $3, coalesce($4, '{}'::jsonb), $5, $6, $7)
	`
	if len(credential.PasswordHashParams) == 0 {
		credential.PasswordHashParams = []byte("{}")
	}
	if _, err := tx.tx.Exec(ctx, credentialQuery,
		credential.UserID,
		credential.PasswordHash,
		credential.PasswordHashAlgorithm,
		credential.PasswordHashParams,
		credential.PasswordChangedAt,
		credential.CreatedAt,
		credential.UpdatedAt,
	); err != nil {
		return fmt.Errorf("create user credential: %w", err)
	}

	if err := tx.UpsertUserSettings(ctx, settings); err != nil {
		return err
	}
	return nil
}

func (tx *Tx) FindAccountByIdentifier(ctx context.Context, identifier string) (AccountWithCredentialRecord, error) {
	const query = `
		select u.id::text, u.username, u.email, u.email_verified_at, u.display_name,
			u.status, u.created_at, u.updated_at,
			c.user_id::text, c.password_hash, c.password_hash_algorithm,
			c.password_hash_params, c.password_changed_at, c.created_at, c.updated_at
		from users u
		join user_credentials c on c.user_id = u.id
		where lower(u.username) = lower($1) or lower(u.email) = lower($1)
	`
	var row AccountWithCredentialRecord
	if err := tx.tx.QueryRow(ctx, query, identifier).Scan(
		&row.User.ID,
		&row.User.Username,
		&row.User.Email,
		&row.User.EmailVerifiedAt,
		&row.User.DisplayName,
		&row.User.Status,
		&row.User.CreatedAt,
		&row.User.UpdatedAt,
		&row.Credential.UserID,
		&row.Credential.PasswordHash,
		&row.Credential.PasswordHashAlgorithm,
		&row.Credential.PasswordHashParams,
		&row.Credential.PasswordChangedAt,
		&row.Credential.CreatedAt,
		&row.Credential.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AccountWithCredentialRecord{}, ErrNotFound
		}
		return AccountWithCredentialRecord{}, fmt.Errorf("find account by identifier: %w", err)
	}
	return row, nil
}

func (tx *Tx) CreateSession(ctx context.Context, session SessionRecord) error {
	const query = `
		insert into user_sessions (
			id, user_id, refresh_token_hash, created_at, last_used_at, expires_at,
			revoked_at, replaced_by_session_id, user_agent, ip_hash
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	if _, err := tx.tx.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.RefreshTokenHash,
		session.CreatedAt,
		session.LastUsedAt,
		session.ExpiresAt,
		session.RevokedAt,
		session.ReplacedBySessionID,
		session.UserAgent,
		session.IPHash,
	); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (tx *Tx) GetActiveSessionWithUserForUpdate(ctx context.Context, sessionID string, now time.Time) (SessionWithUserRecord, error) {
	const query = `
		select s.id::text, s.user_id::text, s.refresh_token_hash, s.created_at,
			s.last_used_at, s.expires_at, s.revoked_at, s.replaced_by_session_id::text,
			s.user_agent, s.ip_hash,
			u.id::text, u.username, u.email, u.email_verified_at, u.display_name,
			u.status, u.created_at, u.updated_at
		from user_sessions s
		join users u on u.id = s.user_id
		where s.id = $1 and s.revoked_at is null and s.expires_at > $2
		for update of s
	`
	var row SessionWithUserRecord
	if err := tx.tx.QueryRow(ctx, query, sessionID, now).Scan(
		&row.Session.ID,
		&row.Session.UserID,
		&row.Session.RefreshTokenHash,
		&row.Session.CreatedAt,
		&row.Session.LastUsedAt,
		&row.Session.ExpiresAt,
		&row.Session.RevokedAt,
		&row.Session.ReplacedBySessionID,
		&row.Session.UserAgent,
		&row.Session.IPHash,
		&row.User.ID,
		&row.User.Username,
		&row.User.Email,
		&row.User.EmailVerifiedAt,
		&row.User.DisplayName,
		&row.User.Status,
		&row.User.CreatedAt,
		&row.User.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionWithUserRecord{}, ErrNotFound
		}
		return SessionWithUserRecord{}, fmt.Errorf("get active session with user: %w", err)
	}
	return row, nil
}

func (tx *Tx) FindActiveSessionByRefreshHashForUpdate(ctx context.Context, refreshTokenHash string, now time.Time) (SessionWithUserRecord, error) {
	const query = `
		select s.id::text, s.user_id::text, s.refresh_token_hash, s.created_at,
			s.last_used_at, s.expires_at, s.revoked_at, s.replaced_by_session_id::text,
			s.user_agent, s.ip_hash,
			u.id::text, u.username, u.email, u.email_verified_at, u.display_name,
			u.status, u.created_at, u.updated_at
		from user_sessions s
		join users u on u.id = s.user_id
		where s.refresh_token_hash = $1 and s.revoked_at is null and s.expires_at > $2
		for update of s
	`
	var row SessionWithUserRecord
	if err := tx.tx.QueryRow(ctx, query, refreshTokenHash, now).Scan(
		&row.Session.ID,
		&row.Session.UserID,
		&row.Session.RefreshTokenHash,
		&row.Session.CreatedAt,
		&row.Session.LastUsedAt,
		&row.Session.ExpiresAt,
		&row.Session.RevokedAt,
		&row.Session.ReplacedBySessionID,
		&row.Session.UserAgent,
		&row.Session.IPHash,
		&row.User.ID,
		&row.User.Username,
		&row.User.Email,
		&row.User.EmailVerifiedAt,
		&row.User.DisplayName,
		&row.User.Status,
		&row.User.CreatedAt,
		&row.User.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionWithUserRecord{}, ErrNotFound
		}
		return SessionWithUserRecord{}, fmt.Errorf("find active session by refresh hash: %w", err)
	}
	return row, nil
}

func (tx *Tx) RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time, replacedBySessionID *string) error {
	const query = `
		update user_sessions
		set revoked_at = $2,
			replaced_by_session_id = $3,
			last_used_at = $2
		where id = $1 and revoked_at is null
	`
	tag, err := tx.tx.Exec(ctx, query, sessionID, revokedAt, replacedBySessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

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

func (tx *Tx) GetUserSettingsForUpdate(ctx context.Context, userID string) (UserSettingsRecord, error) {
	const query = `
		select user_id::text, language, locale, week_start_day, time_zone,
			date_display_format, time_display_format, default_time_zone_mode,
			today_primary_lookahead_days, deadline_awareness_days, created_at, updated_at,
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
