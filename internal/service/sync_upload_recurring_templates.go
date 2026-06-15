package service

import (
	"context"
	"errors"
	"strings"

	"github.com/yasumi/yasumi-project-backend/internal/domain"
	"github.com/yasumi/yasumi-project-backend/internal/repository"
)

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
