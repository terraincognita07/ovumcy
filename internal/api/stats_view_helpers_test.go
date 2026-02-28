package api

import (
	"reflect"
	"testing"
)

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
