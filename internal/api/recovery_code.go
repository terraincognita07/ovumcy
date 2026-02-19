package api

import (
	"crypto/rand"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func normalizeRecoveryCode(raw string) string {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.TrimPrefix(normalized, "LUME")
	if len(normalized) != 12 {
		return strings.ToUpper(strings.TrimSpace(raw))
	}
	return fmt.Sprintf("LUME-%s-%s-%s", normalized[:4], normalized[4:8], normalized[8:12])
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
	randomBytes := make([]byte, 12)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	chars := make([]byte, 12)
	for index, value := range randomBytes {
		chars[index] = alphabet[int(value)%len(alphabet)]
	}

	return fmt.Sprintf("LUME-%s-%s-%s", string(chars[:4]), string(chars[4:8]), string(chars[8:12])), nil
}
