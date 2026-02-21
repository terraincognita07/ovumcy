package api

import (
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestTrimTrailingCycleTrendLengths(t *testing.T) {
	t.Parallel()

	source := []int{1, 2, 3, 4, 5}
	if got := trimTrailingCycleTrendLengths(source, 10); !reflect.DeepEqual(got, source) {
		t.Fatalf("expected unchanged lengths, got %#v", got)
	}

	expected := []int{3, 4, 5}
	if got := trimTrailingCycleTrendLengths(source, 3); !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected last 3 lengths %#v, got %#v", expected, got)
	}
}

func TestBuildStatsChartDataIncludesBaselineWhenPositive(t *testing.T) {
	t.Parallel()

	messages := map[string]string{"stats.cycle_label": "Cycle %d"}
	lengths := []int{28, 30}

	withBaseline := buildStatsChartData(messages, lengths, 29)
	if _, ok := withBaseline["baseline"]; !ok {
		t.Fatal("expected baseline in chart payload when baseline > 0")
	}
	if got, ok := withBaseline["labels"].([]string); !ok || len(got) != 2 {
		t.Fatalf("expected 2 labels, got %#v", withBaseline["labels"])
	}

	withoutBaseline := buildStatsChartData(messages, lengths, 0)
	if _, ok := withoutBaseline["baseline"]; ok {
		t.Fatal("did not expect baseline in chart payload when baseline == 0")
	}
}

func TestBuildStatsChartDataTypeCompatibility(t *testing.T) {
	t.Parallel()

	payload := buildStatsChartData(map[string]string{}, []int{}, 0)
	if _, ok := any(payload).(fiber.Map); !ok {
		t.Fatalf("expected fiber.Map payload type, got %T", payload)
	}
}
