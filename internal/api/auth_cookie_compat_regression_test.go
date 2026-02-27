package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestLoginSetsSealedAuthCookieValue(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "sealed-auth-cookie@example.com", "StrongPass1", true)

	form := url.Values{
		"email":    {user.Email},
		"password": {"StrongPass1"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer response.Body.Close()

	authCookie := responseCookie(response.Cookies(), authCookieName)
	if authCookie == nil || strings.TrimSpace(authCookie.Value) == "" {
		t.Fatal("expected non-empty auth cookie in login response")
	}
	if !strings.HasPrefix(authCookie.Value, secureCookieVersion+".") {
		t.Fatalf("expected sealed auth cookie with %q prefix, got %q", secureCookieVersion+".", authCookie.Value)
	}

	claims := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(authCookie.Value, claims); err == nil {
		t.Fatal("expected sealed auth cookie value not to be directly parseable as JWT")
	}
}

func TestAuthMiddlewareAcceptsLegacyJWTAuthCookieFallback(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "legacy-auth-cookie@example.com", "StrongPass1", true)
	legacyToken := buildLegacyJWTForUser(t, user)

	request := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	request.Header.Set("Cookie", authCookieName+"="+legacyToken)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("dashboard request with legacy jwt cookie failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 with legacy jwt cookie fallback, got %d", response.StatusCode)
	}
}

func buildLegacyJWTForUser(t *testing.T, user models.User) string {
	t.Helper()

	now := time.Now()
	claims := authClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("test-secret-key"))
	if err != nil {
		t.Fatalf("sign legacy jwt token: %v", err)
	}
	return signed
}
