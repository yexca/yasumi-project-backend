package service

import (
	"context"
	"errors"
	"strings"

	"github.com/yasumi/yasumi-project-backend/internal/domain"
	"github.com/yasumi/yasumi-project-backend/internal/repository"
)

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
