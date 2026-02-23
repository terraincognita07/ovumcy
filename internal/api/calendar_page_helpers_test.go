package api

import (
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestBuildCalendarViewDataOwnerPayload(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "calendar-view-data-owner@example.com")

	now := time.Date(2026, time.February, 21, 0, 0, 0, 0, time.UTC)
	monthStart := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)

	messages := map[string]string{
		"meta.title.calendar": "Calendar",
	}
	data, errorMessage, err := handler.buildCalendarViewData(&user, "en", messages, now, monthStart, "2026-02-17")
	if err != nil {
		t.Fatalf("buildCalendarViewData returned error: %v", err)
	}
	if errorMessage != "" {
		t.Fatalf("expected empty error message, got %q", errorMessage)
	}

	if got, ok := data["MonthValue"].(string); !ok || got != "2026-02" {
		t.Fatalf("expected MonthValue=2026-02, got %#v", data["MonthValue"])
	}
	if got, ok := data["PrevMonth"].(string); !ok || got != "2026-01" {
		t.Fatalf("expected PrevMonth=2026-01, got %#v", data["PrevMonth"])
	}
	if got, ok := data["NextMonth"].(string); !ok || got != "2026-03" {
		t.Fatalf("expected NextMonth=2026-03, got %#v", data["NextMonth"])
	}
	if got, ok := data["SelectedDate"].(string); !ok || got != "2026-02-17" {
		t.Fatalf("expected SelectedDate=2026-02-17, got %#v", data["SelectedDate"])
	}
	if got, ok := data["Today"].(string); !ok || got != "2026-02-21" {
		t.Fatalf("expected Today=2026-02-21, got %#v", data["Today"])
	}
	if isOwner, ok := data["IsOwner"].(bool); !ok || !isOwner {
		t.Fatalf("expected IsOwner=true, got %#v", data["IsOwner"])
	}

	days, ok := data["CalendarDays"].([]CalendarDay)
	if !ok {
		t.Fatalf("expected CalendarDays type []CalendarDay, got %T", data["CalendarDays"])
	}
	if len(days) == 0 {
		t.Fatal("expected non-empty calendar days payload")
	}
}

func TestBuildCalendarViewDataPartnerRoleFlag(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	partner := models.User{
		Email:               "calendar-view-data-partner@example.com",
		PasswordHash:        "test-hash",
		Role:                models.RolePartner,
		OnboardingCompleted: true,
		CycleLength:         28,
		PeriodLength:        5,
		CreatedAt:           time.Now().UTC(),
	}
	if err := database.Create(&partner).Error; err != nil {
		t.Fatalf("create partner user: %v", err)
	}

	now := time.Date(2026, time.February, 21, 0, 0, 0, 0, time.UTC)
	monthStart := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)

	data, errorMessage, err := handler.buildCalendarViewData(&partner, "en", map[string]string{}, now, monthStart, "")
	if err != nil {
		t.Fatalf("buildCalendarViewData returned error: %v", err)
	}
	if errorMessage != "" {
		t.Fatalf("expected empty error message, got %q", errorMessage)
	}
	if isOwner, ok := data["IsOwner"].(bool); !ok || isOwner {
		t.Fatalf("expected IsOwner=false, got %#v", data["IsOwner"])
	}
}
