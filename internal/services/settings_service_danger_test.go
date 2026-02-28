package services

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestValidateDeleteAccountPasswordRejectsMissingPassword(t *testing.T) {
	service := NewSettingsService(nil)

	err := service.ValidateDeleteAccountPassword("ignored", "   ")
	if !errors.Is(err, ErrSettingsPasswordMissing) {
		t.Fatalf("expected ErrSettingsPasswordMissing, got %v", err)
	}
}

func TestValidateDeleteAccountPasswordRejectsInvalidPassword(t *testing.T) {
	service := NewSettingsService(nil)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("StrongPass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	err = service.ValidateDeleteAccountPassword(string(passwordHash), "WrongPass1")
	if !errors.Is(err, ErrSettingsPasswordInvalid) {
		t.Fatalf("expected ErrSettingsPasswordInvalid, got %v", err)
	}
}

func TestValidateDeleteAccountPasswordAcceptsMatchingPassword(t *testing.T) {
	service := NewSettingsService(nil)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("StrongPass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if err := service.ValidateDeleteAccountPassword(string(passwordHash), "  StrongPass1  "); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
