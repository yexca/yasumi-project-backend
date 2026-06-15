package repository

import (
	"encoding/json"
	"time"
)

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
	WeatherCity               string    `json:"weather_city"`
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
