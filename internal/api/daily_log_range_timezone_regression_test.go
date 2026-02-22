package api

import (
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/db"
	"github.com/terraincognita07/lume/internal/i18n"
	"github.com/terraincognita07/lume/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func TestFetchLogsForUserIncludesUTCShiftedRowForLocalDayRange(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file path")
	}

	apiDir := filepath.Dir(testFile)
	internalDir := filepath.Dir(apiDir)
	templatesDir := filepath.Join(internalDir, "templates")
	localesDir := filepath.Join(internalDir, "i18n", "locales")
	databasePath := filepath.Join(t.TempDir(), "lume-zulu-shifted-range.db")

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

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("StrongPass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := models.User{
		Email:               "zulu-shifted-range@example.com",
		PasswordHash:        string(passwordHash),
		Role:                models.RoleOwner,
		OnboardingCompleted: true,
		CycleLength:         28,
		PeriodLength:        5,
		CreatedAt:           time.Now().UTC(),
	}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	now := time.Now().UTC()
	if err := database.Exec(
		`INSERT INTO daily_logs (user_id, date, is_period, flow, symptom_ids, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID,
		"2026-02-21T21:00:00Z",
		true,
		models.FlowHeavy,
		"[]",
		"",
		now,
		now,
	).Error; err != nil {
		t.Fatalf("insert utc shifted row: %v", err)
	}

	i18nManager, err := i18n.NewManager("en", localesDir)
	if err != nil {
		t.Fatalf("init i18n: %v", err)
	}

	moscow := time.FixedZone("UTC+3", 3*60*60)
	handler, err := NewHandler(database, "test-secret-key", templatesDir, moscow, i18nManager, false)
	if err != nil {
		t.Fatalf("init handler: %v", err)
	}

	from, err := parseDayParam("2026-02-22", moscow)
	if err != nil {
		t.Fatalf("parse from day: %v", err)
	}
	to := from

	logs, err := handler.fetchLogsForUser(user.ID, from, to)
	if err != nil {
		t.Fatalf("fetchLogsForUser: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected exactly one row in local-day range, got %d", len(logs))
	}
	if !logs[0].IsPeriod {
		t.Fatalf("expected is_period=true for shifted row")
	}
	if logs[0].Flow != models.FlowHeavy {
		t.Fatalf("expected flow %q, got %q", models.FlowHeavy, logs[0].Flow)
	}
}
