package domain

func IsAllowedStatusTransition(from, to BusinessStatus) bool {
	switch from {
	case StatusActive:
		return isOneOf(to, StatusCompleted, StatusPostponed, StatusOnHold, StatusAbandoned)
	case StatusPostponed:
		return isOneOf(to, StatusActive, StatusCompleted, StatusOnHold, StatusAbandoned)
	case StatusOnHold:
		return isOneOf(to, StatusActive, StatusAbandoned)
	case StatusAbandoned:
		return to == StatusActive
	case StatusCompleted:
		return to == StatusActive
	default:
		return false
	}
}

func ValidateStatusTransition(from, to BusinessStatus) *Error {
	if !IsValidBusinessStatus(from) || !IsValidBusinessStatus(to) || !IsAllowedStatusTransition(from, to) {
		return invalidTransitionError(from, to)
	}
	return nil
}
