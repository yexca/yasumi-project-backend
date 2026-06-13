package domain

type ItemType string

const (
	ItemTypeInbox        ItemType = "inbox"
	ItemTypeDateTask     ItemType = "date_task"
	ItemTypeDeadlineTask ItemType = "deadline_task"
	ItemTypeIdea         ItemType = "idea"
)

type BusinessStatus string

const (
	StatusActive    BusinessStatus = "active"
	StatusCompleted BusinessStatus = "completed"
	StatusPostponed BusinessStatus = "postponed"
	StatusOnHold    BusinessStatus = "on_hold"
	StatusAbandoned BusinessStatus = "abandoned"
)

type HiddenReason string

const (
	HiddenReasonConvertedToRecurringTemplate HiddenReason = "converted_to_recurring_template"
	HiddenReasonRecurringSkipped             HiddenReason = "recurring_skipped"
)

type ScheduledTimeZoneMode string

const (
	ScheduledTimeZoneModeFloating ScheduledTimeZoneMode = "floating"
)

type DeadlineTimeZoneMode string

const (
	DeadlineTimeZoneModeDateOnly DeadlineTimeZoneMode = "date_only"
	DeadlineTimeZoneModeFloating DeadlineTimeZoneMode = "floating"
	DeadlineTimeZoneModeFixed    DeadlineTimeZoneMode = "fixed"
)

type ReminderTimeZoneMode string

const (
	ReminderTimeZoneModeFloating ReminderTimeZoneMode = "floating"
	ReminderTimeZoneModeFixed    ReminderTimeZoneMode = "fixed"
)

type RecurrenceFrequency string

const (
	RecurrenceFrequencyDaily   RecurrenceFrequency = "daily"
	RecurrenceFrequencyWeekly  RecurrenceFrequency = "weekly"
	RecurrenceFrequencyMonthly RecurrenceFrequency = "monthly"
	RecurrenceFrequencyYearly  RecurrenceFrequency = "yearly"
)

type RecurrenceBasis string

const (
	RecurrenceBasisCompletionDate RecurrenceBasis = "completion_date"
	RecurrenceBasisScheduledDate  RecurrenceBasis = "scheduled_date"
	RecurrenceBasisDeadlineDate   RecurrenceBasis = "deadline_date"
)

type RecurrenceEndType string

const (
	RecurrenceEndTypeNever      RecurrenceEndType = "never"
	RecurrenceEndTypeAfterCount RecurrenceEndType = "after_count"
	RecurrenceEndTypeOnDate     RecurrenceEndType = "on_date"
)

type RecurringTemplateStatus string

const (
	RecurringTemplateStatusActive    RecurringTemplateStatus = "active"
	RecurringTemplateStatusOnHold    RecurringTemplateStatus = "on_hold"
	RecurringTemplateStatusAbandoned RecurringTemplateStatus = "abandoned"
)

type WeekStartDay string

const (
	WeekStartSunday WeekStartDay = "sunday"
	WeekStartMonday WeekStartDay = "monday"
)

type LanguageCode string

const (
	LanguageEnglish           LanguageCode = "en"
	LanguageSimplifiedChinese LanguageCode = "zh-Hans"
	LanguageJapanese          LanguageCode = "ja"
)

type TimeDisplayFormat string

const (
	TimeDisplay12Hour TimeDisplayFormat = "12h"
	TimeDisplay24Hour TimeDisplayFormat = "24h"
)

type DefaultTimeZoneMode string

const (
	DefaultTimeZoneModeFloating DefaultTimeZoneMode = "floating"
)

type OperationEventType string

const (
	OperationCompleted                    OperationEventType = "completed"
	OperationPostponed                    OperationEventType = "postponed"
	OperationActivatedFromPostponed       OperationEventType = "activated_from_postponed"
	OperationOnHold                       OperationEventType = "on_hold"
	OperationAbandoned                    OperationEventType = "abandoned"
	OperationRestored                     OperationEventType = "restored"
	OperationReopened                     OperationEventType = "reopened"
	OperationDeleted                      OperationEventType = "deleted"
	OperationArchived                     OperationEventType = "archived"
	OperationSkipped                      OperationEventType = "skipped"
	OperationGeneratedNextInstance        OperationEventType = "generated_next_instance"
	OperationConvertedToRecurringTemplate OperationEventType = "converted_to_recurring_template"
)

type TodayReasonKey string

const (
	TodayReasonScheduledToday        TodayReasonKey = "scheduledToday"
	TodayReasonRecurringToday        TodayReasonKey = "recurringToday"
	TodayReasonCarriedForward        TodayReasonKey = "carriedForward"
	TodayReasonPlannedToday          TodayReasonKey = "plannedToday"
	TodayReasonDeadlineToday         TodayReasonKey = "deadlineToday"
	TodayReasonDeadlineTomorrow      TodayReasonKey = "deadlineTomorrow"
	TodayReasonDeadlineSoon          TodayReasonKey = "deadlineSoon"
	TodayReasonSmallTask             TodayReasonKey = "smallTask"
	TodayReasonHighImportance        TodayReasonKey = "highImportance"
	TodayReasonPostponedSeveralTimes TodayReasonKey = "postponedSeveralTimes"
	TodayReasonReviewDateReached     TodayReasonKey = "reviewDateReached"
)

type ErrorCode string

const (
	ErrorUnauthorized               ErrorCode = "unauthorized"
	ErrorForbidden                  ErrorCode = "forbidden"
	ErrorNotFound                   ErrorCode = "not_found"
	ErrorValidationFailed           ErrorCode = "validation_failed"
	ErrorInvalidTransition          ErrorCode = "invalid_transition"
	ErrorUnsupportedOperation       ErrorCode = "unsupported_operation"
	ErrorDuplicateRecurringInstance ErrorCode = "duplicate_recurring_instance"
	ErrorServiceUnavailable         ErrorCode = "service_unavailable"
)

type FieldKey string

const (
	FieldTitle                     FieldKey = "title"
	FieldItemType                  FieldKey = "item_type"
	FieldStatus                    FieldKey = "status"
	FieldScheduledDate             FieldKey = "scheduled_date"
	FieldDeadlineDate              FieldKey = "deadline_date"
	FieldDeadlineTime              FieldKey = "deadline_time"
	FieldDeadlineAt                FieldKey = "deadline_at"
	FieldDeadlineTimeZoneMode      FieldKey = "deadline_time_zone_mode"
	FieldUserID                    FieldKey = "user_id"
	FieldIdempotencyKey            FieldKey = "idempotency_key"
	FieldRecurringSequence         FieldKey = "recurring_sequence"
	FieldLanguage                  FieldKey = "language"
	FieldLocale                    FieldKey = "locale"
	FieldWeekStartDay              FieldKey = "week_start_day"
	FieldTimeZone                  FieldKey = "time_zone"
	FieldDateDisplayFormat         FieldKey = "date_display_format"
	FieldTimeDisplayFormat         FieldKey = "time_display_format"
	FieldDefaultTimeZoneMode       FieldKey = "default_time_zone_mode"
	FieldTodayPrimaryLookaheadDays FieldKey = "today_primary_lookahead_days"
	FieldDeadlineAwarenessDays     FieldKey = "deadline_awareness_days"
)

func IsValidItemType(v ItemType) bool {
	return isOneOf(v, ItemTypeInbox, ItemTypeDateTask, ItemTypeDeadlineTask, ItemTypeIdea)
}

func IsValidBusinessStatus(v BusinessStatus) bool {
	return isOneOf(v, StatusActive, StatusCompleted, StatusPostponed, StatusOnHold, StatusAbandoned)
}

func IsValidHiddenReason(v HiddenReason) bool {
	return isOneOf(v, HiddenReasonConvertedToRecurringTemplate, HiddenReasonRecurringSkipped)
}

func IsValidScheduledTimeZoneMode(v ScheduledTimeZoneMode) bool {
	return isOneOf(v, ScheduledTimeZoneModeFloating)
}

func IsValidDeadlineTimeZoneMode(v DeadlineTimeZoneMode) bool {
	return isOneOf(v, DeadlineTimeZoneModeDateOnly, DeadlineTimeZoneModeFloating, DeadlineTimeZoneModeFixed)
}

func IsValidReminderTimeZoneMode(v ReminderTimeZoneMode) bool {
	return isOneOf(v, ReminderTimeZoneModeFloating, ReminderTimeZoneModeFixed)
}

func IsValidRecurrenceFrequency(v RecurrenceFrequency) bool {
	return isOneOf(v, RecurrenceFrequencyDaily, RecurrenceFrequencyWeekly, RecurrenceFrequencyMonthly, RecurrenceFrequencyYearly)
}

func IsValidRecurrenceBasis(v RecurrenceBasis) bool {
	return isOneOf(v, RecurrenceBasisCompletionDate, RecurrenceBasisScheduledDate, RecurrenceBasisDeadlineDate)
}

func IsValidRecurrenceEndType(v RecurrenceEndType) bool {
	return isOneOf(v, RecurrenceEndTypeNever, RecurrenceEndTypeAfterCount, RecurrenceEndTypeOnDate)
}

func IsValidRecurringTemplateStatus(v RecurringTemplateStatus) bool {
	return isOneOf(v, RecurringTemplateStatusActive, RecurringTemplateStatusOnHold, RecurringTemplateStatusAbandoned)
}

func IsValidWeekStartDay(v WeekStartDay) bool {
	return isOneOf(v, WeekStartSunday, WeekStartMonday)
}

func IsValidLanguageCode(v LanguageCode) bool {
	return isOneOf(v, LanguageEnglish, LanguageSimplifiedChinese, LanguageJapanese)
}

func IsValidTimeDisplayFormat(v TimeDisplayFormat) bool {
	return isOneOf(v, TimeDisplay12Hour, TimeDisplay24Hour)
}

func IsValidDefaultTimeZoneMode(v DefaultTimeZoneMode) bool {
	return isOneOf(v, DefaultTimeZoneModeFloating)
}

func IsValidOperationEventType(v OperationEventType) bool {
	return isOneOf(v,
		OperationCompleted,
		OperationPostponed,
		OperationActivatedFromPostponed,
		OperationOnHold,
		OperationAbandoned,
		OperationRestored,
		OperationReopened,
		OperationDeleted,
		OperationArchived,
		OperationSkipped,
		OperationGeneratedNextInstance,
		OperationConvertedToRecurringTemplate,
	)
}

func IsValidTodayReasonKey(v TodayReasonKey) bool {
	return isOneOf(v,
		TodayReasonScheduledToday,
		TodayReasonRecurringToday,
		TodayReasonCarriedForward,
		TodayReasonPlannedToday,
		TodayReasonDeadlineToday,
		TodayReasonDeadlineTomorrow,
		TodayReasonDeadlineSoon,
		TodayReasonSmallTask,
		TodayReasonHighImportance,
		TodayReasonPostponedSeveralTimes,
		TodayReasonReviewDateReached,
	)
}

func IsValidErrorCode(v ErrorCode) bool {
	return isOneOf(v,
		ErrorUnauthorized,
		ErrorForbidden,
		ErrorNotFound,
		ErrorValidationFailed,
		ErrorInvalidTransition,
		ErrorUnsupportedOperation,
		ErrorDuplicateRecurringInstance,
		ErrorServiceUnavailable,
	)
}

func IsValidFieldKey(v FieldKey) bool {
	return isOneOf(v,
		FieldTitle,
		FieldItemType,
		FieldStatus,
		FieldScheduledDate,
		FieldDeadlineDate,
		FieldDeadlineTime,
		FieldDeadlineAt,
		FieldDeadlineTimeZoneMode,
		FieldUserID,
		FieldIdempotencyKey,
		FieldRecurringSequence,
		FieldLanguage,
		FieldLocale,
		FieldWeekStartDay,
		FieldTimeZone,
		FieldDateDisplayFormat,
		FieldTimeDisplayFormat,
		FieldDefaultTimeZoneMode,
		FieldTodayPrimaryLookaheadDays,
		FieldDeadlineAwarenessDays,
	)
}

func isOneOf[T comparable](value T, allowed ...T) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
