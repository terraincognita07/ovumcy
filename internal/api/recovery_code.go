package api

import (
	"fmt"
	"strings"

	"github.com/terraincognita07/ovumcy/internal/services"
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
	return services.GenerateRecoveryCodeHash()
}
