package api

import (
	"testing"
	"time"
)

func TestLocalizedDashboardDate_Russian(t *testing.T) {
	value := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)

	got := localizedDashboardDate("ru", value)
	want := "18 февраля 2026, среда"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestLocalizedDashboardDate_English(t *testing.T) {
	value := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)

	got := localizedDashboardDate("en", value)
	want := "February 18, 2026, Wednesday"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestLocalizedMonthYear(t *testing.T) {
	value := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)

	if got := localizedMonthYear("ru", value); got != "Февраль 2026" {
		t.Fatalf("expected russian month-year, got %q", got)
	}
	if got := localizedMonthYear("en", value); got != "February 2026" {
		t.Fatalf("expected english month-year, got %q", got)
	}
	if got := localizedMonthYear("de", value); got != "February 2026" {
		t.Fatalf("expected fallback month-year, got %q", got)
	}
}

func TestLocalizedDateLabel(t *testing.T) {
	value := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)

	if got := localizedDateLabel("ru", value); got != "Ср, Фев 18" {
		t.Fatalf("expected russian date label, got %q", got)
	}
	if got := localizedDateLabel("en", value); got != "Wed, Feb 18" {
		t.Fatalf("expected english date label, got %q", got)
	}
	if got := localizedDateLabel("de", value); got != "Wed, Feb 18" {
		t.Fatalf("expected fallback date label, got %q", got)
	}
}

func TestAuthErrorTranslationKey_NormalizesInput(t *testing.T) {
	got := authErrorTranslationKey("  TOO MANY LOGIN ATTEMPTS ")
	want := "auth.error.too_many_login_attempts"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestSettingsStatusTranslationKey(t *testing.T) {
	if got := settingsStatusTranslationKey("  CYCLE_UPDATED "); got != "settings.success.cycle_updated" {
		t.Fatalf("expected cycle_updated key, got %q", got)
	}
	if got := settingsStatusTranslationKey("unknown"); got != "" {
		t.Fatalf("expected empty key for unknown status, got %q", got)
	}
}

func TestLocalizedSymptomFrequencySummary_EnglishPluralization(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		days     int
		expected string
	}{
		{name: "singular count and day", count: 1, days: 1, expected: "1 time (in 1 day)"},
		{name: "plural count and day", count: 2, days: 4, expected: "2 times (in 4 days)"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := localizedSymptomFrequencySummary("en", testCase.count, testCase.days)
			if got != testCase.expected {
				t.Fatalf("expected %q, got %q", testCase.expected, got)
			}
		})
	}
}

func TestLocalizedSymptomFrequencySummary_RussianPluralization(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		days     int
		expected string
	}{
		{name: "one form", count: 1, days: 1, expected: "1 раз (за 1 день)"},
		{name: "few form", count: 2, days: 4, expected: "2 раза (за 4 дня)"},
		{name: "many form", count: 5, days: 7, expected: "5 раз (за 7 дней)"},
		{name: "teens form", count: 11, days: 12, expected: "11 раз (за 12 дней)"},
		{name: "mixed form", count: 21, days: 22, expected: "21 раз (за 22 дня)"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := localizedSymptomFrequencySummary("ru", testCase.count, testCase.days)
			if got != testCase.expected {
				t.Fatalf("expected %q, got %q", testCase.expected, got)
			}
		})
	}
}
