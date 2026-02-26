package services

import (
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestDayHasData(t *testing.T) {
	tests := []struct {
		name  string
		entry models.DailyLog
		want  bool
	}{
		{
			name:  "period day",
			entry: models.DailyLog{IsPeriod: true},
			want:  true,
		},
		{
			name:  "symptoms present",
			entry: models.DailyLog{SymptomIDs: []uint{1}},
			want:  true,
		},
		{
			name:  "notes present",
			entry: models.DailyLog{Notes: "note"},
			want:  true,
		},
		{
			name:  "flow present",
			entry: models.DailyLog{Flow: models.FlowLight},
			want:  true,
		},
		{
			name:  "empty entry",
			entry: models.DailyLog{Flow: models.FlowNone},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DayHasData(tt.entry); got != tt.want {
				t.Fatalf("DayHasData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDayRangeNormalizesToLocationMidnight(t *testing.T) {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	raw := time.Date(2026, 2, 1, 19, 35, 10, 0, time.UTC)
	start, end := DayRange(raw, location)

	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Fatalf("expected midnight start, got %s", start.Format(time.RFC3339))
	}
	if !end.Equal(start.AddDate(0, 0, 1)) {
		t.Fatalf("expected next day end, got %s", end.Format(time.RFC3339))
	}
}
