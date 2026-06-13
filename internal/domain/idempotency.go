package domain

import "fmt"

type RecurrenceAction string

const (
	RecurrenceActionComplete RecurrenceAction = "complete"
	RecurrenceActionSkip     RecurrenceAction = "skip"
)

func IsValidRecurrenceAction(action RecurrenceAction) bool {
	return isOneOf(action, RecurrenceActionComplete, RecurrenceActionSkip)
}

func ActionIdempotencyKey(userID, deviceID, clientActionID string) string {
	return fmt.Sprintf("action:%s:%s:%s", userID, deviceID, clientActionID)
}

func RecurrenceActionIdempotencyKey(userID, recurringTemplateID string, recurringSequence int, action RecurrenceAction) (string, *Error) {
	if !IsValidRecurrenceAction(action) {
		err := validationError("recurrence action is invalid")
		err.addField(FieldIdempotencyKey, "invalid_recurrence_action")
		return "", err
	}
	if recurringSequence < 1 {
		err := validationError("recurring sequence is invalid")
		err.addField(FieldRecurringSequence, "must_be_positive")
		return "", err
	}
	return fmt.Sprintf("recurrence:%s:%s:%d:%s", userID, recurringTemplateID, recurringSequence, action), nil
}

func RecurrenceGenerationIdempotencyKey(userID, recurringTemplateID string, nextSequence int) (string, *Error) {
	if nextSequence < 1 {
		err := validationError("recurring sequence is invalid")
		err.addField(FieldRecurringSequence, "must_be_positive")
		return "", err
	}
	return fmt.Sprintf("recurrence:%s:%s:%d:generate_next", userID, recurringTemplateID, nextSequence), nil
}

func PostponedActivationIdempotencyKey(userID, itemID, activationDate string) string {
	return fmt.Sprintf("postponed-activation:%s:%s:%s", userID, itemID, activationDate)
}
