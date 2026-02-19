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
