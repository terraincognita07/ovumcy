package services

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrSettingsPasswordChangeInvalidInput = errors.New("settings password change invalid input")
	ErrSettingsPasswordMismatch           = errors.New("settings password mismatch")
	ErrSettingsInvalidCurrentPassword     = errors.New("settings invalid current password")
	ErrSettingsNewPasswordMustDiffer      = errors.New("settings new password must differ")
	ErrSettingsWeakPassword               = errors.New("settings weak password")
)

func (service *SettingsService) ValidatePasswordChange(passwordHash string, currentPassword string, newPassword string, confirmPassword string) error {
	currentPassword = strings.TrimSpace(currentPassword)
	newPassword = strings.TrimSpace(newPassword)
	confirmPassword = strings.TrimSpace(confirmPassword)

	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		return ErrSettingsPasswordChangeInvalidInput
	}
	if newPassword != confirmPassword {
		return ErrSettingsPasswordMismatch
	}
	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(currentPassword)) != nil {
		return ErrSettingsInvalidCurrentPassword
	}
	if currentPassword == newPassword {
		return ErrSettingsNewPasswordMustDiffer
	}
	if err := ValidatePasswordStrength(newPassword); err != nil {
		return ErrSettingsWeakPassword
	}
	return nil
}
