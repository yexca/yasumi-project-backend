package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/yasumi/yasumi-project-backend/internal/domain"
	"github.com/yasumi/yasumi-project-backend/internal/repository"
)

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

type SyncUploadService struct {
	repo  syncRepository
	clock Clock
}

type RepositoryAdapter struct {
	repo *repository.Repository
}

func NewRepositoryAdapter(repo *repository.Repository) RepositoryAdapter {
	return RepositoryAdapter{repo: repo}
}

func (a RepositoryAdapter) InTx(ctx context.Context, fn func(context.Context, syncTx) error) error {
	return a.repo.InTx(ctx, func(ctx context.Context, tx *repository.Tx) error {
		return fn(ctx, tx)
	})
}

type syncRepository interface {
	InTx(ctx context.Context, fn func(context.Context, syncTx) error) error
}

type syncTx interface {
	GetAreaForUpdateByUser(ctx context.Context, userID, areaID string) (repository.AreaRecord, error)
	UpsertArea(ctx context.Context, row repository.AreaRecord) error
	GetRecurringTemplateForUpdateByUser(ctx context.Context, userID, templateID string) (repository.RecurringTemplateRecord, error)
	UpsertRecurringTemplate(ctx context.Context, row repository.RecurringTemplateRecord) error
	GetItemForUpdateByUser(ctx context.Context, userID, itemID string) (repository.ItemRecord, error)
	UpsertItem(ctx context.Context, row repository.ItemRecord) error
	InsertOperationHistory(ctx context.Context, row repository.OperationHistoryRecord) error
	FindOperationByIdempotencyKey(ctx context.Context, userID, idempotencyKey string) (repository.OperationHistoryRecord, error)
	GetUserSettingsForUpdate(ctx context.Context, userID string) (repository.UserSettingsRecord, error)
	UpsertUserSettings(ctx context.Context, row repository.UserSettingsRecord) error
}

func NewSyncUploadService(repo syncRepository, clock Clock) *SyncUploadService {
	if clock == nil {
		clock = SystemClock{}
	}
	return &SyncUploadService{repo: repo, clock: clock}
}

type SyncUpload struct {
	ClientBatchID string         `json:"client_batch_id"`
	DeviceID      string         `json:"device_id"`
	BaseRevision  *int64         `json:"base_revision"`
	Mutations     []SyncMutation `json:"mutations"`
}

type SyncMutation struct {
	Table          string          `json:"table"`
	Op             string          `json:"op"`
	Row            json.RawMessage `json:"row"`
	ClientObserved ObservedState   `json:"client_observed"`
}

type ObservedState struct {
	Revision     *int64  `json:"revision"`
	Status       *string `json:"status"`
	DeletedAt    *string `json:"deleted_at"`
	ArchivedAt   *string `json:"archived_at"`
	HiddenReason *string `json:"hidden_reason"`
}

type SyncUploadResult struct {
	ClientBatchID string          `json:"client_batch_id"`
	Accepted      []AcceptedWrite `json:"accepted"`
	DuplicateOf   *AcceptedWrite  `json:"duplicate_of,omitempty"`
}

type AcceptedWrite struct {
	Table          string  `json:"table"`
	ID             string  `json:"id"`
	Op             string  `json:"op"`
	Revision       *int64  `json:"revision,omitempty"`
	IdempotencyKey *string `json:"idempotency_key,omitempty"`
}

func (s *SyncUploadService) AcceptUpload(ctx context.Context, userID string, upload SyncUpload) (SyncUploadResult, error) {
	upload.DeviceID = strings.TrimSpace(upload.DeviceID)
	if upload.DeviceID == "" {
		return SyncUploadResult{}, validationFieldError("sync upload is invalid", domain.FieldUserID, "device_id_required")
	}
	if len(upload.Mutations) == 0 {
		return SyncUploadResult{ClientBatchID: upload.ClientBatchID, Accepted: []AcceptedWrite{}}, nil
	}

	result := SyncUploadResult{
		ClientBatchID: upload.ClientBatchID,
		Accepted:      []AcceptedWrite{},
	}
	companionOps, allOps, err := companionOperations(upload.Mutations)
	if err != nil {
		return SyncUploadResult{}, err
	}

	err = s.repo.InTx(ctx, func(ctx context.Context, tx syncTx) error {
		for _, op := range allOps {
			if op.IdempotencyKey == nil {
				continue
			}
			found, err := tx.FindOperationByIdempotencyKey(ctx, userID, *op.IdempotencyKey)
			if err == nil {
				result.DuplicateOf = &AcceptedWrite{
					Table:          "operation_history",
					ID:             found.ID,
					Op:             "insert",
					IdempotencyKey: found.IdempotencyKey,
				}
				return nil
			}
			if !errors.Is(err, repository.ErrNotFound) {
				return mapRepositoryError(err)
			}
		}

		for _, mutation := range upload.Mutations {
			accepted, err := s.acceptMutation(ctx, tx, userID, upload.DeviceID, companionOps, mutation)
			if err != nil {
				return err
			}
			if accepted != nil {
				result.Accepted = append(result.Accepted, *accepted)
			}
		}
		return nil
	})
	if err != nil {
		return SyncUploadResult{}, err
	}
	return result, nil
}

func (s *SyncUploadService) acceptMutation(ctx context.Context, tx syncTx, userID, deviceID string, companionOps map[string]operationRow, mutation SyncMutation) (*AcceptedWrite, error) {
	table := strings.TrimSpace(mutation.Table)
	op := strings.TrimSpace(mutation.Op)
	if op == "delete" {
		return nil, &domain.Error{
			Code:      domain.ErrorUnsupportedOperation,
			Message:   "physical delete is not supported for synced user data",
			Fields:    map[domain.FieldKey]string{},
			Retryable: false,
		}
	}

	switch table {
	case "areas":
		if op != "insert" && op != "update" && op != "upsert" {
			return nil, unsupportedOperation("areas", op)
		}
		return s.acceptArea(ctx, tx, userID, deviceID, mutation)
	case "recurring_task_templates":
		if op != "insert" && op != "update" && op != "upsert" {
			return nil, unsupportedOperation("recurring_task_templates", op)
		}
		return s.acceptRecurringTemplate(ctx, tx, userID, deviceID, mutation)
	case "items":
		if op != "insert" && op != "update" && op != "upsert" {
			return nil, unsupportedOperation("items", op)
		}
		return s.acceptItem(ctx, tx, userID, deviceID, companionOps, mutation)
	case "operation_history":
		if op != "insert" {
			return nil, unsupportedOperation("operation_history", op)
		}
		return s.acceptOperation(ctx, tx, userID, deviceID, mutation)
	case "user_settings":
		if op != "insert" && op != "update" && op != "upsert" {
			return nil, unsupportedOperation("user_settings", op)
		}
		return s.acceptUserSettings(ctx, tx, userID, deviceID, mutation)
	default:
		return nil, unsupportedOperation(table, op)
	}
}

func (s *SyncUploadService) acceptArea(ctx context.Context, tx syncTx, userID, deviceID string, mutation SyncMutation) (*AcceptedWrite, error) {
	row, err := decodeAreaRow(mutation.Row)
	if err != nil {
		return nil, err
	}
	if err := normalizeOwnedRow(userID, &row.UserID); err != nil {
		return nil, err
	}
	if row.ID == "" {
		return nil, validationFieldError("area is invalid", domain.FieldName, "id_required")
	}
	if err := domain.ValidateArea(domain.Area{
		Name:         row.Name,
		HiddenReason: domain.HiddenReason(deref(row.HiddenReason)),
	}); err != nil {
		return nil, err
	}

	now := s.clock.Now().UTC()
	previous, err := tx.GetAreaForUpdateByUser(ctx, userID, row.ID)
	exists := err == nil
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, mapRepositoryError(err)
	}

	if exists {
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

	if err := tx.UpsertArea(ctx, row); err != nil {
		return nil, mapRepositoryError(err)
	}
	return &AcceptedWrite{
		Table:    "areas",
		ID:       row.ID,
		Op:       mutation.Op,
		Revision: &row.Revision,
	}, nil
}

func (s *SyncUploadService) acceptRecurringTemplate(ctx context.Context, tx syncTx, userID, deviceID string, mutation SyncMutation) (*AcceptedWrite, error) {
	row, err := decodeRecurringTemplateRow(mutation.Row)
	if err != nil {
		return nil, err
	}
	if err := normalizeOwnedRow(userID, &row.UserID); err != nil {
		return nil, err
	}
	if row.ID == "" {
		return nil, validationFieldError("recurring template is invalid", domain.FieldTitle, "id_required")
	}
	if err := domain.ValidateRecurringTemplate(domain.RecurringTemplate{
		Title:           row.Title,
		Frequency:       domain.RecurrenceFrequency(row.Frequency),
		Interval:        row.Interval,
		RecurrenceBasis: domain.RecurrenceBasis(row.RecurrenceBasis),
		StartDate:       row.StartDate,
		EndType:         domain.RecurrenceEndType(row.EndType),
		EndDate:         deref(row.EndDate),
		EndAfterCount:   row.EndAfterCount,
		CompletedCount:  row.CompletedCount,
		NextSequence:    row.NextSequence,
		Status:          domain.RecurringTemplateStatus(row.Status),
		HiddenReason:    domain.HiddenReason(deref(row.HiddenReason)),
	}); err != nil {
		return nil, err
	}

	now := s.clock.Now().UTC()
	previous, err := tx.GetRecurringTemplateForUpdateByUser(ctx, userID, row.ID)
	exists := err == nil
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, mapRepositoryError(err)
	}

	if exists {
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
	if len(row.Weekdays) == 0 {
		row.Weekdays = []byte("[]")
	}
	if len(row.ReminderRule) == 0 {
		row.ReminderRule = []byte("{}")
	}
	if len(row.GeneratedTaskDefaults) == 0 {
		row.GeneratedTaskDefaults = []byte("{}")
	}

	if err := tx.UpsertRecurringTemplate(ctx, row); err != nil {
		return nil, mapRepositoryError(err)
	}
	return &AcceptedWrite{
		Table:    "recurring_task_templates",
		ID:       row.ID,
		Op:       mutation.Op,
		Revision: &row.Revision,
	}, nil
}

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

func (s *SyncUploadService) acceptOperation(ctx context.Context, tx syncTx, userID, deviceID string, mutation SyncMutation) (*AcceptedWrite, error) {
	row, err := decodeOperationRow(mutation.Row)
	if err != nil {
		return nil, err
	}
	if err := normalizeOwnedRow(userID, &row.UserID); err != nil {
		return nil, err
	}
	if row.ID == "" {
		return nil, validationFieldError("operation history is invalid", domain.FieldIdempotencyKey, "id_required")
	}
	if !domain.IsValidOperationEventType(domain.OperationEventType(row.EventType)) {
		return nil, validationFieldError("operation history is invalid", domain.FieldStatus, "invalid_event_type")
	}
	if row.IdempotencyKey != nil && strings.TrimSpace(*row.IdempotencyKey) == "" {
		return nil, validationFieldError("operation history is invalid", domain.FieldIdempotencyKey, "required")
	}
	if row.IdempotencyKey != nil {
		found, err := tx.FindOperationByIdempotencyKey(ctx, userID, *row.IdempotencyKey)
		if err == nil {
			return &AcceptedWrite{
				Table:          "operation_history",
				ID:             found.ID,
				Op:             mutation.Op,
				IdempotencyKey: found.IdempotencyKey,
			}, nil
		}
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, mapRepositoryError(err)
		}
	}
	row.CreatedAt = defaultTime(row.CreatedAt, s.clock.Now().UTC())
	if strings.TrimSpace(row.CreatedByDeviceID) == "" {
		row.CreatedByDeviceID = deviceID
	}
	if err := tx.InsertOperationHistory(ctx, row); err != nil {
		return nil, mapRepositoryError(err)
	}
	return &AcceptedWrite{
		Table:          "operation_history",
		ID:             row.ID,
		Op:             mutation.Op,
		IdempotencyKey: row.IdempotencyKey,
	}, nil
}

func (s *SyncUploadService) acceptUserSettings(ctx context.Context, tx syncTx, userID, deviceID string, mutation SyncMutation) (*AcceptedWrite, error) {
	row, err := decodeUserSettingsRow(mutation.Row)
	if err != nil {
		return nil, err
	}
	if err := normalizeOwnedRow(userID, &row.UserID); err != nil {
		return nil, err
	}
	if err := domain.ValidateSettings(domain.Settings{
		Language:                  domain.LanguageCode(row.Language),
		Locale:                    row.Locale,
		WeekStartDay:              domain.WeekStartDay(row.WeekStartDay),
		TimeZone:                  row.TimeZone,
		DateDisplayFormat:         row.DateDisplayFormat,
		TimeDisplayFormat:         domain.TimeDisplayFormat(row.TimeDisplayFormat),
		DefaultTimeZoneMode:       domain.DefaultTimeZoneMode(row.DefaultTimeZoneMode),
		TodayPrimaryLookaheadDays: row.TodayPrimaryLookaheadDays,
		DeadlineAwarenessDays:     row.DeadlineAwarenessDays,
	}); err != nil {
		return nil, err
	}
	now := s.clock.Now().UTC()
	previous, err := tx.GetUserSettingsForUpdate(ctx, userID)
	exists := err == nil
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, mapRepositoryError(err)
	}
	if exists {
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
	if err := tx.UpsertUserSettings(ctx, row); err != nil {
		return nil, mapRepositoryError(err)
	}
	return &AcceptedWrite{
		Table:    "user_settings",
		ID:       row.UserID,
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

type operationRow = repository.OperationHistoryRecord

func companionOperations(mutations []SyncMutation) (map[string]operationRow, []operationRow, error) {
	out := map[string]operationRow{}
	all := []operationRow{}
	for _, mutation := range mutations {
		if mutation.Table != "operation_history" || mutation.Op != "insert" {
			continue
		}
		row, err := decodeOperationRow(mutation.Row)
		if err != nil {
			return nil, nil, err
		}
		all = append(all, row)
		if row.ItemID != nil {
			out[*row.ItemID] = row
		}
	}
	return out, all, nil
}

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
