package services

import (
	"errors"
	"testing"
)

func TestValidatePasswordStrength_RejectsWeakPasswords(t *testing.T) {
	testCases := []string{
		"Short1",
		"alllowercase1",
		"ALLUPPERCASE1",
		"NoDigitsHere",
	}

	for _, password := range testCases {
		if err := ValidatePasswordStrength(password); !errors.Is(err, ErrWeakPassword) {
			t.Fatalf("expected ErrWeakPassword for %q, got %v", password, err)
		}
	}
}

func TestValidatePasswordStrength_AcceptsStrongPassword(t *testing.T) {
	if err := ValidatePasswordStrength("StrongPass1"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
