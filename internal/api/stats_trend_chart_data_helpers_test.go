package api

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

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
