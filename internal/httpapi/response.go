package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yasumi/yasumi-project-backend/internal/auth"
	"github.com/yasumi/yasumi-project-backend/internal/domain"
)

type apiError struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields"`
	Retryable bool              `json:"retryable"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func writeAPIError(w http.ResponseWriter, status int, err apiError) {
	if err.Fields == nil {
		err.Fields = map[string]string{}
	}
	writeJSON(w, status, err)
}

func domainError(err error) (int, apiError) {
	var domainErr *domain.Error
	if errors.As(err, &domainErr) {
		return statusForDomainCode(domainErr.Code), apiError{
			Code:      string(domainErr.Code),
			Message:   domainErr.Error(),
			Fields:    fieldMap(domainErr.Fields),
			Retryable: domainErr.Retryable,
		}
	}
	return http.StatusServiceUnavailable, apiError{
		Code:      string(domain.ErrorServiceUnavailable),
		Message:   "service is unavailable",
		Fields:    map[string]string{},
		Retryable: true,
	}
}

func authOrDomainError(err error) (int, apiError) {
	if errors.Is(err, auth.ErrUnauthenticated) {
		return http.StatusUnauthorized, apiError{
			Code:      string(domain.ErrorUnauthorized),
			Message:   "invalid session",
			Fields:    map[string]string{},
			Retryable: false,
		}
	}
	return domainError(err)
}

func statusForDomainCode(code domain.ErrorCode) int {
	switch code {
	case domain.ErrorUnauthorized:
		return http.StatusUnauthorized
	case domain.ErrorForbidden:
		return http.StatusForbidden
	case domain.ErrorNotFound:
		return http.StatusNotFound
	case domain.ErrorValidationFailed:
		return http.StatusBadRequest
	case domain.ErrorInvalidCredentials, domain.ErrorSessionExpired:
		return http.StatusUnauthorized
	case domain.ErrorAccountDisabled:
		return http.StatusForbidden
	case domain.ErrorUsernameAlreadyTaken, domain.ErrorEmailAlreadyRegistered, domain.ErrorInvalidTransition, domain.ErrorUnsupportedOperation, domain.ErrorDuplicateRecurringInstance:
		return http.StatusConflict
	case domain.ErrorServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func fieldMap(fields map[domain.FieldKey]string) map[string]string {
	out := make(map[string]string, len(fields))
	for key, value := range fields {
		out[string(key)] = value
	}
	return out
}
