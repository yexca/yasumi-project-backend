package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestBackendConstantsCoverSharedContract(t *testing.T) {
	root := contractRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "shared_constants", "01-shared-constants-reference.md"))
	if err != nil {
		t.Fatalf("read shared constants: %v", err)
	}
	text := string(raw)

	assertSharedValues(t, text, "ITEM_TYPES", func(value string) bool {
		return IsValidItemType(ItemType(value))
	})
	assertSharedValues(t, text, "BUSINESS_STATUSES", func(value string) bool {
		return IsValidBusinessStatus(BusinessStatus(value))
	})
	assertSharedValues(t, text, "DEADLINE_TIME_ZONE_MODES", func(value string) bool {
		return IsValidDeadlineTimeZoneMode(DeadlineTimeZoneMode(value))
	})
	assertSharedValues(t, text, "OPERATION_EVENT_TYPES", func(value string) bool {
		return IsValidOperationEventType(OperationEventType(value))
	})
	assertSharedValues(t, text, "ERROR_CODES", func(value string) bool {
		return IsValidErrorCode(ErrorCode(value))
	})
	assertSharedValues(t, text, "ERROR_FIELD_KEYS", func(value string) bool {
		return IsValidFieldKey(FieldKey(value))
	})
}

func TestContractFixtureItemShapesMatchBackendValidation(t *testing.T) {
	root := contractRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "contracts", "fixtures", "data", "item-shapes-and-deadlines.json"))
	if err != nil {
		t.Fatalf("read item fixture: %v", err)
	}
	var fixture struct {
		Scenarios []struct {
			ID    string `json:"id"`
			Input struct {
				Rows []fixtureItem `json:"rows"`
			} `json:"input"`
			Expected struct {
				AcceptedRowIDs []string `json:"acceptedRowIds"`
				RejectedRowIDs []string `json:"rejectedRowIds"`
				ErrorsByRowID  map[string]struct {
					Code   string            `json:"code"`
					Fields map[string]string `json:"fields"`
				} `json:"errorsByRowId"`
			} `json:"expected"`
		} `json:"scenarios"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	for _, scenario := range fixture.Scenarios {
		accepted := stringSet(scenario.Expected.AcceptedRowIDs)
		rejected := stringSet(scenario.Expected.RejectedRowIDs)
		for _, row := range scenario.Input.Rows {
			err := ValidateItem(row.domainItem())
			switch {
			case accepted[row.ID] && err != nil:
				t.Fatalf("%s valid fixture %s failed validation: %v", scenario.ID, row.ID, err)
			case rejected[row.ID] && err == nil:
				t.Fatalf("%s invalid fixture %s passed validation", scenario.ID, row.ID)
			case rejected[row.ID]:
				expected := scenario.Expected.ErrorsByRowID[row.ID]
				if string(err.Code) != expected.Code {
					t.Fatalf("%s invalid fixture %s code = %s, want %s", scenario.ID, row.ID, err.Code, expected.Code)
				}
				for field, reason := range expected.Fields {
					if got := err.Fields[FieldKey(field)]; got != reason {
						t.Fatalf("%s invalid fixture %s field %s = %s, want %s", scenario.ID, row.ID, field, got, reason)
					}
				}
			}
		}
	}
}

func TestPowerSyncRulesScopeEveryStreamByAuthenticatedUser(t *testing.T) {
	root := projectRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "env", "powersync", "sync-config.yaml"))
	if err != nil {
		t.Fatalf("read sync config: %v", err)
	}
	text := string(raw)
	for _, table := range []string{"areas", "recurring_task_templates", "items", "operation_history", "user_settings"} {
		want := "FROM " + table + " WHERE user_id = auth.user_id()"
		if !regexp.MustCompile(regexp.QuoteMeta(want)).MatchString(text) {
			t.Fatalf("sync config missing scoped query %q", want)
		}
	}
}

func TestNoCalendarIntegrationInActiveBackendSurface(t *testing.T) {
	root := projectRoot(t)
	for _, dir := range []string{"cmd", "internal", "env"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}
			if filepath.Ext(path) == ".go" && regexp.MustCompile(`_test\.go$`).MatchString(path) {
				return nil
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if regexp.MustCompile(`(?i)calendar|oauth|provider_token`).Match(raw) {
				t.Fatalf("future-scope calendar integration found in %s", path)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s: %v", dir, err)
		}
	}
}

type fixtureItem struct {
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
	HiddenReason          *string `json:"hidden_reason"`
}

func (i fixtureItem) domainItem() Item {
	return Item{
		Title:                 i.Title,
		ItemType:              ItemType(i.ItemType),
		Status:                BusinessStatus(i.Status),
		ScheduledDate:         derefString(i.ScheduledDate),
		ScheduledTimeZoneMode: ScheduledTimeZoneMode(derefString(i.ScheduledTimeZoneMode)),
		DeadlineDate:          derefString(i.DeadlineDate),
		DeadlineTime:          derefString(i.DeadlineTime),
		DeadlineAt:            derefString(i.DeadlineAt),
		DeadlineTimeZoneMode:  DeadlineTimeZoneMode(derefString(i.DeadlineTimeZoneMode)),
		HiddenReason:          HiddenReason(derefString(i.HiddenReason)),
	}
}

func assertSharedValues(t *testing.T, text, name string, valid func(string) bool) {
	t.Helper()
	values := sharedArrayValues(t, text, name)
	for _, value := range values {
		if !valid(value) {
			t.Fatalf("%s contains %q but backend does not accept it", name, value)
		}
	}
}

func sharedArrayValues(t *testing.T, text, name string) []string {
	t.Helper()
	re := regexp.MustCompile(`(?s)export const ` + regexp.QuoteMeta(name) + ` = \[(.*?)\] as const`)
	match := re.FindStringSubmatch(text)
	if len(match) != 2 {
		t.Fatalf("shared constants array %s not found", name)
	}
	valueRe := regexp.MustCompile(`"([^"]+)"`)
	matches := valueRe.FindAllStringSubmatch(match[1], -1)
	values := make([]string, 0, len(matches))
	for _, match := range matches {
		values = append(values, match[1])
	}
	return values
}

func projectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("project root not found")
		}
		dir = parent
	}
}

func contractRoot(t *testing.T) string {
	t.Helper()
	if dir := os.Getenv("YASUMI_CONTRACTS_DIR"); dir != "" {
		if _, err := os.Stat(dir); err != nil {
			t.Fatalf("configured YASUMI_CONTRACTS_DIR is unavailable: %v", err)
		}
		return dir
	}
	dir := filepath.Join(projectRoot(t), "..", "dev_documents")
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("contract compatibility tests require dev_documents or YASUMI_CONTRACTS_DIR: %v", err)
	}
	return dir
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func stringSet(values []string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, value := range values {
		out[value] = true
	}
	return out
}
