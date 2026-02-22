package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/terraincognita07/lume/internal/models"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func OpenSQLite(dbPath string) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	dsn := fmt.Sprintf("%s?_foreign_keys=on&_busy_timeout=5000", dbPath)
	database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormlogger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  gormlogger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := database.AutoMigrate(
		&models.User{},
		&models.SymptomType{},
		&models.DailyLog{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}
	if err := database.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_normalized ON users(lower(trim(email)))`).Error; err != nil {
		return nil, fmt.Errorf("create normalized email index: %w", err)
	}

	return database, nil
}
