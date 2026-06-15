package service

import (
	"context"
	"errors"
	"strings"

	"github.com/yasumi/yasumi-project-backend/internal/domain"
	"github.com/yasumi/yasumi-project-backend/internal/repository"
)

func (s *SyncUploadService) acceptItem(ctx context.Context, tx syncTx, userID, deviceID string, companionOps map[string]operationRow, mutation SyncMutation) (*AcceptedWrite, error) {
	row, err := decodeItemRow(mutation.Row)
	if err != nil {
		return nil, err
	}
	if err := normalizeOwnedRow(userID, &row.UserID); err != nil {
		return nil, err
	}
	if row.ID == "" {
		return nil, validationFieldError("item is invalid", domain.FieldItemType, "id_required")
	}
	if err := domain.ValidateItem(domain.Item{
		Title:                 row.Title,
		ItemType:              domain.ItemType(row.ItemType),
		Status:                domain.BusinessStatus(row.Status),
		ScheduledDate:         deref(row.ScheduledDate),
		ScheduledTimeZoneMode: domain.ScheduledTimeZoneMode(deref(row.ScheduledTimeZoneMode)),
		DeadlineDate:          deref(row.DeadlineDate),
		DeadlineTime:          deref(row.DeadlineTime),
		DeadlineAt:            timeString(row.DeadlineAt),
		DeadlineTimeZoneMode:  domain.DeadlineTimeZoneMode(deref(row.DeadlineTimeZoneMode)),
		HiddenReason:          domain.HiddenReason(deref(row.HiddenReason)),
		DeletedAt:             timeString(row.DeletedAt),
		ArchivedAt:            timeString(row.ArchivedAt),
	}); err != nil {
		return nil, err
	}

	now := s.clock.Now().UTC()
	previous, err := tx.GetItemForUpdateByUser(ctx, userID, row.ID)
	exists := err == nil
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, mapRepositoryError(err)
	}

	if exists {
		if err := validateSemanticItemUpdate(previous, row, mutation.ClientObserved, companionOps[row.ID]); err != nil {
			return nil, err
		}
		row.CreatedAt = previous.CreatedAt
		row.CreatedByDeviceID = previous.CreatedByDeviceID
		row.Revision = previous.Revision + 1
	} else {
		row.CreatedAt = defaultTime(row.CreatedAt, now)
		if strings.TrimSpace(row.CreatedByDeviceID) == "" {
			row.CreatedByDeviceID = deviceID
		}
		row.Revision = 1
	}

	row.UpdatedAt = now
	row.ServerUpdatedAt = now
	row.ClientUpdatedAt = defaultTime(row.ClientUpdatedAt, now)
	if strings.TrimSpace(row.UpdatedByDeviceID) == "" {
		row.UpdatedByDeviceID = deviceID
	}
	if len(row.PressureMetadata) == 0 {
		row.PressureMetadata = []byte("{}")
	}

	if err := tx.UpsertItem(ctx, row); err != nil {
		return nil, mapRepositoryError(err)
	}
	return &AcceptedWrite{
		Table:    "items",
		ID:       row.ID,
		Op:       mutation.Op,
		Revision: &row.Revision,
	}, nil
}

func validateSemanticItemUpdate(previous repository.ItemRecord, next repository.ItemRecord, observed ObservedState, op operationRow) error {
	semantic := previous.Status != next.Status ||
		!sameTimePtr(previous.DeletedAt, next.DeletedAt) ||
		!sameTimePtr(previous.ArchivedAt, next.ArchivedAt) ||
		deref(previous.HiddenReason) != deref(next.HiddenReason)
	if !semantic {
		return nil
	}

	hasIntent := op.ID != "" || observed.Revision != nil
	if !hasIntent {
		return &domain.Error{
			Code:      domain.ErrorInvalidTransition,
			Message:   "semantic item update is missing action intent",
			Fields:    map[domain.FieldKey]string{domain.FieldStatus: "action_intent_required"},
			Retryable: false,
		}
	}
	if observed.Revision != nil {
		if previous.Revision != *observed.Revision {
			return &domain.Error{
				Code:      domain.ErrorInvalidTransition,
				Message:   "observed revision does not match accepted server state",
				Fields:    map[domain.FieldKey]string{domain.FieldStatus: "observed_revision_mismatch"},
				Retryable: false,
			}
		}
		if previous.Status != deref(observed.Status) ||
			timeString(previous.DeletedAt) != deref(observed.DeletedAt) ||
			timeString(previous.ArchivedAt) != deref(observed.ArchivedAt) ||
			deref(previous.HiddenReason) != deref(observed.HiddenReason) {
			return &domain.Error{
				Code:      domain.ErrorInvalidTransition,
				Message:   "observed state does not match accepted server state",
				Fields:    map[domain.FieldKey]string{domain.FieldStatus: "observed_state_mismatch"},
				Retryable: false,
			}
		}
	}
	if err := validateSemanticMetadataTransition(previous, next, op); err != nil {
		return err
	}
	if previous.Status != next.Status {
		if err := domain.ValidateStatusTransition(domain.BusinessStatus(previous.Status), domain.BusinessStatus(next.Status)); err != nil {
			return err
		}
		if op.ID != "" && !operationMatchesStatus(op.EventType, next.Status) {
			return &domain.Error{
				Code:      domain.ErrorInvalidTransition,
				Message:   "operation intent does not match status change",
				Fields:    map[domain.FieldKey]string{domain.FieldStatus: "operation_intent_mismatch"},
				Retryable: false,
			}
		}
		if op.ID != "" && domain.OperationEventType(op.EventType) == domain.OperationActivatedFromPostponed {
			if err := validatePostponedActivation(previous, next, op); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateSemanticMetadataTransition(previous repository.ItemRecord, next repository.ItemRecord, op operationRow) error {
	eventType := domain.OperationEventType(op.EventType)

	if !sameTimePtr(previous.DeletedAt, next.DeletedAt) && op.ID != "" {
		switch {
		case isNilTime(previous.DeletedAt) && !isNilTime(next.DeletedAt):
			if eventType != domain.OperationDeleted {
				return semanticEventMismatch("deleted_at", "deleted")
			}
		case !isNilTime(previous.DeletedAt) && isNilTime(next.DeletedAt):
			if eventType != domain.OperationRestored {
				return semanticEventMismatch("deleted_at", "restored")
			}
		}
	}

	if !sameTimePtr(previous.ArchivedAt, next.ArchivedAt) && op.ID != "" {
		switch {
		case isNilTime(previous.ArchivedAt) && !isNilTime(next.ArchivedAt):
			if eventType != domain.OperationArchived {
				return semanticEventMismatch("archived_at", "archived")
			}
		case !isNilTime(previous.ArchivedAt) && isNilTime(next.ArchivedAt):
			if eventType != domain.OperationRestored {
				return semanticEventMismatch("archived_at", "restored")
			}
		}
	}

	if deref(previous.HiddenReason) != deref(next.HiddenReason) && op.ID != "" {
		switch {
		case deref(previous.HiddenReason) == "" && deref(next.HiddenReason) == string(domain.HiddenReasonRecurringSkipped):
			if eventType != domain.OperationSkipped {
				return semanticEventMismatch("hidden_reason", "skipped")
			}
		case deref(previous.HiddenReason) != "" && deref(next.HiddenReason) == "":
			if eventType != domain.OperationRestored {
				return semanticEventMismatch("hidden_reason", "restored")
			}
		}
	}

	return nil
}

func validatePostponedActivation(previous repository.ItemRecord, next repository.ItemRecord, op operationRow) error {
	if previous.Status != string(domain.StatusPostponed) || next.Status != string(domain.StatusActive) {
		return &domain.Error{
			Code:      domain.ErrorInvalidTransition,
			Message:   "activated_from_postponed requires postponed to active transition",
			Fields:    map[domain.FieldKey]string{domain.FieldStatus: "operation_intent_mismatch"},
			Retryable: false,
		}
	}
	if !isNilTime(previous.DeletedAt) || !isNilTime(previous.ArchivedAt) {
		return &domain.Error{
			Code:      domain.ErrorInvalidTransition,
			Message:   "postponed activation is not allowed for deleted or archived items",
			Fields:    map[domain.FieldKey]string{domain.FieldStatus: "activation_not_allowed"},
			Retryable: false,
		}
	}
	activationDate := activationDateFromKey(op.IdempotencyKey)
	if activationDate == "" {
		return &domain.Error{
			Code:      domain.ErrorValidationFailed,
			Message:   "postponed activation requires a valid idempotency key",
			Fields:    map[domain.FieldKey]string{domain.FieldIdempotencyKey: "invalid_postponed_activation_key"},
			Retryable: false,
		}
	}

	switch previous.ItemType {
	case string(domain.ItemTypeDateTask):
		if previous.ScheduledDate == nil || *previous.ScheduledDate == "" || *previous.ScheduledDate > activationDate {
			return &domain.Error{
				Code:      domain.ErrorInvalidTransition,
				Message:   "date task is not due for postponed activation",
				Fields:    map[domain.FieldKey]string{domain.FieldScheduledDate: "activation_not_due"},
				Retryable: false,
			}
		}
	case string(domain.ItemTypeDeadlineTask):
		if previous.PlannedWorkDate == nil || *previous.PlannedWorkDate == "" || *previous.PlannedWorkDate > activationDate {
			return &domain.Error{
				Code:      domain.ErrorInvalidTransition,
				Message:   "deadline task is not due for postponed activation",
				Fields:    map[domain.FieldKey]string{domain.FieldScheduledDate: "activation_not_due"},
				Retryable: false,
			}
		}
	default:
		return &domain.Error{
			Code:      domain.ErrorInvalidTransition,
			Message:   "postponed activation is only valid for task items",
			Fields:    map[domain.FieldKey]string{domain.FieldItemType: "activation_not_supported"},
			Retryable: false,
		}
	}

	return nil
}

func operationMatchesStatus(eventType, status string) bool {
	switch domain.OperationEventType(eventType) {
	case domain.OperationCompleted:
		return status == string(domain.StatusCompleted)
	case domain.OperationPostponed:
		return status == string(domain.StatusPostponed)
	case domain.OperationOnHold:
		return status == string(domain.StatusOnHold)
	case domain.OperationAbandoned:
		return status == string(domain.StatusAbandoned)
	case domain.OperationRestored, domain.OperationReopened, domain.OperationActivatedFromPostponed:
		return status == string(domain.StatusActive)
	default:
		return true
	}
}

func semanticEventMismatch(field, expectedEvent string) error {
	return &domain.Error{
		Code:      domain.ErrorInvalidTransition,
		Message:   "operation intent does not match semantic metadata change",
		Fields:    map[domain.FieldKey]string{domain.FieldStatus: field + "_requires_" + expectedEvent},
		Retryable: false,
	}
}

func activationDateFromKey(idempotencyKey *string) string {
	if idempotencyKey == nil {
		return ""
	}
	parts := strings.Split(*idempotencyKey, ":")
	if len(parts) != 4 || parts[0] != "postponed-activation" {
		return ""
	}
	return parts[3]
}
