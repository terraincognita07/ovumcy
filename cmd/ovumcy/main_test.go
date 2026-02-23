package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/api"
)

func TestResolveSecretKey(t *testing.T) {
	t.Setenv("SECRET_KEY", "")
	if _, err := resolveSecretKey(); err == nil {
		t.Fatal("expected error when SECRET_KEY is empty")
	}

	t.Setenv("SECRET_KEY", "change_me_in_production")
	if _, err := resolveSecretKey(); err == nil {
		t.Fatal("expected error when SECRET_KEY uses insecure placeholder")
	}

	t.Setenv("SECRET_KEY", "replace_with_at_least_32_random_characters")
	if _, err := resolveSecretKey(); err == nil {
		t.Fatal("expected error when SECRET_KEY uses example placeholder")
	}

	t.Setenv("SECRET_KEY", "too-short-secret")
	if _, err := resolveSecretKey(); err == nil {
		t.Fatal("expected error when SECRET_KEY is too short")
	}

	valid := "0123456789abcdef0123456789abcdef"
	t.Setenv("SECRET_KEY", valid)

	secret, err := resolveSecretKey()
	if err != nil {
		t.Fatalf("expected valid secret, got error: %v", err)
	}
	if secret != valid {
		t.Fatalf("expected %q, got %q", valid, secret)
	}
}

func TestCSRFMiddlewareConfigUsesCookieSecureFlag(t *testing.T) {
	secureConfig := csrfMiddlewareConfig(true)
	if !secureConfig.CookieSecure {
		t.Fatal("expected csrf cookie secure flag to be enabled")
	}
	if !secureConfig.CookieHTTPOnly {
		t.Fatal("expected csrf cookie to be httpOnly")
	}
	if secureConfig.CookieName != "ovumcy_csrf" {
		t.Fatalf("expected csrf cookie name ovumcy_csrf, got %q", secureConfig.CookieName)
	}
	if secureConfig.KeyLookup != "form:csrf_token" {
		t.Fatalf("expected csrf key lookup form:csrf_token, got %q", secureConfig.KeyLookup)
	}

	insecureConfig := csrfMiddlewareConfig(false)
	if insecureConfig.CookieSecure {
		t.Fatal("expected csrf cookie secure flag to be disabled")
	}
}

func TestResolvePort(t *testing.T) {
	t.Setenv("PORT", "")
	port, err := resolvePort()
	if err != nil {
		t.Fatalf("expected default port, got error: %v", err)
	}
	if port != "8080" {
		t.Fatalf("expected default port 8080, got %q", port)
	}

	t.Setenv("PORT", "9090")
	port, err = resolvePort()
	if err != nil {
		t.Fatalf("expected valid port, got error: %v", err)
	}
	if port != "9090" {
		t.Fatalf("expected port 9090, got %q", port)
	}

	t.Setenv("PORT", "0")
	if _, err := resolvePort(); err == nil {
		t.Fatal("expected invalid port 0 to fail")
	}

	t.Setenv("PORT", "70000")
	if _, err := resolvePort(); err == nil {
		t.Fatal("expected invalid high port to fail")
	}

	t.Setenv("PORT", "not-a-number")
	if _, err := resolvePort(); err == nil {
		t.Fatal("expected invalid non-numeric port to fail")
	}
}

func TestRedirectWithErrorCodeKeepsLoginEmailInFlash(t *testing.T) {
	app := fiber.New()
	app.Post("/limited", func(c *fiber.Ctx) error {
		return redirectWithErrorCode(c, "/login", "too_many_login_attempts", false)
	})

	form := "email=rate-limit%40example.com"
	request := httptest.NewRequest(http.MethodPost, "/limited", strings.NewReader(form))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("limited request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", response.StatusCode)
	}
	if location := response.Header.Get("Location"); location != "/login" {
		t.Fatalf("expected redirect location /login, got %q", location)
	}

	flashCookie := testResponseCookie(response.Cookies(), "ovumcy_flash")
	if flashCookie == nil {
		t.Fatal("expected flash cookie in rate-limit redirect response")
	}

	flash := decodeFlashPayload(t, flashCookie.Value)
	if flash.AuthError != "too_many_login_attempts" {
		t.Fatalf("expected auth error code in flash, got %q", flash.AuthError)
	}
	if flash.LoginEmail != "rate-limit@example.com" {
		t.Fatalf("expected login email to persist in flash, got %q", flash.LoginEmail)
	}
}

func decodeFlashPayload(t *testing.T, raw string) api.FlashPayload {
	t.Helper()

	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		t.Fatalf("decode flash payload: %v", err)
	}

	payload := api.FlashPayload{}
	if err := json.Unmarshal(decoded, &payload); err != nil {
		t.Fatalf("unmarshal flash payload: %v", err)
	}
	return payload
}

func testResponseCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie != nil && cookie.Name == name {
			return cookie
		}
	}
	return nil
}
