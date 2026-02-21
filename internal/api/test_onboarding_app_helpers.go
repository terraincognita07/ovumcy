package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/db"
	"github.com/terraincognita07/lume/internal/i18n"
	"github.com/terraincognita07/lume/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func newOnboardingTestApp(t *testing.T) (*fiber.App, *gorm.DB) {
	t.Helper()
	return newOnboardingTestAppWithCookieSecure(t, false)
}

func newOnboardingTestAppWithCookieSecure(t *testing.T, cookieSecure bool) (*fiber.App, *gorm.DB) {
	t.Helper()

	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file path")
	}

	apiDir := filepath.Dir(testFile)
	internalDir := filepath.Dir(apiDir)
	templatesDir := filepath.Join(internalDir, "templates")
	localesDir := filepath.Join(internalDir, "i18n", "locales")
	databasePath := filepath.Join(t.TempDir(), "lume-onboarding-test.db")

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

	i18nManager, err := i18n.NewManager("en", localesDir)
	if err != nil {
		t.Fatalf("init i18n: %v", err)
	}

	handler, err := NewHandler(database, "test-secret-key", templatesDir, time.UTC, i18nManager, cookieSecure)
	if err != nil {
		t.Fatalf("init handler: %v", err)
	}

	app := fiber.New()
	app.Use(handler.LanguageMiddleware)
	RegisterRoutes(app, handler)
	return app, database
}

func createOnboardingTestUser(t *testing.T, database *gorm.DB, email string, password string, onboardingCompleted bool) models.User {
	t.Helper()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := models.User{
		Email:               strings.ToLower(strings.TrimSpace(email)),
		PasswordHash:        string(passwordHash),
		Role:                models.RoleOwner,
		OnboardingCompleted: onboardingCompleted,
		CycleLength:         28,
		PeriodLength:        5,
		AutoPeriodFill:      true,
		CreatedAt:           time.Now().UTC(),
	}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func loginAndExtractAuthCookie(t *testing.T, app *fiber.App, email string, password string) string {
	t.Helper()

	form := url.Values{
		"email":    {email},
		"password": {password},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected login status 303, got %d", response.StatusCode)
	}

	for _, cookie := range response.Cookies() {
		if cookie.Name == "lume_auth" && cookie.Value != "" {
			return cookie.Name + "=" + cookie.Value
		}
	}

	t.Fatal("auth cookie is missing in login response")
	return ""
}
