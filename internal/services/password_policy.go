package services

import (
	"errors"
	"unicode"
)

var ErrWeakPassword = errors.New("weak password")

func ValidatePasswordStrength(password string) error {
	if len([]rune(password)) < 8 {
		return ErrWeakPassword
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if hasUpper && hasLower && hasDigit {
		return nil
	}
	return ErrWeakPassword
}
