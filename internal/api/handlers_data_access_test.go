package api

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/db"
	"github.com/terraincognita07/lume/internal/models"
	"gorm.io/gorm"
)

func newDataAccessTestHandler(t *testing.T) (*Handler, *gorm.DB) {
	t.Helper()

	databasePath := filepath.Join(t.TempDir(), "lume-data-access-test.db")
	database, err := db.OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	handler := &Handler{
		db:       database,
		location: time.UTC,
	}
	return handler, database
}

func createDataAccessTestUser(t *testing.T, database *gorm.DB, email string) models.User {
	t.Helper()

	user := models.User{
		Email:               email,
		PasswordHash:        "test-hash",
		Role:                models.RoleOwner,
		OnboardingCompleted: true,
		CycleLength:         28,
		PeriodLength:        5,
		CreatedAt:           time.Now().UTC(),
	}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func TestValidateSymptomIDsDeduplicatesAndSorts(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "validate-symptoms@example.com")

	symptoms := []models.SymptomType{
		{UserID: user.ID, Name: "A", Icon: "A", Color: "#111111"},
		{UserID: user.ID, Name: "B", Icon: "B", Color: "#222222"},
	}
	if err := database.Create(&symptoms).Error; err != nil {
		t.Fatalf("create symptoms: %v", err)
	}

	ids, err := handler.validateSymptomIDs(user.ID, []uint{symptoms[1].ID, symptoms[0].ID, symptoms[1].ID})
	if err != nil {
		t.Fatalf("validateSymptomIDs returned error: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 ids, got %d", len(ids))
	}
	if ids[0] != symptoms[0].ID || ids[1] != symptoms[1].ID {
		t.Fatalf("expected sorted unique ids [%d %d], got %#v", symptoms[0].ID, symptoms[1].ID, ids)
	}

	if _, err := handler.validateSymptomIDs(user.ID, []uint{symptoms[0].ID, 999999}); err == nil {
		t.Fatal("expected error for invalid symptom id")
	}
}

func TestRemoveSymptomFromLogsUpdatesEntries(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "remove-symptom@example.com")

	symptoms := []models.SymptomType{
		{UserID: user.ID, Name: "A", Icon: "A", Color: "#111111"},
		{UserID: user.ID, Name: "B", Icon: "B", Color: "#222222"},
	}
	if err := database.Create(&symptoms).Error; err != nil {
		t.Fatalf("create symptoms: %v", err)
	}

	day := time.Date(2026, time.February, 17, 0, 0, 0, 0, time.UTC)
	entry := models.DailyLog{
		UserID:     user.ID,
		Date:       day,
		IsPeriod:   false,
		Flow:       models.FlowNone,
		SymptomIDs: []uint{symptoms[0].ID, symptoms[1].ID, symptoms[0].ID},
	}
	if err := database.Create(&entry).Error; err != nil {
		t.Fatalf("create log: %v", err)
	}

	if err := handler.removeSymptomFromLogs(user.ID, symptoms[0].ID); err != nil {
		t.Fatalf("removeSymptomFromLogs returned error: %v", err)
	}

	updated := models.DailyLog{}
	if err := database.First(&updated, entry.ID).Error; err != nil {
		t.Fatalf("load updated log: %v", err)
	}
	if len(updated.SymptomIDs) != 1 || updated.SymptomIDs[0] != symptoms[1].ID {
		t.Fatalf("expected only remaining symptom id %d, got %#v", symptoms[1].ID, updated.SymptomIDs)
	}
}

func TestDayHasDataForDate(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "day-has-data@example.com")

	day := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)

	hasData, err := handler.dayHasDataForDate(user.ID, day)
	if err != nil {
		t.Fatalf("dayHasDataForDate returned error: %v", err)
	}
	if hasData {
		t.Fatal("expected false when no entries exist")
	}

	entry := models.DailyLog{
		UserID:   user.ID,
		Date:     day,
		IsPeriod: false,
		Flow:     models.FlowNone,
		Notes:    "note",
	}
	if err := database.Create(&entry).Error; err != nil {
		t.Fatalf("create log: %v", err)
	}

	hasData, err = handler.dayHasDataForDate(user.ID, day)
	if err != nil {
		t.Fatalf("dayHasDataForDate returned error: %v", err)
	}
	if !hasData {
		t.Fatal("expected true when notes exist for the day")
	}
}

func TestRefreshUserLastPeriodStart(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "refresh-last-period@example.com")

	first := time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	second := time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC)
	logs := []models.DailyLog{
		{UserID: user.ID, Date: first, IsPeriod: true, Flow: models.FlowMedium, SymptomIDs: []uint{}},
		{UserID: user.ID, Date: second, IsPeriod: true, Flow: models.FlowMedium, SymptomIDs: []uint{}},
	}
	if err := database.Create(&logs).Error; err != nil {
		t.Fatalf("create period logs: %v", err)
	}

	if err := handler.refreshUserLastPeriodStart(user.ID); err != nil {
		t.Fatalf("refreshUserLastPeriodStart returned error: %v", err)
	}

	updated := models.User{}
	if err := database.First(&updated, user.ID).Error; err != nil {
		t.Fatalf("load updated user: %v", err)
	}
	if updated.LastPeriodStart == nil {
		t.Fatal("expected last_period_start to be populated")
	}
	if updated.LastPeriodStart.Format("2006-01-02") != second.Format("2006-01-02") {
		t.Fatalf("expected latest period start %s, got %s", second.Format("2006-01-02"), updated.LastPeriodStart.Format("2006-01-02"))
	}

	if err := database.Where("user_id = ?", user.ID).Delete(&models.DailyLog{}).Error; err != nil {
		t.Fatalf("delete logs: %v", err)
	}
	if err := handler.refreshUserLastPeriodStart(user.ID); err != nil {
		t.Fatalf("refreshUserLastPeriodStart second call returned error: %v", err)
	}

	updated = models.User{}
	if err := database.First(&updated, user.ID).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if updated.LastPeriodStart != nil {
		t.Fatalf("expected last_period_start to be cleared, got %v", updated.LastPeriodStart)
	}
}
