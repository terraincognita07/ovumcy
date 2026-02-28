package services

import (
	"errors"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var recoveryCodePattern = regexp.MustCompile(`^OVUM-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$`)

func TestBuildAndParsePasswordResetToken(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Date(2026, time.March, 1, 10, 0, 0, 0, time.UTC)
	passwordHash := "$2a$10$testhashvaluefortokenclaims"

	token, err := BuildPasswordResetToken(secret, 42, passwordHash, 30*time.Minute, now)
	if err != nil {
		t.Fatalf("BuildPasswordResetToken() unexpected error: %v", err)
	}

	claims, err := ParsePasswordResetToken(secret, token, now.Add(1*time.Minute))
	if err != nil {
		t.Fatalf("ParsePasswordResetToken() unexpected error: %v", err)
	}
	if claims.UserID != 42 {
		t.Fatalf("expected UserID=42, got %d", claims.UserID)
	}
	if claims.Purpose != passwordResetTokenPurpose {
		t.Fatalf("expected purpose %q, got %q", passwordResetTokenPurpose, claims.Purpose)
	}
	if claims.PasswordState == "" {
		t.Fatalf("expected non-empty password state")
	}
}

func TestParsePasswordResetTokenRejectsExpired(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Date(2026, time.March, 1, 10, 0, 0, 0, time.UTC)
	passwordHash := "$2a$10$testhashvaluefortokenclaims"

	token, err := BuildPasswordResetToken(secret, 42, passwordHash, 1*time.Minute, now)
	if err != nil {
		t.Fatalf("BuildPasswordResetToken() unexpected error: %v", err)
	}

	_, err = ParsePasswordResetToken(secret, token, now.Add(2*time.Minute))
	if !errors.Is(err, ErrPasswordResetTokenExpired) {
		t.Fatalf("expected ErrPasswordResetTokenExpired, got %v", err)
	}
}

func TestParsePasswordResetTokenRejectsWrongPurpose(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Date(2026, time.March, 1, 10, 0, 0, 0, time.UTC)

	claims := PasswordResetClaims{
		UserID:        7,
		Purpose:       "another-purpose",
		PasswordState: "state",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(7, 10),
			ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	_, err = ParsePasswordResetToken(secret, signed, now.Add(1*time.Minute))
	if !errors.Is(err, ErrPasswordResetTokenInvalidPurpose) {
		t.Fatalf("expected ErrPasswordResetTokenInvalidPurpose, got %v", err)
	}
}

func TestPasswordStateFingerprintMatch(t *testing.T) {
	hash := "$2a$10$testhashvaluefortokenclaims"
	fingerprint := PasswordStateFingerprint(hash)
	if fingerprint == "" {
		t.Fatalf("expected non-empty fingerprint")
	}
	if !IsPasswordStateFingerprintMatch(fingerprint, hash) {
		t.Fatalf("expected fingerprint match")
	}
	if IsPasswordStateFingerprintMatch(fingerprint, "another-hash") {
		t.Fatalf("expected fingerprint mismatch")
	}
}

func TestGenerateRecoveryCodeHash(t *testing.T) {
	code, hash, err := GenerateRecoveryCodeHash()
	if err != nil {
		t.Fatalf("GenerateRecoveryCodeHash() unexpected error: %v", err)
	}
	if !recoveryCodePattern.MatchString(code) {
		t.Fatalf("expected recovery code format, got %q", code)
	}
	if hash == "" {
		t.Fatalf("expected non-empty recovery hash")
	}
}
