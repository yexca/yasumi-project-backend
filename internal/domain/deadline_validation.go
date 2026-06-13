package domain

type DeadlineShape struct {
	Date         string
	Time         string
	At           string
	TimeZoneMode DeadlineTimeZoneMode
}

func ValidateDeadline(shape DeadlineShape) *Error {
	err := validationError("deadline shape is invalid")

	switch shape.TimeZoneMode {
	case DeadlineTimeZoneModeDateOnly:
		if shape.Date == "" {
			err.addField(FieldDeadlineDate, "required_for_date_only_deadline")
		}
		if shape.Time != "" {
			err.addField(FieldDeadlineTime, "date_only_deadline_must_not_set_time")
		}
		if shape.At != "" {
			err.addField(FieldDeadlineAt, "date_only_deadline_must_not_set_fixed_instant")
		}
	case DeadlineTimeZoneModeFloating:
		if shape.Date == "" {
			err.addField(FieldDeadlineDate, "required_for_floating_deadline")
		}
		if shape.Time == "" {
			err.addField(FieldDeadlineTime, "required_for_floating_deadline")
		}
	case DeadlineTimeZoneModeFixed:
		if shape.At == "" {
			err.addField(FieldDeadlineAt, "required_for_fixed_deadline")
		}
		if shape.Date != "" || shape.Time != "" {
			reason := "fixed_deadline_must_not_mix_local_fields"
			err.addField(FieldDeadlineAt, reason)
			if shape.Date != "" {
				err.addField(FieldDeadlineDate, reason)
			}
			if shape.Time != "" {
				err.addField(FieldDeadlineTime, reason)
			}
		}
	default:
		err.addField(FieldDeadlineTimeZoneMode, "invalid_or_missing_deadline_mode")
	}

	if err.hasFields() {
		return err
	}
	return nil
}
