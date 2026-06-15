package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/yasumi/yasumi-project-backend/internal/domain"
	"github.com/yasumi/yasumi-project-backend/internal/repository"
)

func decodeItemRow(raw json.RawMessage) (repository.ItemRecord, error) {
	var row repository.ItemRecord
	if err := json.Unmarshal(raw, &row); err != nil {
		return repository.ItemRecord{}, invalidJSONError("item row")
	}
	return row, nil
}

func decodeAreaRow(raw json.RawMessage) (repository.AreaRecord, error) {
	var row repository.AreaRecord
	if err := json.Unmarshal(raw, &row); err != nil {
		return repository.AreaRecord{}, invalidJSONError("area row")
	}
	return row, nil
}

func decodeRecurringTemplateRow(raw json.RawMessage) (repository.RecurringTemplateRecord, error) {
	var row repository.RecurringTemplateRecord
	if err := json.Unmarshal(raw, &row); err != nil {
		return repository.RecurringTemplateRecord{}, invalidJSONError("recurring template row")
	}
	return row, nil
}

func decodeOperationRow(raw json.RawMessage) (repository.OperationHistoryRecord, error) {
	var row repository.OperationHistoryRecord
	if err := json.Unmarshal(raw, &row); err != nil {
		return repository.OperationHistoryRecord{}, invalidJSONError("operation history row")
	}
	return row, nil
}

func decodeUserSettingsRow(raw json.RawMessage) (repository.UserSettingsRecord, error) {
	var row repository.UserSettingsRecord
	if err := json.Unmarshal(raw, &row); err != nil {
		return repository.UserSettingsRecord{}, invalidJSONError("user settings row")
	}
	return row, nil
}

func normalizeOwnedRow(authenticatedUserID string, rowUserID *string) error {
	*rowUserID = strings.TrimSpace(*rowUserID)
	if *rowUserID == "" {
		*rowUserID = authenticatedUserID
		return nil
	}
	if *rowUserID != authenticatedUserID {
		return &domain.Error{
			Code:      domain.ErrorForbidden,
			Message:   "row is owned by another user",
			Fields:    map[domain.FieldKey]string{domain.FieldUserID: "owner_mismatch"},
			Retryable: false,
		}
	}
	return nil
}

func unsupportedOperation(table, op string) error {
	return &domain.Error{
		Code:      domain.ErrorUnsupportedOperation,
		Message:   fmt.Sprintf("unsupported sync mutation %s on %s", op, table),
		Fields:    map[domain.FieldKey]string{},
		Retryable: false,
	}
}

func validationFieldError(message string, field domain.FieldKey, reason string) error {
	return &domain.Error{
		Code:      domain.ErrorValidationFailed,
		Message:   message,
		Fields:    map[domain.FieldKey]string{field: reason},
		Retryable: false,
	}
}

func invalidJSONError(name string) error {
	return &domain.Error{
		Code:      domain.ErrorValidationFailed,
		Message:   name + " is invalid JSON",
		Fields:    map[domain.FieldKey]string{},
		Retryable: false,
	}
}

func mapRepositoryError(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return &domain.Error{
			Code:      domain.ErrorForbidden,
			Message:   "row is not writable by this user",
			Fields:    map[domain.FieldKey]string{domain.FieldUserID: "owner_mismatch"},
			Retryable: false,
		}
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "items_unique_recurring_instance" {
		return &domain.Error{
			Code:      domain.ErrorDuplicateRecurringInstance,
			Message:   "generated recurring instance already exists",
			Fields:    map[domain.FieldKey]string{domain.FieldRecurringSequence: "duplicate_generated_instance"},
			Retryable: false,
		}
	}
	return &domain.Error{
		Code:      domain.ErrorValidationFailed,
		Message:   "sync write failed validation",
		Fields:    map[domain.FieldKey]string{},
		Retryable: false,
	}
}

func defaultTime(value time.Time, fallback time.Time) time.Time {
	if value.IsZero() {
		return fallback
	}
	return value.UTC()
}

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func timeString(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func sameTimePtr(a, b *time.Time) bool {
	if a == nil || a.IsZero() {
		return b == nil || b.IsZero()
	}
	if b == nil || b.IsZero() {
		return false
	}
	return a.Equal(*b)
}

func isNilTime(value *time.Time) bool {
	return value == nil || value.IsZero()
}
