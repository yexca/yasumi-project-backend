package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDomainFixtureStatusTransitions(t *testing.T) {
	var fixture struct {
		Scenarios []struct {
			ID    string `json:"id"`
			Input struct {
				Transitions [][]string `json:"transitions"`
			} `json:"input"`
			Expected struct {
				AllAccepted bool `json:"allAccepted"`
				AllRejected bool `json:"allRejected"`
			} `json:"expected"`
		} `json:"scenarios"`
	}
	loadFixture(t, fixturePath("domain", "status-transitions.json"), &fixture)

	for _, scenario := range fixture.Scenarios {
		switch scenario.ID {
		case "status-allowed-transitions", "status-invalid-transitions":
			t.Run(scenario.ID, func(t *testing.T) {
				for _, pair := range scenario.Input.Transitions {
					err := ValidateStatusTransition(BusinessStatus(pair[0]), BusinessStatus(pair[1]))
					if scenario.Expected.AllAccepted && err != nil {
						t.Fatalf("transition %v error = %v, want accepted", pair, err)
					}
					if scenario.Expected.AllRejected {
						if err == nil {
							t.Fatalf("transition %v error = nil, want rejected", pair)
						}
						if err.Code != ErrorInvalidTransition {
							t.Fatalf("transition %v code = %q, want %q", pair, err.Code, ErrorInvalidTransition)
						}
					}
				}
			})
		}
	}
}

func TestDomainFixtureMetadataPrecedence(t *testing.T) {
	var fixture struct {
		Scenarios []struct {
			ID    string `json:"id"`
			Input struct {
				Rows []struct {
					ID           string  `json:"id"`
					Status       string  `json:"status"`
					DeletedAt    *string `json:"deleted_at"`
					ArchivedAt   *string `json:"archived_at"`
					HiddenReason *string `json:"hidden_reason"`
				} `json:"rows"`
			} `json:"input"`
			Expected struct {
				VisibilityByID map[string]string `json:"visibilityById"`
			} `json:"expected"`
		} `json:"scenarios"`
	}
	loadFixture(t, fixturePath("domain", "status-transitions.json"), &fixture)

	for _, scenario := range fixture.Scenarios {
		if scenario.ID != "metadata-precedence" {
			continue
		}
		for _, row := range scenario.Input.Rows {
			got := VisibilityForItem(Item{
				Status:       BusinessStatus(row.Status),
				DeletedAt:    derefFixture(row.DeletedAt),
				ArchivedAt:   derefFixture(row.ArchivedAt),
				HiddenReason: HiddenReason(derefFixture(row.HiddenReason)),
			})
			if string(got) != scenario.Expected.VisibilityByID[row.ID] {
				t.Fatalf("row %s visibility = %q, want %q", row.ID, got, scenario.Expected.VisibilityByID[row.ID])
			}
		}
	}
}

func TestDomainFixtureItemShapesAndDeadlines(t *testing.T) {
	var fixture struct {
		Scenarios []struct {
			ID    string `json:"id"`
			Input struct {
				Rows []struct {
					ID                    string  `json:"id"`
					ItemType              string  `json:"item_type"`
					Title                 string  `json:"title"`
					Status                string  `json:"status"`
					ScheduledDate         *string `json:"scheduled_date"`
					ScheduledTimeZoneMode *string `json:"scheduled_time_zone_mode"`
					DeadlineDate          *string `json:"deadline_date"`
					DeadlineTime          *string `json:"deadline_time"`
					DeadlineAt            *string `json:"deadline_at"`
					DeadlineTimeZoneMode  *string `json:"deadline_time_zone_mode"`
				} `json:"rows"`
			} `json:"input"`
			Expected struct {
				ErrorsByRowID map[string]struct {
					Code   string            `json:"code"`
					Fields map[string]string `json:"fields"`
				} `json:"errorsByRowId"`
			} `json:"expected"`
		} `json:"scenarios"`
	}
	loadFixture(t, fixturePath("data", "item-shapes-and-deadlines.json"), &fixture)

	for _, scenario := range fixture.Scenarios {
		t.Run(scenario.ID, func(t *testing.T) {
			for _, row := range scenario.Input.Rows {
				err := ValidateItem(Item{
					Title:                 row.Title,
					ItemType:              ItemType(row.ItemType),
					Status:                BusinessStatus(row.Status),
					ScheduledDate:         derefFixture(row.ScheduledDate),
					ScheduledTimeZoneMode: ScheduledTimeZoneMode(derefFixture(row.ScheduledTimeZoneMode)),
					DeadlineDate:          derefFixture(row.DeadlineDate),
					DeadlineTime:          derefFixture(row.DeadlineTime),
					DeadlineAt:            derefFixture(row.DeadlineAt),
					DeadlineTimeZoneMode:  DeadlineTimeZoneMode(derefFixture(row.DeadlineTimeZoneMode)),
				})

				expectedErr, shouldReject := scenario.Expected.ErrorsByRowID[row.ID]
				if shouldReject {
					if err == nil {
						t.Fatalf("row %s error = nil, want rejection", row.ID)
					}
					if string(err.Code) != expectedErr.Code {
						t.Fatalf("row %s code = %q, want %q", row.ID, err.Code, expectedErr.Code)
					}
					for field, reason := range expectedErr.Fields {
						if got := err.Fields[FieldKey(field)]; got != reason {
							t.Fatalf("row %s field %s = %q, want %q", row.ID, field, got, reason)
						}
					}
					continue
				}

				if err != nil {
					t.Fatalf("row %s error = %v, want accepted", row.ID, err)
				}
			}
		})
	}
}

func loadFixture(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("fixture not available in current workspace: %s", path)
		}
		t.Fatalf("read fixture %s: %v", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("unmarshal fixture %s: %v", path, err)
	}
}

func fixturePath(parts ...string) string {
	_, current, _, _ := runtime.Caller(0)
	base := filepath.Join(filepath.Dir(current), "..", "..", "..", "dev_documents", "contracts", "fixtures")
	all := append([]string{base}, parts...)
	return filepath.Join(all...)
}

func derefFixture(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
