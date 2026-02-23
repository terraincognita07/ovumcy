package api

import (
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestBuildStatsPageDataOwnerBaselineAndFlags(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "stats-page-owner@example.com")
	user.Role = models.RoleOwner

	now := time.Date(2026, time.February, 21, 0, 0, 0, 0, time.UTC)
	messages := map[string]string{"meta.title.stats": "Stats"}

	data, errorMessage, err := handler.buildStatsPageData(&user, "en", messages, now)
	if err != nil {
		t.Fatalf("buildStatsPageData returned error: %v", err)
	}
	if errorMessage != "" {
		t.Fatalf("expected empty error message, got %q", errorMessage)
	}
	if baseline, ok := data["ChartBaseline"].(int); !ok || baseline != 28 {
		t.Fatalf("expected ChartBaseline=28, got %#v", data["ChartBaseline"])
	}
	if isOwner, ok := data["IsOwner"].(bool); !ok || !isOwner {
		t.Fatalf("expected IsOwner=true, got %#v", data["IsOwner"])
	}
	if _, ok := data["ChartData"].(fiber.Map); !ok {
		t.Fatalf("expected ChartData fiber.Map, got %T", data["ChartData"])
	}
}

func TestBuildStatsPageDataPartnerNoBaseline(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	partner := models.User{
		Email:               "stats-page-partner@example.com",
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
	data, errorMessage, err := handler.buildStatsPageData(&partner, "en", map[string]string{}, now)
	if err != nil {
		t.Fatalf("buildStatsPageData returned error: %v", err)
	}
	if errorMessage != "" {
		t.Fatalf("expected empty error message, got %q", errorMessage)
	}
	if baseline, ok := data["ChartBaseline"].(int); !ok || baseline != 0 {
		t.Fatalf("expected ChartBaseline=0 for partner, got %#v", data["ChartBaseline"])
	}
	if isOwner, ok := data["IsOwner"].(bool); !ok || isOwner {
		t.Fatalf("expected IsOwner=false for partner, got %#v", data["IsOwner"])
	}
}
