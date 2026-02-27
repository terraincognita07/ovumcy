package api

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestRecoveryCodeCookieIsNotPlaintextJSON(t *testing.T) {
	app, _ := newOnboardingTestApp(t)
	_, recoveryCookie := registerAndExtractRecoveryCookies(
		t,
		app,
		"recovery-cookie-encoding@example.com",
		"StrongPass1",
	)
	if recoveryCookie == "" {
		t.Fatal("expected recovery cookie in register response")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(recoveryCookie)
	if err == nil {
		payload := recoveryCodePagePayload{}
		if json.Unmarshal(decoded, &payload) == nil {
			t.Fatalf("expected recovery cookie to be sealed; got plaintext payload: %#v", payload)
		}
	}

	if strings.Contains(recoveryCookie, "OVUM-") {
		t.Fatalf("expected recovery cookie value not to expose plaintext recovery code")
	}
}
