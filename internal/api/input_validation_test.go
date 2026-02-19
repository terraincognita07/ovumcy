package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestParseCredentialsValidation(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Post("/credentials", func(c *fiber.Ctx) error {
		credentials, err := parseCredentials(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(credentials)
	})

	t.Run("valid form input", func(t *testing.T) {
		form := url.Values{}
		form.Set("email", "USER@EXAMPLE.COM")
		form.Set("password", "StrongPass1")
		form.Set("confirm_password", "StrongPass1")
		form.Set("remember_me", "1")

		req := httptest.NewRequest(http.MethodPost, "/credentials", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var payload credentialsInput
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if payload.Email != "user@example.com" {
			t.Fatalf("expected normalized email, got %q", payload.Email)
		}
		if !payload.RememberMe {
			t.Fatal("expected remember_me=true from form value")
		}
	})

	t.Run("invalid email is rejected", func(t *testing.T) {
		form := url.Values{}
		form.Set("email", "not-email")
		form.Set("password", "StrongPass1")

		req := httptest.NewRequest(http.MethodPost, "/credentials", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", resp.StatusCode)
		}
	})
}

func TestParseDayPayloadSources(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Post("/day", func(c *fiber.Ctx) error {
		payload, err := parseDayPayload(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(payload)
	})

	t.Run("parses JSON payload", func(t *testing.T) {
		body := `{"is_period":true,"flow":"heavy","symptom_ids":[1,3],"notes":"abc"}`
		req := httptest.NewRequest(http.MethodPost, "/day", strings.NewReader(body))
		req.Header.Set("Content-Type", fiber.MIMEApplicationJSON)

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var payload dayPayload
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if !payload.IsPeriod || payload.Flow != "heavy" || len(payload.SymptomIDs) != 2 || payload.Notes != "abc" {
			t.Fatalf("unexpected payload parsed from json: %+v", payload)
		}
	})

	t.Run("parses form payload and normalizes", func(t *testing.T) {
		form := url.Values{}
		form.Set("is_period", "on")
		form.Set("flow", " Medium ")
		form.Add("symptom_ids", "2")
		form.Add("symptom_ids", "4")
		form.Set("notes", " note ")

		req := httptest.NewRequest(http.MethodPost, "/day", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var payload dayPayload
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if !payload.IsPeriod {
			t.Fatal("expected is_period=true from form")
		}
		if payload.Flow != "medium" {
			t.Fatalf("expected normalized flow=medium, got %q", payload.Flow)
		}
		if payload.Notes != "note" {
			t.Fatalf("expected trimmed notes, got %q", payload.Notes)
		}
		if len(payload.SymptomIDs) != 2 || payload.SymptomIDs[0] != 2 || payload.SymptomIDs[1] != 4 {
			t.Fatalf("unexpected symptom IDs: %#v", payload.SymptomIDs)
		}
	})
}
