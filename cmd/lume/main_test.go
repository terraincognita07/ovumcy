package main

import "testing"

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
	if secureConfig.CookieName != "lume_csrf" {
		t.Fatalf("expected csrf cookie name lume_csrf, got %q", secureConfig.CookieName)
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
