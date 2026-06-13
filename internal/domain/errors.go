package domain

import "strings"

type Error struct {
	Code      ErrorCode
	Message   string
	Fields    map[FieldKey]string
	Retryable bool
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	return string(e.Code)
}

func (e *Error) addField(field FieldKey, reason string) {
	if e.Fields == nil {
		e.Fields = map[FieldKey]string{}
	}
	e.Fields[field] = reason
}

func (e *Error) hasFields() bool {
	return e != nil && len(e.Fields) > 0
}

func validationError(message string) *Error {
	return &Error{
		Code:      ErrorValidationFailed,
		Message:   message,
		Fields:    map[FieldKey]string{},
		Retryable: false,
	}
}

func invalidTransitionError(from, to BusinessStatus) *Error {
	err := &Error{
		Code:      ErrorInvalidTransition,
		Message:   strings.TrimSpace("status transition is not allowed: " + string(from) + " -> " + string(to)),
		Fields:    map[FieldKey]string{},
		Retryable: false,
	}
	err.addField(FieldStatus, "transition_not_allowed")
	return err
}
