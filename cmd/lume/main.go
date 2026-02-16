package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/terraincognita07/lume/internal/api"
	"github.com/terraincognita07/lume/internal/db"
	"github.com/terraincognita07/lume/internal/i18n"
	"github.com/terraincognita07/lume/internal/services"
)

func main() {
	location := mustLoadLocation(getEnv("TZ", "UTC"))
	time.Local = location

	secretKey := getEnv("SECRET_KEY", "change_me_in_production")
	dbPath := getEnv("DB_PATH", filepath.Join("data", "lume.db"))
	port := getEnv("PORT", "8080")
	defaultLanguage := getEnv("DEFAULT_LANGUAGE", "ru")

	database, err := db.OpenSQLite(dbPath)
	if err != nil {
		log.Fatalf("database init failed: %v", err)
	}

	i18nManager, err := i18n.NewManager(defaultLanguage, filepath.Join("internal", "i18n", "locales"))
	if err != nil {
		log.Fatalf("i18n init failed: %v", err)
	}

	handler, err := api.NewHandler(database, secretKey, filepath.Join("internal", "templates"), location, i18nManager)
	if err != nil {
		log.Fatalf("handler init failed: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:               "Lume",
		DisableStartupMessage: true,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(compress.New())
	app.Use(handler.LanguageMiddleware)
	app.Use(csrf.New(csrf.Config{
		KeyLookup:      "form:csrf_token",
		CookieName:     "lume_csrf",
		CookieSameSite: "Lax",
		CookieHTTPOnly: false,
		CookieSecure:   false,
		ContextKey:     "csrf",
	}))

	app.Static("/static", filepath.Join("web", "static"))
	api.RegisterRoutes(app, handler)

	notifier := services.NewNotificationService(database, location)
	lifecycleCtx, cancelLifecycle := context.WithCancel(context.Background())
	defer cancelLifecycle()
	notifier.Start(lifecycleCtx)

	sigCtx, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	go func() {
		<-sigCtx.Done()
		cancelLifecycle()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			log.Printf("server shutdown failed: %v", err)
		}
	}()

	log.Printf("Lume listening on http://0.0.0.0:%s (db: %s, tz: %s)", port, dbPath, location.String())
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

func mustLoadLocation(name string) *time.Location {
	location, err := time.LoadLocation(name)
	if err != nil {
		log.Printf("invalid TZ %q, falling back to UTC", name)
		return time.UTC
	}
	return location
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
