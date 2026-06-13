package domain

import "strings"

type Item struct {
	Title                string
	ItemType             ItemType
	Status               BusinessStatus
	ScheduledDate        string
	DeadlineDate         string
	DeadlineTime         string
	DeadlineAt           string
	DeadlineTimeZoneMode DeadlineTimeZoneMode
	HiddenReason         HiddenReason
	DeletedAt            string
	ArchivedAt           string
}

func ValidateItem(item Item) *Error {
	err := validationError("item is invalid")

	if strings.TrimSpace(item.Title) == "" {
		err.addField(FieldTitle, "required")
	}
	if !IsValidItemType(item.ItemType) {
		err.addField(FieldItemType, "invalid")
	}
	if !IsValidBusinessStatus(item.Status) {
		err.addField(FieldStatus, "invalid")
	}
	if item.HiddenReason != "" && !IsValidHiddenReason(item.HiddenReason) {
		err.addField(FieldStatus, "invalid_hidden_reason")
	}

	switch item.ItemType {
	case ItemTypeDateTask:
		if item.ScheduledDate == "" {
			err.addField(FieldScheduledDate, "required_for_date_task")
		}
	case ItemTypeDeadlineTask:
		if deadlineErr := ValidateDeadline(DeadlineShape{
			Date:         item.DeadlineDate,
			Time:         item.DeadlineTime,
			At:           item.DeadlineAt,
			TimeZoneMode: item.DeadlineTimeZoneMode,
		}); deadlineErr != nil {
			for field, reason := range deadlineErr.Fields {
				err.addField(field, reason)
			}
		}
	case ItemTypeIdea:
		addIdeaDeadlineErrors(err, item)
	}

	if err.hasFields() {
		return err
	}
	return nil
}

func addIdeaDeadlineErrors(err *Error, item Item) {
	reason := "not_allowed_for_idea"
	if item.DeadlineDate != "" {
		err.addField(FieldDeadlineDate, reason)
	}
	if item.DeadlineTime != "" {
		err.addField(FieldDeadlineTime, reason)
	}
	if item.DeadlineAt != "" {
		err.addField(FieldDeadlineAt, reason)
	}
	if item.DeadlineTimeZoneMode != "" {
		err.addField(FieldDeadlineTimeZoneMode, reason)
	}
}

type VisibilityState string

const (
	VisibilityDeleted   VisibilityState = "deleted"
	VisibilityArchived  VisibilityState = "archived"
	VisibilityHidden    VisibilityState = "hidden"
	VisibilityActive    VisibilityState = "active"
	VisibilityCompleted VisibilityState = "completed"
	VisibilityPostponed VisibilityState = "postponed"
	VisibilityOnHold    VisibilityState = "on_hold"
	VisibilityAbandoned VisibilityState = "abandoned"
)

func VisibilityForItem(item Item) VisibilityState {
	if item.DeletedAt != "" {
		return VisibilityDeleted
	}
	if item.ArchivedAt != "" {
		return VisibilityArchived
	}
	if item.HiddenReason != "" {
		return VisibilityHidden
	}

	switch item.Status {
	case StatusCompleted:
		return VisibilityCompleted
	case StatusPostponed:
		return VisibilityPostponed
	case StatusOnHold:
		return VisibilityOnHold
	case StatusAbandoned:
		return VisibilityAbandoned
	default:
		return VisibilityActive
	}
}

func IsNormalPlanningVisible(item Item) bool {
	return VisibilityForItem(item) == VisibilityActive
}
