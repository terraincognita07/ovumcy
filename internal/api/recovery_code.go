package api

import (
	"fmt"
	"strings"

	"github.com/terraincognita07/ovumcy/internal/security"
	"golang.org/x/crypto/bcrypt"
)

const recoveryCodePrefix = "OVUM"

func normalizeRecoveryCode(raw string) string {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.TrimPrefix(normalized, recoveryCodePrefix)
	if len(normalized) != 12 {
		return strings.ToUpper(strings.TrimSpace(raw))
	}
	return fmt.Sprintf("%s-%s-%s-%s", recoveryCodePrefix, normalized[:4], normalized[4:8], normalized[8:12])
}

func generateRecoveryCodeHash() (string, string, error) {
	code, err := generateRecoveryCode()
	if err != nil {
		return "", "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return code, string(hash), nil
}

func generateRecoveryCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	value, err := security.RandomString(12, alphabet)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s-%s-%s", recoveryCodePrefix, value[:4], value[4:8], value[8:12]), nil
}
