package api

import (
	"github.com/terraincognita07/ovumcy/internal/services"
)

const recoveryCodePrefix = "OVUM"

func normalizeRecoveryCode(raw string) string {
	return services.NormalizeRecoveryCode(raw)
}

func generateRecoveryCodeHash() (string, string, error) {
	return services.GenerateRecoveryCodeHash()
}
