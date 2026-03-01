package services

import (
	"errors"
	"testing"
)

func TestNormalizeAuthEmail(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "normalizes case and spaces", raw: " USER@EXAMPLE.COM ", want: "user@example.com"},
		{name: "invalid email returns empty", raw: "not-email", want: ""},
		{name: "empty returns empty", raw: "   ", want: ""},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if got := NormalizeAuthEmail(testCase.raw); got != testCase.want {
				t.Fatalf("NormalizeAuthEmail(%q) = %q, want %q", testCase.raw, got, testCase.want)
			}
		})
	}
}

func TestNormalizeCredentialsInput(t *testing.T) {
	email, password, err := NormalizeCredentialsInput(" USER@EXAMPLE.COM ", "  StrongPass1  ")
	if err != nil {
		t.Fatalf("expected valid credentials input, got %v", err)
	}
	if email != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", email)
	}
	if password != "StrongPass1" {
		t.Fatalf("expected trimmed password, got %q", password)
	}

	_, _, err = NormalizeCredentialsInput("not-email", "StrongPass1")
	if !errors.Is(err, ErrAuthCredentialsInvalid) {
		t.Fatalf("expected ErrAuthCredentialsInvalid for invalid email, got %v", err)
	}

	_, _, err = NormalizeCredentialsInput("user@example.com", " ")
	if !errors.Is(err, ErrAuthCredentialsInvalid) {
		t.Fatalf("expected ErrAuthCredentialsInvalid for empty password, got %v", err)
	}
}

func TestValidateRecoveryCodeFormat(t *testing.T) {
	if err := ValidateRecoveryCodeFormat("OVUM-ABCD-2345-EFGH"); err != nil {
		t.Fatalf("expected valid recovery code format, got %v", err)
	}
	if err := ValidateRecoveryCodeFormat("OVUM-INVALID"); !errors.Is(err, ErrAuthRecoveryCodeInvalid) {
		t.Fatalf("expected ErrAuthRecoveryCodeInvalid, got %v", err)
	}
}
