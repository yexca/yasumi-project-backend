package domain

import "strings"

type Area struct {
	Name         string
	HiddenReason HiddenReason
}

func ValidateArea(area Area) *Error {
	err := validationError("area is invalid")

	if strings.TrimSpace(area.Name) == "" {
		err.addField(FieldName, "required")
	}
	if area.HiddenReason != "" && !IsValidHiddenReason(area.HiddenReason) {
		err.addField(FieldStatus, "invalid_hidden_reason")
	}

	if err.hasFields() {
		return err
	}
	return nil
}

type RecurringTemplate struct {
	Title           string
	Frequency       RecurrenceFrequency
	Interval        int
	RecurrenceBasis RecurrenceBasis
	StartDate       string
	EndType         RecurrenceEndType
	EndDate         string
	EndAfterCount   *int
	CompletedCount  int
	NextSequence    int
	Status          RecurringTemplateStatus
	HiddenReason    HiddenReason
}

func ValidateRecurringTemplate(template RecurringTemplate) *Error {
	err := validationError("recurring template is invalid")

	if strings.TrimSpace(template.Title) == "" {
		err.addField(FieldTitle, "required")
	}
	if !IsValidRecurrenceFrequency(template.Frequency) {
		err.addField(FieldFrequency, "invalid")
	}
	if template.Interval < 1 {
		err.addField(FieldInterval, "must_be_positive")
	}
	if !IsValidRecurrenceBasis(template.RecurrenceBasis) {
		err.addField(FieldRecurrenceBasis, "invalid")
	}
	if template.StartDate == "" {
		err.addField(FieldStartDate, "required")
	}
	if !IsValidRecurrenceEndType(template.EndType) {
		err.addField(FieldEndType, "invalid")
	}
	if template.CompletedCount < 0 {
		err.addField(FieldRecurringSequence, "completed_count_must_be_non_negative")
	}
	if template.NextSequence < 1 {
		err.addField(FieldRecurringSequence, "must_be_positive")
	}
	if !IsValidRecurringTemplateStatus(template.Status) {
		err.addField(FieldStatus, "invalid")
	}
	if template.HiddenReason != "" && !IsValidHiddenReason(template.HiddenReason) {
		err.addField(FieldStatus, "invalid_hidden_reason")
	}

	switch template.EndType {
	case RecurrenceEndTypeNever:
		if template.EndDate != "" || template.EndAfterCount != nil {
			err.addField(FieldEndType, "never_must_not_set_end_fields")
		}
	case RecurrenceEndTypeOnDate:
		if template.EndDate == "" {
			err.addField(FieldEndDate, "required_for_on_date")
		}
		if template.EndAfterCount != nil {
			err.addField(FieldEndType, "on_date_must_not_set_end_after_count")
		}
	case RecurrenceEndTypeAfterCount:
		if template.EndAfterCount == nil || *template.EndAfterCount < 1 {
			err.addField(FieldEndType, "after_count_must_be_positive")
		}
		if template.EndDate != "" {
			err.addField(FieldEndType, "after_count_must_not_set_end_date")
		}
	}

	if err.hasFields() {
		return err
	}
	return nil
}
