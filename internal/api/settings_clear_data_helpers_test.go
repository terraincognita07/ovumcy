package api

import (
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"gorm.io/gorm"
)

type clearDataScenario struct {
	app           *fiber.App
	database      *gorm.DB
	user          models.User
	authCookie    string
	customSymptom models.SymptomType
}

func setupClearDataScenario(t *testing.T) clearDataScenario {
	t.Helper()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "clear-data@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	lastPeriodStart := time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"cycle_length":      41,
		"period_length":     8,
		"auto_period_fill":  false,
		"last_period_start": lastPeriodStart,
	}).Error; err != nil {
		t.Fatalf("update user baseline settings: %v", err)
	}

	builtinSymptom := models.SymptomType{
		UserID:    user.ID,
		Name:      "Builtin",
		Icon:      "A",
		Color:     "#111111",
		IsBuiltin: true,
	}
	customSymptom := models.SymptomType{
		UserID:    user.ID,
		Name:      "Custom",
		Icon:      "B",
		Color:     "#222222",
		IsBuiltin: false,
	}
	if err := database.Create(&builtinSymptom).Error; err != nil {
		t.Fatalf("create builtin symptom: %v", err)
	}
	if err := database.Create(&customSymptom).Error; err != nil {
		t.Fatalf("create custom symptom: %v", err)
	}

	logEntry := models.DailyLog{
		UserID:     user.ID,
		Date:       time.Date(2026, time.February, 12, 0, 0, 0, 0, time.UTC),
		IsPeriod:   true,
		Flow:       models.FlowMedium,
		SymptomIDs: []uint{customSymptom.ID},
		Notes:      "test",
	}
	if err := database.Create(&logEntry).Error; err != nil {
		t.Fatalf("create log entry: %v", err)
	}

	return clearDataScenario{
		app:           app,
		database:      database,
		user:          user,
		authCookie:    authCookie,
		customSymptom: customSymptom,
	}
}

func assertClearDataPostconditions(t *testing.T, database *gorm.DB, user models.User) {
	t.Helper()

	var logsCount int64
	if err := database.Model(&models.DailyLog{}).Where("user_id = ?", user.ID).Count(&logsCount).Error; err != nil {
		t.Fatalf("count logs: %v", err)
	}
	if logsCount != 0 {
		t.Fatalf("expected logs to be deleted, got %d", logsCount)
	}

	var customCount int64
	if err := database.Model(&models.SymptomType{}).Where("user_id = ? AND is_builtin = ?", user.ID, false).Count(&customCount).Error; err != nil {
		t.Fatalf("count custom symptoms: %v", err)
	}
	if customCount != 1 {
		t.Fatalf("expected custom symptoms to stay unchanged, got %d", customCount)
	}

	var builtinCount int64
	if err := database.Model(&models.SymptomType{}).Where("user_id = ? AND is_builtin = ?", user.ID, true).Count(&builtinCount).Error; err != nil {
		t.Fatalf("count builtin symptoms: %v", err)
	}
	if builtinCount != 1 {
		t.Fatalf("expected builtin symptoms to be preserved, got %d", builtinCount)
	}

	updatedUser := models.User{}
	if err := database.First(&updatedUser, user.ID).Error; err != nil {
		t.Fatalf("load updated user: %v", err)
	}
	if updatedUser.CycleLength != 28 {
		t.Fatalf("expected cycle length reset to 28, got %d", updatedUser.CycleLength)
	}
	if updatedUser.PeriodLength != 5 {
		t.Fatalf("expected period length reset to 5, got %d", updatedUser.PeriodLength)
	}
	if !updatedUser.AutoPeriodFill {
		t.Fatalf("expected auto period fill reset to true")
	}
	if updatedUser.LastPeriodStart != nil {
		t.Fatalf("expected last period start to be cleared, got %v", updatedUser.LastPeriodStart)
	}
}
