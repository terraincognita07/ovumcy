package api

import (
	"reflect"
	"testing"

	"github.com/terraincognita07/lume/internal/models"
)

func TestOwnerBaselineCycleLength(t *testing.T) {
	tests := []struct {
		name string
		user *models.User
		want int
	}{
		{name: "nil user", user: nil, want: 0},
		{name: "partner", user: &models.User{Role: models.RolePartner, CycleLength: 29}, want: 0},
		{name: "owner invalid cycle", user: &models.User{Role: models.RoleOwner, CycleLength: 120}, want: 0},
		{name: "owner valid cycle", user: &models.User{Role: models.RoleOwner, CycleLength: 28}, want: 28},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := ownerBaselineCycleLength(testCase.user)
			if got != testCase.want {
				t.Fatalf("expected %d, got %d", testCase.want, got)
			}
		})
	}
}

func TestBuildCycleTrendLabels(t *testing.T) {
	if got := buildCycleTrendLabels(map[string]string{}, 0); len(got) != 0 {
		t.Fatalf("expected empty labels for zero points, got %#v", got)
	}

	defaultLabels := buildCycleTrendLabels(map[string]string{}, 3)
	expectedDefault := []string{"Cycle 1", "Cycle 2", "Cycle 3"}
	if !reflect.DeepEqual(defaultLabels, expectedDefault) {
		t.Fatalf("expected default labels %#v, got %#v", expectedDefault, defaultLabels)
	}

	customMessages := map[string]string{"stats.cycle_label": "Цикл %d"}
	customLabels := buildCycleTrendLabels(customMessages, 2)
	expectedCustom := []string{"Цикл 1", "Цикл 2"}
	if !reflect.DeepEqual(customLabels, expectedCustom) {
		t.Fatalf("expected custom labels %#v, got %#v", expectedCustom, customLabels)
	}
}

func TestLocalizeSymptomFrequencySummaries(t *testing.T) {
	counts := []SymptomCount{
		{Name: "A", Count: 1, TotalDays: 1},
		{Name: "B", Count: 2, TotalDays: 4},
	}

	localizeSymptomFrequencySummaries("en", counts)
	if counts[0].FrequencySummary != "1 time (in 1 day)" {
		t.Fatalf("unexpected localized summary for first item: %q", counts[0].FrequencySummary)
	}
	if counts[1].FrequencySummary != "2 times (in 4 days)" {
		t.Fatalf("unexpected localized summary for second item: %q", counts[1].FrequencySummary)
	}
}
