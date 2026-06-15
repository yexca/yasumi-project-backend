package service

import (
	"context"
	"errors"
	"strings"

	"github.com/yasumi/yasumi-project-backend/internal/domain"
	"github.com/yasumi/yasumi-project-backend/internal/repository"
)

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
