package services

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestValidatePasswordChangeRejectsInvalidInput(t *testing.T) {
	service := NewSettingsService(nil)

	err := service.ValidatePasswordChange("hash", " ", "NewPass1", "NewPass1")
	if !errors.Is(err, ErrSettingsPasswordChangeInvalidInput) {
		t.Fatalf("expected ErrSettingsPasswordChangeInvalidInput, got %v", err)
	}
}

func TestValidatePasswordChangeRejectsMismatch(t *testing.T) {
	service := NewSettingsService(nil)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("StrongPass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	err = service.ValidatePasswordChange(string(passwordHash), "StrongPass1", "NewPass1", "OtherPass1")
	if !errors.Is(err, ErrSettingsPasswordMismatch) {
		t.Fatalf("expected ErrSettingsPasswordMismatch, got %v", err)
	}
}

func TestValidatePasswordChangeRejectsInvalidCurrentPassword(t *testing.T) {
	service := NewSettingsService(nil)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("StrongPass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	err = service.ValidatePasswordChange(string(passwordHash), "WrongPass1", "NewPass1", "NewPass1")
	if !errors.Is(err, ErrSettingsInvalidCurrentPassword) {
		t.Fatalf("expected ErrSettingsInvalidCurrentPassword, got %v", err)
	}
}

func TestValidatePasswordChangeRejectsUnchangedPassword(t *testing.T) {
	service := NewSettingsService(nil)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("StrongPass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	err = service.ValidatePasswordChange(string(passwordHash), "StrongPass1", "StrongPass1", "StrongPass1")
	if !errors.Is(err, ErrSettingsNewPasswordMustDiffer) {
		t.Fatalf("expected ErrSettingsNewPasswordMustDiffer, got %v", err)
	}
}

func TestValidatePasswordChangeRejectsWeakPassword(t *testing.T) {
	service := NewSettingsService(nil)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("StrongPass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	err = service.ValidatePasswordChange(string(passwordHash), "StrongPass1", "12345678", "12345678")
	if !errors.Is(err, ErrSettingsWeakPassword) {
		t.Fatalf("expected ErrSettingsWeakPassword, got %v", err)
	}
}

func TestValidatePasswordChangeAcceptsValidInput(t *testing.T) {
	service := NewSettingsService(nil)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("StrongPass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if err := service.ValidatePasswordChange(string(passwordHash), "StrongPass1", "EvenStronger2", "EvenStronger2"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
