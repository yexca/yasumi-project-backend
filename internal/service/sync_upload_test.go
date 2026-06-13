package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yasumi/yasumi-project-backend/internal/domain"
	"github.com/yasumi/yasumi-project-backend/internal/repository"
)

const (
	serviceUserID = "00000000-0000-4000-8000-000000000001"
	otherUserID   = "00000000-0000-4000-8000-000000000002"
	serviceItemID = "10000000-0000-4000-8000-000000000001"
	serviceAreaID = "40000000-0000-4000-8000-000000000001"
	serviceTplID  = "20000000-0000-4000-8000-000000000001"
	serviceOpID   = "30000000-0000-4000-8000-000000000001"
	serviceDevice = "device-01"
)

func TestAcceptUploadNormalizesMissingUserIDAndAssignsMetadata(t *testing.T) {
	repo := newFakeSyncRepo()
	svc := NewSyncUploadService(repo.Repository(), fixedClock{})

	result, err := svc.AcceptUpload(context.Background(), serviceUserID, SyncUpload{
		ClientBatchID: "batch-01",
		DeviceID:      serviceDevice,
		Mutations: []SyncMutation{{
			Table: "items",
			Op:    "insert",
			Row:   rawJSON(`{"id":"` + serviceItemID + `","item_type":"inbox","title":"Capture","status":"active"}`),
		}},
	})
	if err != nil {
		t.Fatalf("AcceptUpload error = %v", err)
	}
	if len(result.Accepted) != 1 {
		t.Fatalf("accepted = %+v, want one item", result.Accepted)
	}
	item := repo.items[serviceItemID]
	if item.UserID != serviceUserID {
		t.Fatalf("item user_id = %q, want authenticated user", item.UserID)
	}
	if item.Revision != 1 {
		t.Fatalf("revision = %d, want 1", item.Revision)
	}
	if item.ServerUpdatedAt.IsZero() || item.UpdatedAt.IsZero() {
		t.Fatal("server metadata was not assigned")
	}
}

func TestAcceptUploadRejectsCrossUserRow(t *testing.T) {
	repo := newFakeSyncRepo()
	svc := NewSyncUploadService(repo.Repository(), fixedClock{})

	_, err := svc.AcceptUpload(context.Background(), serviceUserID, SyncUpload{
		DeviceID: serviceDevice,
		Mutations: []SyncMutation{{
			Table: "items",
			Op:    "insert",
			Row:   rawJSON(`{"id":"` + serviceItemID + `","user_id":"` + otherUserID + `","item_type":"inbox","title":"Capture","status":"active"}`),
		}},
	})
	assertDomainCode(t, err, domain.ErrorForbidden)
}

func TestAcceptUploadRejectsSemanticChangeWithoutIntent(t *testing.T) {
	repo := newFakeSyncRepo()
	repo.items[serviceItemID] = repository.ItemRecord{
		ID:                serviceItemID,
		UserID:            serviceUserID,
		ItemType:          "inbox",
		Title:             "Capture",
		Status:            "active",
		CreatedAt:         fixedNow.Add(-time.Hour),
		CreatedByDeviceID: serviceDevice,
		Revision:          2,
	}
	svc := NewSyncUploadService(repo.Repository(), fixedClock{})

	_, err := svc.AcceptUpload(context.Background(), serviceUserID, SyncUpload{
		DeviceID: serviceDevice,
		Mutations: []SyncMutation{{
			Table: "items",
			Op:    "update",
			Row:   rawJSON(`{"id":"` + serviceItemID + `","user_id":"` + serviceUserID + `","item_type":"inbox","title":"Capture","status":"completed"}`),
		}},
	})
	assertDomainCode(t, err, domain.ErrorInvalidTransition)
}

func TestAcceptUploadAcceptsSemanticChangeWithCompanionOperation(t *testing.T) {
	idempotencyKey := domain.ActionIdempotencyKey(serviceUserID, serviceDevice, "complete-01")
	repo := newFakeSyncRepo()
	repo.items[serviceItemID] = repository.ItemRecord{
		ID:                serviceItemID,
		UserID:            serviceUserID,
		ItemType:          "inbox",
		Title:             "Capture",
		Status:            "active",
		CreatedAt:         fixedNow.Add(-time.Hour),
		CreatedByDeviceID: serviceDevice,
		Revision:          2,
	}
	svc := NewSyncUploadService(repo.Repository(), fixedClock{})

	result, err := svc.AcceptUpload(context.Background(), serviceUserID, SyncUpload{
		DeviceID: serviceDevice,
		Mutations: []SyncMutation{
			{
				Table: "operation_history",
				Op:    "insert",
				Row: rawJSON(`{
					"id":"` + serviceOpID + `",
					"user_id":"` + serviceUserID + `",
					"item_id":"` + serviceItemID + `",
					"event_type":"completed",
					"idempotency_key":"` + idempotencyKey + `"
				}`),
			},
			{
				Table: "items",
				Op:    "update",
				Row:   rawJSON(`{"id":"` + serviceItemID + `","user_id":"` + serviceUserID + `","item_type":"inbox","title":"Capture","status":"completed"}`),
			},
		},
	})
	if err != nil {
		t.Fatalf("AcceptUpload error = %v", err)
	}
	if result.DuplicateOf != nil {
		t.Fatalf("duplicate = %+v, want nil", result.DuplicateOf)
	}
	if got := repo.items[serviceItemID].Status; got != "completed" {
		t.Fatalf("status = %q, want completed", got)
	}
	if _, ok := repo.operations[idempotencyKey]; !ok {
		t.Fatal("companion operation was not inserted")
	}
}

func TestAcceptUploadAcceptsAreaAndRecurringTemplateWrites(t *testing.T) {
	repo := newFakeSyncRepo()
	svc := NewSyncUploadService(repo.Repository(), fixedClock{})

	result, err := svc.AcceptUpload(context.Background(), serviceUserID, SyncUpload{
		DeviceID: serviceDevice,
		Mutations: []SyncMutation{
			{
				Table: "areas",
				Op:    "insert",
				Row:   rawJSON(`{"id":"` + serviceAreaID + `","name":"Home","sort_order":1}`),
			},
			{
				Table: "recurring_task_templates",
				Op:    "insert",
				Row: rawJSON(`{
					"id":"` + serviceTplID + `",
					"title":"Stretch",
					"frequency":"daily",
					"interval":1,
					"recurrence_basis":"scheduled_date",
					"start_date":"2026-06-13",
					"end_type":"never",
					"completed_count":0,
					"next_sequence":1,
					"status":"active"
				}`),
			},
		},
	})
	if err != nil {
		t.Fatalf("AcceptUpload error = %v", err)
	}
	if len(result.Accepted) != 2 {
		t.Fatalf("accepted = %+v, want two writes", result.Accepted)
	}
	if repo.areas[serviceAreaID].UserID != serviceUserID {
		t.Fatalf("area user_id = %q, want %q", repo.areas[serviceAreaID].UserID, serviceUserID)
	}
	if repo.templates[serviceTplID].Revision != 1 {
		t.Fatalf("template revision = %d, want 1", repo.templates[serviceTplID].Revision)
	}
}

func TestAcceptUploadRejectsObservedRevisionMismatch(t *testing.T) {
	repo := newFakeSyncRepo()
	repo.items[serviceItemID] = repository.ItemRecord{
		ID:                serviceItemID,
		UserID:            serviceUserID,
		ItemType:          "inbox",
		Title:             "Capture",
		Status:            "active",
		CreatedAt:         fixedNow.Add(-time.Hour),
		CreatedByDeviceID: serviceDevice,
		Revision:          4,
	}
	svc := NewSyncUploadService(repo.Repository(), fixedClock{})
	observedRevision := int64(3)
	observedStatus := "active"

	_, err := svc.AcceptUpload(context.Background(), serviceUserID, SyncUpload{
		DeviceID: serviceDevice,
		Mutations: []SyncMutation{{
			Table: "items",
			Op:    "update",
			Row:   rawJSON(`{"id":"` + serviceItemID + `","user_id":"` + serviceUserID + `","item_type":"inbox","title":"Capture","status":"completed"}`),
			ClientObserved: ObservedState{
				Revision: &observedRevision,
				Status:   &observedStatus,
			},
		}},
	})
	assertDomainCode(t, err, domain.ErrorInvalidTransition)
}

func TestAcceptUploadAssignsSettingsRevisionFromServerState(t *testing.T) {
	repo := newFakeSyncRepo()
	repo.settings = map[string]repository.UserSettingsRecord{
		serviceUserID: {
			UserID:                    serviceUserID,
			Language:                  "ja",
			Locale:                    "ja-JP",
			WeekStartDay:              "monday",
			TimeZone:                  "Asia/Tokyo",
			DateDisplayFormat:         "yyyy-MM-dd",
			TimeDisplayFormat:         "24h",
			DefaultTimeZoneMode:       "floating",
			TodayPrimaryLookaheadDays: 3,
			DeadlineAwarenessDays:     14,
			CreatedAt:                 fixedNow.Add(-2 * time.Hour),
			CreatedByDeviceID:         serviceDevice,
			Revision:                  7,
		},
	}
	svc := NewSyncUploadService(repo.Repository(), fixedClock{})

	result, err := svc.AcceptUpload(context.Background(), serviceUserID, SyncUpload{
		DeviceID: serviceDevice,
		Mutations: []SyncMutation{{
			Table: "user_settings",
			Op:    "upsert",
			Row: rawJSON(`{
				"user_id":"` + serviceUserID + `",
				"language":"ja",
				"locale":"ja-JP",
				"week_start_day":"monday",
				"time_zone":"Asia/Tokyo",
				"date_display_format":"yyyy-MM-dd",
				"time_display_format":"24h",
				"default_time_zone_mode":"floating",
				"today_primary_lookahead_days":4,
				"deadline_awareness_days":14,
				"revision":1
			}`),
		}},
	})
	if err != nil {
		t.Fatalf("AcceptUpload error = %v", err)
	}
	if got := *result.Accepted[0].Revision; got != 8 {
		t.Fatalf("accepted revision = %d, want 8", got)
	}
	if got := repo.settings[serviceUserID].Revision; got != 8 {
		t.Fatalf("stored revision = %d, want 8", got)
	}
}

func TestAcceptUploadRejectsRestoreWithoutRestoredOperation(t *testing.T) {
	repo := newFakeSyncRepo()
	deletedAt := fixedNow.Add(-24 * time.Hour)
	archivedAt := fixedNow.Add(-48 * time.Hour)
	repo.items[serviceItemID] = repository.ItemRecord{
		ID:                   serviceItemID,
		UserID:               serviceUserID,
		ItemType:             "deadline_task",
		Title:                "Restore completed record",
		Status:               "completed",
		DeadlineDate:         stringPtr("2026-06-20"),
		DeadlineTimeZoneMode: stringPtr("date_only"),
		DeletedAt:            &deletedAt,
		ArchivedAt:           &archivedAt,
		CreatedAt:            fixedNow.Add(-72 * time.Hour),
		CreatedByDeviceID:    serviceDevice,
		Revision:             11,
	}
	svc := NewSyncUploadService(repo.Repository(), fixedClock{})
	revision := int64(11)
	status := "completed"
	deletedAtString := deletedAt.UTC().Format(time.RFC3339)
	archivedAtString := archivedAt.UTC().Format(time.RFC3339)

	_, err := svc.AcceptUpload(context.Background(), serviceUserID, SyncUpload{
		DeviceID: serviceDevice,
		Mutations: []SyncMutation{
			{
				Table: "operation_history",
				Op:    "insert",
				Row: rawJSON(`{
					"id":"` + serviceOpID + `",
					"user_id":"` + serviceUserID + `",
					"item_id":"` + serviceItemID + `",
					"event_type":"completed"
				}`),
			},
			{
				Table: "items",
				Op:    "update",
				Row:   rawJSON(`{"id":"` + serviceItemID + `","user_id":"` + serviceUserID + `","item_type":"deadline_task","title":"Restore completed record","status":"completed","deadline_date":"2026-06-20","deadline_time_zone_mode":"date_only","deleted_at":null,"archived_at":null}`),
				ClientObserved: ObservedState{
					Revision:   &revision,
					Status:     &status,
					DeletedAt:  &deletedAtString,
					ArchivedAt: &archivedAtString,
				},
			},
		},
	})
	assertDomainCode(t, err, domain.ErrorInvalidTransition)
}

func TestAcceptUploadAcceptsPostponedActivationWhenDue(t *testing.T) {
	idempotencyKey := domain.PostponedActivationIdempotencyKey(serviceUserID, serviceItemID, "2026-06-13")
	repo := newFakeSyncRepo()
	scheduledDate := "2026-06-13"
	repo.items[serviceItemID] = repository.ItemRecord{
		ID:                serviceItemID,
		UserID:            serviceUserID,
		ItemType:          "date_task",
		Title:             "Return library books",
		Status:            "postponed",
		ScheduledDate:     &scheduledDate,
		CreatedAt:         fixedNow.Add(-time.Hour),
		CreatedByDeviceID: serviceDevice,
		Revision:          7,
	}
	svc := NewSyncUploadService(repo.Repository(), fixedClock{})
	revision := int64(7)
	status := "postponed"

	result, err := svc.AcceptUpload(context.Background(), serviceUserID, SyncUpload{
		DeviceID: serviceDevice,
		Mutations: []SyncMutation{
			{
				Table: "operation_history",
				Op:    "insert",
				Row: rawJSON(`{
					"id":"` + serviceOpID + `",
					"user_id":"` + serviceUserID + `",
					"item_id":"` + serviceItemID + `",
					"event_type":"activated_from_postponed",
					"idempotency_key":"` + idempotencyKey + `"
				}`),
			},
			{
				Table: "items",
				Op:    "update",
				Row:   rawJSON(`{"id":"` + serviceItemID + `","user_id":"` + serviceUserID + `","item_type":"date_task","title":"Return library books","status":"active","scheduled_date":"2026-06-13"}`),
				ClientObserved: ObservedState{
					Revision: &revision,
					Status:   &status,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("AcceptUpload error = %v", err)
	}
	if len(result.Accepted) != 2 {
		t.Fatalf("accepted = %+v, want two writes", result.Accepted)
	}
	if got := repo.items[serviceItemID].Status; got != "active" {
		t.Fatalf("status = %q, want active", got)
	}
}

func TestAcceptUploadRejectsPostponedActivationWhenDeadlineTaskIsNotDue(t *testing.T) {
	idempotencyKey := domain.PostponedActivationIdempotencyKey(serviceUserID, serviceItemID, "2026-06-13")
	repo := newFakeSyncRepo()
	deadlineDate := "2026-06-14"
	mode := "date_only"
	repo.items[serviceItemID] = repository.ItemRecord{
		ID:                   serviceItemID,
		UserID:               serviceUserID,
		ItemType:             "deadline_task",
		Title:                "Submit tax document",
		Status:               "postponed",
		DeadlineDate:         &deadlineDate,
		DeadlineTimeZoneMode: &mode,
		CreatedAt:            fixedNow.Add(-time.Hour),
		CreatedByDeviceID:    serviceDevice,
		Revision:             3,
	}
	svc := NewSyncUploadService(repo.Repository(), fixedClock{})
	revision := int64(3)
	status := "postponed"

	_, err := svc.AcceptUpload(context.Background(), serviceUserID, SyncUpload{
		DeviceID: serviceDevice,
		Mutations: []SyncMutation{
			{
				Table: "operation_history",
				Op:    "insert",
				Row: rawJSON(`{
					"id":"` + serviceOpID + `",
					"user_id":"` + serviceUserID + `",
					"item_id":"` + serviceItemID + `",
					"event_type":"activated_from_postponed",
					"idempotency_key":"` + idempotencyKey + `"
				}`),
			},
			{
				Table: "items",
				Op:    "update",
				Row:   rawJSON(`{"id":"` + serviceItemID + `","user_id":"` + serviceUserID + `","item_type":"deadline_task","title":"Submit tax document","status":"active","deadline_date":"2026-06-14","deadline_time_zone_mode":"date_only"}`),
				ClientObserved: ObservedState{
					Revision: &revision,
					Status:   &status,
				},
			},
		},
	})
	assertDomainCode(t, err, domain.ErrorInvalidTransition)
}

func TestAcceptUploadDuplicateIdempotencyKeyReturnsExistingResult(t *testing.T) {
	idempotencyKey := domain.ActionIdempotencyKey(serviceUserID, serviceDevice, "complete-01")
	repo := newFakeSyncRepo()
	repo.operations[idempotencyKey] = repository.OperationHistoryRecord{
		ID:             serviceOpID,
		UserID:         serviceUserID,
		EventType:      "completed",
		IdempotencyKey: &idempotencyKey,
	}
	svc := NewSyncUploadService(repo.Repository(), fixedClock{})

	result, err := svc.AcceptUpload(context.Background(), serviceUserID, SyncUpload{
		ClientBatchID: "retry-01",
		DeviceID:      serviceDevice,
		Mutations: []SyncMutation{{
			Table: "operation_history",
			Op:    "insert",
			Row: rawJSON(`{
				"id":"30000000-0000-4000-8000-000000000099",
				"user_id":"` + serviceUserID + `",
				"event_type":"completed",
				"idempotency_key":"` + idempotencyKey + `"
			}`),
		}},
	})
	if err != nil {
		t.Fatalf("AcceptUpload error = %v", err)
	}
	if result.DuplicateOf == nil || result.DuplicateOf.ID != serviceOpID {
		t.Fatalf("duplicate = %+v, want existing operation", result.DuplicateOf)
	}
}

type fixedClock struct{}

var fixedNow = time.Date(2026, 6, 13, 9, 30, 0, 0, time.UTC)

func (fixedClock) Now() time.Time {
	return fixedNow
}

func rawJSON(value string) []byte {
	return []byte(strings.ReplaceAll(value, "\t", ""))
}

func stringPtr(value string) *string {
	return &value
}

func assertDomainCode(t *testing.T, err error, code domain.ErrorCode) {
	t.Helper()
	var domainErr *domain.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("error = %v, want domain error", err)
	}
	if domainErr.Code != code {
		t.Fatalf("code = %q, want %q; err=%v", domainErr.Code, code, err)
	}
}

type fakeSyncRepo struct {
	areas      map[string]repository.AreaRecord
	templates  map[string]repository.RecurringTemplateRecord
	items      map[string]repository.ItemRecord
	operations map[string]repository.OperationHistoryRecord
	settings   map[string]repository.UserSettingsRecord
}

func newFakeSyncRepo() *fakeSyncRepo {
	return &fakeSyncRepo{
		areas:      map[string]repository.AreaRecord{},
		templates:  map[string]repository.RecurringTemplateRecord{},
		items:      map[string]repository.ItemRecord{},
		operations: map[string]repository.OperationHistoryRecord{},
		settings:   map[string]repository.UserSettingsRecord{},
	}
}

func (r *fakeSyncRepo) Repository() *fakeSyncRepo {
	return r
}

func (r *fakeSyncRepo) InTx(ctx context.Context, fn func(context.Context, syncTx) error) error {
	return fn(ctx, fakeSyncTx{repo: r})
}

type fakeSyncTx struct {
	repo *fakeSyncRepo
}

func (tx fakeSyncTx) GetAreaForUpdateByUser(_ context.Context, userID, areaID string) (repository.AreaRecord, error) {
	row, ok := tx.repo.areas[areaID]
	if !ok || row.UserID != userID {
		return repository.AreaRecord{}, repository.ErrNotFound
	}
	return row, nil
}

func (tx fakeSyncTx) UpsertArea(_ context.Context, row repository.AreaRecord) error {
	if existing, ok := tx.repo.areas[row.ID]; ok && existing.UserID != row.UserID {
		return repository.ErrNotFound
	}
	tx.repo.areas[row.ID] = row
	return nil
}

func (tx fakeSyncTx) GetRecurringTemplateForUpdateByUser(_ context.Context, userID, templateID string) (repository.RecurringTemplateRecord, error) {
	row, ok := tx.repo.templates[templateID]
	if !ok || row.UserID != userID {
		return repository.RecurringTemplateRecord{}, repository.ErrNotFound
	}
	return row, nil
}

func (tx fakeSyncTx) UpsertRecurringTemplate(_ context.Context, row repository.RecurringTemplateRecord) error {
	if existing, ok := tx.repo.templates[row.ID]; ok && existing.UserID != row.UserID {
		return repository.ErrNotFound
	}
	tx.repo.templates[row.ID] = row
	return nil
}

func (tx fakeSyncTx) GetItemForUpdateByUser(_ context.Context, userID, itemID string) (repository.ItemRecord, error) {
	row, ok := tx.repo.items[itemID]
	if !ok || row.UserID != userID {
		return repository.ItemRecord{}, repository.ErrNotFound
	}
	return row, nil
}

func (tx fakeSyncTx) UpsertItem(_ context.Context, row repository.ItemRecord) error {
	if existing, ok := tx.repo.items[row.ID]; ok && existing.UserID != row.UserID {
		return repository.ErrNotFound
	}
	tx.repo.items[row.ID] = row
	return nil
}

func (tx fakeSyncTx) InsertOperationHistory(_ context.Context, row repository.OperationHistoryRecord) error {
	if row.IdempotencyKey != nil {
		if _, ok := tx.repo.operations[*row.IdempotencyKey]; ok {
			return repository.ErrNotFound
		}
		tx.repo.operations[*row.IdempotencyKey] = row
		return nil
	}
	tx.repo.operations[row.ID] = row
	return nil
}

func (tx fakeSyncTx) FindOperationByIdempotencyKey(_ context.Context, userID, idempotencyKey string) (repository.OperationHistoryRecord, error) {
	row, ok := tx.repo.operations[idempotencyKey]
	if !ok || row.UserID != userID {
		return repository.OperationHistoryRecord{}, repository.ErrNotFound
	}
	return row, nil
}

func (tx fakeSyncTx) GetUserSettingsForUpdate(_ context.Context, userID string) (repository.UserSettingsRecord, error) {
	row, ok := tx.repo.settings[userID]
	if !ok {
		return repository.UserSettingsRecord{}, repository.ErrNotFound
	}
	return row, nil
}

func (tx fakeSyncTx) UpsertUserSettings(_ context.Context, row repository.UserSettingsRecord) error {
	tx.repo.settings[row.UserID] = row
	return nil
}
