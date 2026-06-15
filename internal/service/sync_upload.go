package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

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
