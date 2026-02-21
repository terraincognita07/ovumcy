package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func evaluateAuthPageBuilder(t *testing.T, query url.Values, handler fiber.Handler) map[string]any {
	t.Helper()

	app := fiber.New()
	app.Get("/", handler)

	requestPath := "/"
	if encoded := query.Encode(); encoded != "" {
		requestPath += "?" + encoded
	}
	request := httptest.NewRequest(http.MethodGet, requestPath, nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app test failed: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", response.StatusCode)
	}

	payload := map[string]any{}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload failed: %v", err)
	}
	return payload
}

func TestBuildLoginPageDataUsesFlashPriorityAndSetupFlag(t *testing.T) {
	t.Parallel()

	query := url.Values{
		"error": {"weak password"},
		"email": {"query@example.com"},
	}
	flash := FlashPayload{
		AuthError:  "invalid credentials",
		LoginEmail: " Flash@Example.com ",
	}

	payload := evaluateAuthPageBuilder(t, query, func(c *fiber.Ctx) error {
		return c.JSON(buildLoginPageData(c, map[string]string{}, flash, true))
	})

	if payload["ErrorKey"] != "auth.error.invalid_credentials" {
		t.Fatalf("expected flash error key, got %#v", payload["ErrorKey"])
	}
	if payload["Email"] != "flash@example.com" {
		t.Fatalf("expected normalized flash email, got %#v", payload["Email"])
	}
	if payload["IsFirstLaunch"] != true {
		t.Fatalf("expected IsFirstLaunch=true, got %#v", payload["IsFirstLaunch"])
	}
}

func TestBuildRegisterPageDataFallsBackToQueryValues(t *testing.T) {
	t.Parallel()

	query := url.Values{
		"error": {"weak password"},
		"email": {"Query@Example.com"},
	}

	payload := evaluateAuthPageBuilder(t, query, func(c *fiber.Ctx) error {
		return c.JSON(buildRegisterPageData(c, map[string]string{}, FlashPayload{}, false))
	})

	if payload["ErrorKey"] != "auth.error.weak_password" {
		t.Fatalf("expected query error key, got %#v", payload["ErrorKey"])
	}
	if payload["Email"] != "query@example.com" {
		t.Fatalf("expected normalized query email, got %#v", payload["Email"])
	}
	if payload["IsFirstLaunch"] != false {
		t.Fatalf("expected IsFirstLaunch=false, got %#v", payload["IsFirstLaunch"])
	}
}

func TestBuildForgotPasswordPageDataPrefersFlashError(t *testing.T) {
	t.Parallel()

	query := url.Values{
		"error": {"weak password"},
	}
	flash := FlashPayload{AuthError: "invalid credentials"}

	payload := evaluateAuthPageBuilder(t, query, func(c *fiber.Ctx) error {
		return c.JSON(buildForgotPasswordPageData(c, map[string]string{}, flash))
	})

	if payload["ErrorKey"] != "auth.error.invalid_credentials" {
		t.Fatalf("expected flash error key, got %#v", payload["ErrorKey"])
	}
}

func TestBuildResetPasswordPageDataValidTokenAndForcedFlag(t *testing.T) {
	t.Parallel()

	handler := &Handler{secretKey: []byte("test-reset-secret")}
	token, err := handler.buildPasswordResetToken(42, 30*time.Minute)
	if err != nil {
		t.Fatalf("buildPasswordResetToken returned error: %v", err)
	}

	query := url.Values{
		"token":  {token},
		"forced": {"1"},
		"error":  {"weak password"},
	}
	flash := FlashPayload{AuthError: "invalid credentials"}

	payload := evaluateAuthPageBuilder(t, query, func(c *fiber.Ctx) error {
		return c.JSON(handler.buildResetPasswordPageData(c, map[string]string{}, flash))
	})

	if payload["Token"] != token {
		t.Fatalf("expected token in payload, got %#v", payload["Token"])
	}
	if payload["InvalidToken"] != false {
		t.Fatalf("expected InvalidToken=false, got %#v", payload["InvalidToken"])
	}
	if payload["ForcedReset"] != true {
		t.Fatalf("expected ForcedReset=true, got %#v", payload["ForcedReset"])
	}
	if payload["ErrorKey"] != "auth.error.invalid_credentials" {
		t.Fatalf("expected flash error key, got %#v", payload["ErrorKey"])
	}
}

func TestBuildResetPasswordPageDataMarksInvalidToken(t *testing.T) {
	t.Parallel()

	handler := &Handler{secretKey: []byte("test-reset-secret")}
	query := url.Values{
		"token": {"invalid-token"},
	}

	payload := evaluateAuthPageBuilder(t, query, func(c *fiber.Ctx) error {
		return c.JSON(handler.buildResetPasswordPageData(c, map[string]string{}, FlashPayload{}))
	})

	if payload["InvalidToken"] != true {
		t.Fatalf("expected InvalidToken=true, got %#v", payload["InvalidToken"])
	}
}
