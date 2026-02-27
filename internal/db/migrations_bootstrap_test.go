package db

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	embeddedmigrations "github.com/terraincognita07/ovumcy/migrations"
	"gorm.io/gorm"
)

func TestOpenSQLiteAppliesEmbeddedMigrationsOnCleanDatabase(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "ovumcy-clean.db")
	database := openSQLiteForMigrationBootstrapTest(t, databasePath)

	assertUsersSchemaReconciled(t, database)
	assertDailyLogsSchemaReconciled(t, database)
	assertNormalizedEmailIndexExists(t, database)
	assertAllEmbeddedMigrationsApplied(t, database)
}

func TestOpenSQLiteUpgradesLegacyInitSchema(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "ovumcy-legacy.db")
	seedLegacyInitSchema(t, databasePath)

	database := openSQLiteForMigrationBootstrapTest(t, databasePath)

	assertUsersSchemaReconciled(t, database)
	assertDailyLogsSchemaReconciled(t, database)
	assertNormalizedEmailIndexExists(t, database)
	assertAllEmbeddedMigrationsApplied(t, database)

	var migratedUser struct {
		Email               string `gorm:"column:email"`
		DisplayName         string `gorm:"column:display_name"`
		OnboardingCompleted bool   `gorm:"column:onboarding_completed"`
		CycleLength         int    `gorm:"column:cycle_length"`
		PeriodLength        int    `gorm:"column:period_length"`
		AutoPeriodFill      bool   `gorm:"column:auto_period_fill"`
	}
	if err := database.
		Table("users").
		Select("email", "display_name", "onboarding_completed", "cycle_length", "period_length", "auto_period_fill").
		Where("email = ?", "legacy@example.com").
		First(&migratedUser).Error; err != nil {
		t.Fatalf("load migrated legacy user: %v", err)
	}

	if migratedUser.DisplayName != "" {
		t.Fatalf("expected display_name default to be empty, got %q", migratedUser.DisplayName)
	}
	if migratedUser.OnboardingCompleted {
		t.Fatal("expected onboarding_completed default to be false")
	}
	if migratedUser.CycleLength != 28 {
		t.Fatalf("expected cycle_length default to be 28, got %d", migratedUser.CycleLength)
	}
	if migratedUser.PeriodLength != 5 {
		t.Fatalf("expected period_length default to be 5, got %d", migratedUser.PeriodLength)
	}
	if !migratedUser.AutoPeriodFill {
		t.Fatal("expected auto_period_fill default to be true")
	}

	var migratedLog struct {
		Flow       string  `gorm:"column:flow"`
		SymptomIDs *string `gorm:"column:symptom_ids"`
		Notes      string  `gorm:"column:notes"`
	}
	if err := database.
		Table("daily_logs").
		Select("flow", "symptom_ids", "notes").
		Where("notes = ?", "legacy-log").
		First(&migratedLog).Error; err != nil {
		t.Fatalf("load migrated legacy daily log: %v", err)
	}

	if migratedLog.Flow != "light" {
		t.Fatalf("expected migrated flow=light, got %q", migratedLog.Flow)
	}
	if migratedLog.SymptomIDs == nil || strings.TrimSpace(*migratedLog.SymptomIDs) != "[1,2]" {
		t.Fatalf("expected migrated symptom_ids to remain [1,2], got %v", migratedLog.SymptomIDs)
	}
}

func TestOpenSQLiteMigrationBootstrapIsIdempotent(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "ovumcy-idempotent.db")

	firstOpen, err := OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("first open sqlite: %v", err)
	}
	firstRecords := loadMigrationRecords(t, firstOpen)

	firstSQLDB, err := firstOpen.DB()
	if err != nil {
		t.Fatalf("first open sql db: %v", err)
	}
	if err := firstSQLDB.Close(); err != nil {
		t.Fatalf("close first sql db: %v", err)
	}

	secondOpen := openSQLiteForMigrationBootstrapTest(t, databasePath)
	secondRecords := loadMigrationRecords(t, secondOpen)

	if !reflect.DeepEqual(firstRecords, secondRecords) {
		t.Fatalf("expected migration records to remain unchanged between boots, before=%v after=%v", firstRecords, secondRecords)
	}
}

func openSQLiteForMigrationBootstrapTest(t *testing.T, databasePath string) *gorm.DB {
	t.Helper()

	database, err := OpenSQLite(databasePath)
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

	return database
}

func seedLegacyInitSchema(t *testing.T, databasePath string) {
	t.Helper()

	dsn := fmt.Sprintf("%s?_foreign_keys=on&_busy_timeout=5000", databasePath)
	database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open legacy sqlite: %v", err)
	}

	initSQL, err := fs.ReadFile(embeddedmigrations.Files, "001_init.sql")
	if err != nil {
		t.Fatalf("read 001 migration: %v", err)
	}
	if err := database.Exec(string(initSQL)).Error; err != nil {
		t.Fatalf("apply 001 migration: %v", err)
	}

	if err := database.Exec(
		`INSERT INTO users (email, password_hash, role, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)`,
		"legacy@example.com",
		"legacy-hash",
		"owner",
	).Error; err != nil {
		t.Fatalf("insert legacy user: %v", err)
	}

	var legacyUser struct {
		ID uint `gorm:"column:id"`
	}
	if err := database.Raw(`SELECT id FROM users WHERE email = ?`, "legacy@example.com").Scan(&legacyUser).Error; err != nil {
		t.Fatalf("load legacy user id: %v", err)
	}
	if legacyUser.ID == 0 {
		t.Fatal("expected non-zero legacy user id")
	}

	if err := database.Exec(
		`INSERT INTO daily_logs (user_id, date, is_period, flow, symptom_ids, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		legacyUser.ID,
		"2026-01-10",
		true,
		"light",
		"[1,2]",
		"legacy-log",
	).Error; err != nil {
		t.Fatalf("insert legacy daily log: %v", err)
	}

	if database.Migrator().HasTable("schema_migrations") {
		t.Fatal("expected legacy schema to not have schema_migrations table")
	}

	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("open legacy sql db: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close legacy sql db: %v", err)
	}
}

func assertUsersSchemaReconciled(t *testing.T, database *gorm.DB) {
	t.Helper()

	columns := loadTableColumns(t, database, "users")
	expectedColumns := []string{
		"display_name",
		"onboarding_completed",
		"cycle_length",
		"period_length",
		"auto_period_fill",
		"last_period_start",
	}

	for _, column := range expectedColumns {
		if _, exists := columns[column]; !exists {
			t.Fatalf("expected users.%s column to exist after migrations", column)
		}
	}
}

func assertDailyLogsSchemaReconciled(t *testing.T, database *gorm.DB) {
	t.Helper()

	columns := loadTableColumns(t, database, "daily_logs")
	if _, exists := columns["symptom_ids"]; !exists {
		t.Fatal("expected daily_logs.symptom_ids column to exist after migrations")
	}

	notNullFlags := loadTableColumnNotNullFlags(t, database, "daily_logs")
	if notNullFlags["symptom_ids"] {
		t.Fatal("expected daily_logs.symptom_ids to remain nullable")
	}

	tableDefinition := loadSQLiteObjectSQL(t, database, "table", "daily_logs")
	normalized := strings.ToLower(strings.Join(strings.Fields(tableDefinition), ""))
	if strings.Contains(normalized, "check(flowin(") {
		t.Fatalf("expected daily_logs flow CHECK constraint to be removed, got %q", tableDefinition)
	}
}

func assertNormalizedEmailIndexExists(t *testing.T, database *gorm.DB) {
	t.Helper()

	indexSQL := loadSQLiteObjectSQL(t, database, "index", "idx_users_email_normalized")
	definition := strings.ToLower(strings.Join(strings.Fields(indexSQL), ""))
	if definition == "" {
		t.Fatal("expected normalized email index definition to exist")
	}
	if !strings.Contains(definition, "lower(trim(email))") {
		t.Fatalf("expected normalized email index to use lower(trim(email)), got %q", indexSQL)
	}
}

func assertAllEmbeddedMigrationsApplied(t *testing.T, database *gorm.DB) {
	t.Helper()

	expectedVersions := embeddedMigrationVersionsForTest(t)
	actualVersions := make([]string, 0)

	var rows []struct {
		Version string `gorm:"column:version"`
	}
	if err := database.Raw(`SELECT version FROM schema_migrations ORDER BY version ASC`).Scan(&rows).Error; err != nil {
		t.Fatalf("load applied migration versions: %v", err)
	}
	for _, row := range rows {
		actualVersions = append(actualVersions, row.Version)
	}

	if !reflect.DeepEqual(expectedVersions, actualVersions) {
		t.Fatalf("unexpected applied migration versions: expected=%v actual=%v", expectedVersions, actualVersions)
	}
}

type migrationRecord struct {
	Version   string `gorm:"column:version"`
	Name      string `gorm:"column:name"`
	AppliedAt string `gorm:"column:applied_at"`
}

func loadMigrationRecords(t *testing.T, database *gorm.DB) []migrationRecord {
	t.Helper()

	records := make([]migrationRecord, 0)
	if err := database.Raw(
		`SELECT version, name, applied_at FROM schema_migrations ORDER BY version ASC`,
	).Scan(&records).Error; err != nil {
		t.Fatalf("load migration records: %v", err)
	}
	return records
}

func loadTableColumns(t *testing.T, database *gorm.DB, tableName string) map[string]struct{} {
	t.Helper()

	escapedTable := strings.ReplaceAll(tableName, `"`, `""`)
	query := fmt.Sprintf(`PRAGMA table_info("%s")`, escapedTable)

	var rows []struct {
		Name string `gorm:"column:name"`
	}
	if err := database.Raw(query).Scan(&rows).Error; err != nil {
		t.Fatalf("load table columns for %s: %v", tableName, err)
	}

	columns := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		columns[strings.ToLower(strings.TrimSpace(row.Name))] = struct{}{}
	}
	return columns
}

func loadTableColumnNotNullFlags(t *testing.T, database *gorm.DB, tableName string) map[string]bool {
	t.Helper()

	escapedTable := strings.ReplaceAll(tableName, `"`, `""`)
	query := fmt.Sprintf(`PRAGMA table_info("%s")`, escapedTable)

	var rows []struct {
		Name    string `gorm:"column:name"`
		NotNull int    `gorm:"column:notnull"`
	}
	if err := database.Raw(query).Scan(&rows).Error; err != nil {
		t.Fatalf("load table nullability for %s: %v", tableName, err)
	}

	flags := make(map[string]bool, len(rows))
	for _, row := range rows {
		flags[strings.ToLower(strings.TrimSpace(row.Name))] = row.NotNull == 1
	}
	return flags
}

func loadSQLiteObjectSQL(t *testing.T, database *gorm.DB, objectType string, objectName string) string {
	t.Helper()

	var row struct {
		SQL string `gorm:"column:sql"`
	}
	if err := database.Raw(
		`SELECT sql FROM sqlite_master WHERE type = ? AND name = ?`,
		objectType,
		objectName,
	).Scan(&row).Error; err != nil {
		t.Fatalf("load sqlite master sql for %s %s: %v", objectType, objectName, err)
	}
	return row.SQL
}

func embeddedMigrationVersionsForTest(t *testing.T) []string {
	t.Helper()

	migrations, err := loadEmbeddedMigrations()
	if err != nil {
		t.Fatalf("load embedded migrations: %v", err)
	}

	versions := make([]string, 0, len(migrations))
	for _, migration := range migrations {
		versions = append(versions, migration.Version)
	}
	return versions
}
