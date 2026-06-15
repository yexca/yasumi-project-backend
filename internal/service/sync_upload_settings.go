package service

import (
	"context"
	"errors"
	"strings"

	"github.com/yasumi/yasumi-project-backend/internal/domain"
	"github.com/yasumi/yasumi-project-backend/internal/repository"
)

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
		WeatherCity:               row.WeatherCity,
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
