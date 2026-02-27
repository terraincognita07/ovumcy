package db

import (
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"

	embeddedmigrations "github.com/terraincognita07/ovumcy/migrations"
	"gorm.io/gorm"
)

var migrationFilePattern = regexp.MustCompile(`^(\d+)_.*\.sql$`)
var addColumnStatementPattern = regexp.MustCompile(`(?i)^ALTER\s+TABLE\s+([^\s]+)\s+ADD\s+COLUMN\s+([^\s]+)\b`)

type embeddedMigration struct {
	Version string
	Order   int
	Name    string
	SQL     string
}

func applyEmbeddedMigrations(database *gorm.DB) error {
	if err := ensureSchemaMigrationsTable(database); err != nil {
		return err
	}

	migrations, err := loadEmbeddedMigrations()
	if err != nil {
		return err
	}

	appliedVersions, err := loadAppliedMigrationVersions(database)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if _, alreadyApplied := appliedVersions[migration.Version]; alreadyApplied {
			continue
		}

		if err := applyMigration(database, migration); err != nil {
			return err
		}
	}

	return nil
}

func ensureSchemaMigrationsTable(database *gorm.DB) error {
	const createTableSQL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
	if err := database.Exec(createTableSQL).Error; err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}
	return nil
}

func loadEmbeddedMigrations() ([]embeddedMigration, error) {
	entries, err := fs.ReadDir(embeddedmigrations.Files, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}

	migrations := make([]embeddedMigration, 0, len(entries))
	seenVersions := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := strings.TrimSpace(entry.Name())
		matches := migrationFilePattern.FindStringSubmatch(fileName)
		if len(matches) != 2 {
			continue
		}

		version := matches[1]
		order, err := strconv.Atoi(version)
		if err != nil {
			return nil, fmt.Errorf("parse migration version from %s: %w", fileName, err)
		}

		if existing, exists := seenVersions[version]; exists {
			return nil, fmt.Errorf("duplicate migration version %s in %s and %s", version, existing, fileName)
		}
		seenVersions[version] = fileName

		rawSQL, err := fs.ReadFile(embeddedmigrations.Files, fileName)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", fileName, err)
		}

		migrations = append(migrations, embeddedMigration{
			Version: version,
			Order:   order,
			Name:    fileName,
			SQL:     string(rawSQL),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		if migrations[i].Order == migrations[j].Order {
			return migrations[i].Name < migrations[j].Name
		}
		return migrations[i].Order < migrations[j].Order
	})

	return migrations, nil
}

type appliedMigrationVersion struct {
	Version string `gorm:"column:version"`
}

func loadAppliedMigrationVersions(database *gorm.DB) (map[string]struct{}, error) {
	rows := make([]appliedMigrationVersion, 0)
	if err := database.Raw(`SELECT version FROM schema_migrations`).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("load applied migration versions: %w", err)
	}

	versions := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		versions[row.Version] = struct{}{}
	}
	return versions, nil
}

func applyMigration(database *gorm.DB, migration embeddedMigration) error {
	return database.Transaction(func(tx *gorm.DB) error {
		statements := splitSQLStatements(migration.SQL)
		if len(statements) == 0 {
			return errors.New("migration has no SQL statements")
		}

		for _, statement := range statements {
			skip, err := shouldSkipStatement(tx, statement)
			if err != nil {
				return fmt.Errorf("inspect migration %s: %w", migration.Name, err)
			}
			if skip {
				continue
			}

			if err := tx.Exec(statement).Error; err != nil {
				return fmt.Errorf("execute migration %s statement %q: %w", migration.Name, statement, err)
			}
		}

		if err := tx.Exec(
			`INSERT INTO schema_migrations(version, name) VALUES (?, ?)`,
			migration.Version,
			migration.Name,
		).Error; err != nil {
			return fmt.Errorf("record migration %s: %w", migration.Name, err)
		}

		return nil
	})
}

func splitSQLStatements(sqlText string) []string {
	rawParts := strings.Split(sqlText, ";")
	statements := make([]string, 0, len(rawParts))
	for _, rawPart := range rawParts {
		statement := strings.TrimSpace(rawPart)
		if statement == "" {
			continue
		}
		statements = append(statements, statement)
	}
	return statements
}

func shouldSkipStatement(database *gorm.DB, statement string) (bool, error) {
	matches := addColumnStatementPattern.FindStringSubmatch(strings.TrimSpace(statement))
	if len(matches) != 3 {
		return false, nil
	}

	tableName := normalizeSQLIdentifier(matches[1])
	columnName := normalizeSQLIdentifier(matches[2])
	exists, err := tableColumnExists(database, tableName, columnName)
	if err != nil {
		return false, err
	}
	return exists, nil
}

type pragmaTableColumn struct {
	Name string `gorm:"column:name"`
}

func tableColumnExists(database *gorm.DB, tableName string, columnName string) (bool, error) {
	escapedTable := strings.ReplaceAll(tableName, `"`, `""`)
	query := fmt.Sprintf(`PRAGMA table_info("%s")`, escapedTable)

	columns := make([]pragmaTableColumn, 0)
	if err := database.Raw(query).Scan(&columns).Error; err != nil {
		return false, fmt.Errorf("load table_info for %s: %w", tableName, err)
	}
	for _, column := range columns {
		if strings.EqualFold(strings.TrimSpace(column.Name), columnName) {
			return true, nil
		}
	}
	return false, nil
}

func normalizeSQLIdentifier(identifier string) string {
	normalized := strings.TrimSpace(identifier)
	normalized = strings.Trim(normalized, "\"`[]")
	return strings.TrimSpace(normalized)
}
