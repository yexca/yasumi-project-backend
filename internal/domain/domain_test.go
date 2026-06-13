package domain

import "testing"

func TestValidateItemAcceptsValidMVPShapes(t *testing.T) {
	rows := []Item{
		{
			Title:    "Capture a loose thought",
			ItemType: ItemTypeInbox,
			Status:   StatusActive,
		},
		{
			Title:                 "Prepare weekly notes",
			ItemType:              ItemTypeDateTask,
			Status:                StatusActive,
			ScheduledDate:         "2026-06-14",
			ScheduledTimeZoneMode: ScheduledTimeZoneModeFloating,
		},
		{
			Title:                "Submit quarterly report",
			ItemType:             ItemTypeDeadlineTask,
			Status:               StatusActive,
			DeadlineDate:         "2026-06-16",
			DeadlineTimeZoneMode: DeadlineTimeZoneModeDateOnly,
		},
		{
			Title:    "Research calmer planning language",
			ItemType: ItemTypeIdea,
			Status:   StatusActive,
		},
	}

	for _, row := range rows {
		if err := ValidateItem(row); err != nil {
			t.Fatalf("%s ValidateItem() error = %v fields=%v", row.ItemType, err, err.Fields)
		}
	}
}

func TestValidateItemRejectsInvalidItemShapes(t *testing.T) {
	tests := []struct {
		name       string
		item       Item
		field      FieldKey
		wantReason string
	}{
		{
			name: "empty title",
			item: Item{
				Title:    "  ",
				ItemType: ItemTypeInbox,
				Status:   StatusActive,
			},
			field:      FieldTitle,
			wantReason: "required",
		},
		{
			name: "invalid item type",
			item: Item{
				Title:    "Bad type",
				ItemType: "task",
				Status:   StatusActive,
			},
			field:      FieldItemType,
			wantReason: "invalid",
		},
		{
			name: "invalid status",
			item: Item{
				Title:    "Bad status",
				ItemType: ItemTypeInbox,
				Status:   "deleted",
			},
			field:      FieldStatus,
			wantReason: "invalid",
		},
		{
			name: "date task without scheduled date",
			item: Item{
				Title:    "Missing date",
				ItemType: ItemTypeDateTask,
				Status:   StatusActive,
			},
			field:      FieldScheduledDate,
			wantReason: "required_for_date_task",
		},
		{
			name: "date task with unsupported scheduled timezone mode",
			item: Item{
				Title:                 "Unsupported scheduled mode",
				ItemType:              ItemTypeDateTask,
				Status:                StatusActive,
				ScheduledDate:         "2026-06-14",
				ScheduledTimeZoneMode: "fixed",
			},
			field:      FieldScheduledDate,
			wantReason: "invalid_scheduled_time_zone_mode",
		},
		{
			name: "idea with deadline",
			item: Item{
				Title:                "Idea with deadline",
				ItemType:             ItemTypeIdea,
				Status:               StatusActive,
				DeadlineDate:         "2026-06-20",
				DeadlineTimeZoneMode: DeadlineTimeZoneModeDateOnly,
			},
			field:      FieldDeadlineDate,
			wantReason: "not_allowed_for_idea",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateItem(tt.item)
			if err == nil {
				t.Fatal("ValidateItem() error = nil, want error")
			}
			if err.Code != ErrorValidationFailed {
				t.Fatalf("Code = %q, want %q", err.Code, ErrorValidationFailed)
			}
			if got := err.Fields[tt.field]; got != tt.wantReason {
				t.Fatalf("field %q reason = %q, want %q", tt.field, got, tt.wantReason)
			}
		})
	}
}

func TestValidateDeadlineModes(t *testing.T) {
	valid := []DeadlineShape{
		{Date: "2026-06-20", TimeZoneMode: DeadlineTimeZoneModeDateOnly},
		{Date: "2026-06-20", Time: "17:00:00", At: "2026-06-20T08:00:00Z", TimeZoneMode: DeadlineTimeZoneModeFloating},
		{At: "2026-06-20T08:00:00Z", TimeZoneMode: DeadlineTimeZoneModeFixed},
	}

	for _, shape := range valid {
		if err := ValidateDeadline(shape); err != nil {
			t.Fatalf("ValidateDeadline(%+v) error = %v fields=%v", shape, err, err.Fields)
		}
	}

	invalid := []struct {
		name       string
		shape      DeadlineShape
		field      FieldKey
		wantReason string
	}{
		{
			name:       "time without date",
			shape:      DeadlineShape{Time: "17:00:00", TimeZoneMode: DeadlineTimeZoneModeFloating},
			field:      FieldDeadlineDate,
			wantReason: "required_for_floating_deadline",
		},
		{
			name:       "fixed mixed with local fields",
			shape:      DeadlineShape{Date: "2026-06-20", Time: "17:00:00", At: "2026-06-20T08:00:00Z", TimeZoneMode: DeadlineTimeZoneModeFixed},
			field:      FieldDeadlineAt,
			wantReason: "fixed_deadline_must_not_mix_local_fields",
		},
		{
			name:       "missing mode",
			shape:      DeadlineShape{Date: "2026-06-20"},
			field:      FieldDeadlineTimeZoneMode,
			wantReason: "invalid_or_missing_deadline_mode",
		},
	}

	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeadline(tt.shape)
			if err == nil {
				t.Fatal("ValidateDeadline() error = nil, want error")
			}
			if got := err.Fields[tt.field]; got != tt.wantReason {
				t.Fatalf("field %q reason = %q, want %q", tt.field, got, tt.wantReason)
			}
		})
	}
}

func TestStatusTransitions(t *testing.T) {
	allowed := [][2]BusinessStatus{
		{StatusActive, StatusCompleted},
		{StatusActive, StatusPostponed},
		{StatusActive, StatusOnHold},
		{StatusActive, StatusAbandoned},
		{StatusPostponed, StatusActive},
		{StatusPostponed, StatusCompleted},
		{StatusPostponed, StatusOnHold},
		{StatusPostponed, StatusAbandoned},
		{StatusOnHold, StatusActive},
		{StatusOnHold, StatusAbandoned},
		{StatusAbandoned, StatusActive},
		{StatusCompleted, StatusActive},
	}

	for _, pair := range allowed {
		if err := ValidateStatusTransition(pair[0], pair[1]); err != nil {
			t.Fatalf("ValidateStatusTransition(%q, %q) error = %v", pair[0], pair[1], err)
		}
	}

	rejected := [][2]BusinessStatus{
		{StatusCompleted, StatusPostponed},
		{StatusCompleted, StatusAbandoned},
		{StatusOnHold, StatusCompleted},
		{StatusAbandoned, StatusCompleted},
		{StatusActive, StatusActive},
	}

	for _, pair := range rejected {
		err := ValidateStatusTransition(pair[0], pair[1])
		if err == nil {
			t.Fatalf("ValidateStatusTransition(%q, %q) error = nil, want error", pair[0], pair[1])
		}
		if err.Code != ErrorInvalidTransition {
			t.Fatalf("Code = %q, want %q", err.Code, ErrorInvalidTransition)
		}
		if got := err.Fields[FieldStatus]; got != "transition_not_allowed" {
			t.Fatalf("status reason = %q, want transition_not_allowed", got)
		}
	}
}

func TestMetadataPrecedence(t *testing.T) {
	rows := []struct {
		item Item
		want VisibilityState
	}{
		{
			item: Item{
				Status:       StatusActive,
				DeletedAt:    "2026-06-12T09:00:00Z",
				ArchivedAt:   "2026-06-11T09:00:00Z",
				HiddenReason: HiddenReasonConvertedToRecurringTemplate,
			},
			want: VisibilityDeleted,
		},
		{
			item: Item{
				Status:       StatusActive,
				ArchivedAt:   "2026-06-11T09:00:00Z",
				HiddenReason: HiddenReasonConvertedToRecurringTemplate,
			},
			want: VisibilityArchived,
		},
		{
			item: Item{Status: StatusCompleted},
			want: VisibilityCompleted,
		},
	}

	for _, row := range rows {
		if got := VisibilityForItem(row.item); got != row.want {
			t.Fatalf("VisibilityForItem(%+v) = %q, want %q", row.item, got, row.want)
		}
		if IsNormalPlanningVisible(row.item) {
			t.Fatalf("IsNormalPlanningVisible(%+v) = true, want false", row.item)
		}
	}
}

func TestDefaultSettings(t *testing.T) {
	tests := []struct {
		name       string
		language   LanguageCode
		timeZone   string
		wantLocale string
		wantWeek   WeekStartDay
		wantTime   TimeDisplayFormat
	}{
		{"english", LanguageEnglish, "America/Los_Angeles", "en-US", WeekStartSunday, TimeDisplay12Hour},
		{"simplified chinese", LanguageSimplifiedChinese, "Asia/Shanghai", "zh-CN", WeekStartMonday, TimeDisplay24Hour},
		{"japanese", LanguageJapanese, "Asia/Tokyo", "ja-JP", WeekStartMonday, TimeDisplay24Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := DefaultSettings(tt.language, tt.timeZone)
			if settings.Locale != tt.wantLocale {
				t.Fatalf("Locale = %q, want %q", settings.Locale, tt.wantLocale)
			}
			if settings.WeekStartDay != tt.wantWeek {
				t.Fatalf("WeekStartDay = %q, want %q", settings.WeekStartDay, tt.wantWeek)
			}
			if settings.TimeDisplayFormat != tt.wantTime {
				t.Fatalf("TimeDisplayFormat = %q, want %q", settings.TimeDisplayFormat, tt.wantTime)
			}
			if settings.TodayPrimaryLookaheadDays != 3 || settings.DeadlineAwarenessDays != 14 {
				t.Fatalf("recommendation defaults = %d/%d, want 3/14", settings.TodayPrimaryLookaheadDays, settings.DeadlineAwarenessDays)
			}
			if err := ValidateSettings(settings); err != nil {
				t.Fatalf("ValidateSettings() error = %v fields=%v", err, err.Fields)
			}
		})
	}
}

func TestIdempotencyKeyBuilders(t *testing.T) {
	userID := "00000000-0000-4000-8000-000000000001"
	deviceID := "device-alpha"
	templateID := "50000000-0000-4000-8000-000000000001"

	if got := ActionIdempotencyKey(userID, deviceID, "action-1"); got != "action:00000000-0000-4000-8000-000000000001:device-alpha:action-1" {
		t.Fatalf("ActionIdempotencyKey() = %q", got)
	}

	actionKey, err := RecurrenceActionIdempotencyKey(userID, templateID, 1, RecurrenceActionComplete)
	if err != nil {
		t.Fatalf("RecurrenceActionIdempotencyKey() error = %v", err)
	}
	if actionKey != "recurrence:00000000-0000-4000-8000-000000000001:50000000-0000-4000-8000-000000000001:1:complete" {
		t.Fatalf("recurrence action key = %q", actionKey)
	}

	generationKey, err := RecurrenceGenerationIdempotencyKey(userID, templateID, 2)
	if err != nil {
		t.Fatalf("RecurrenceGenerationIdempotencyKey() error = %v", err)
	}
	if generationKey != "recurrence:00000000-0000-4000-8000-000000000001:50000000-0000-4000-8000-000000000001:2:generate_next" {
		t.Fatalf("recurrence generation key = %q", generationKey)
	}

	if got := PostponedActivationIdempotencyKey(userID, "20000000-0000-4000-8000-000000000001", "2026-06-13"); got != "postponed-activation:00000000-0000-4000-8000-000000000001:20000000-0000-4000-8000-000000000001:2026-06-13" {
		t.Fatalf("PostponedActivationIdempotencyKey() = %q", got)
	}
}
