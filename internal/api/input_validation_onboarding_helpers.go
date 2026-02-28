package api

import "github.com/terraincognita07/ovumcy/internal/services"

func isValidOnboardingCycleLength(value int) bool {
	return services.IsValidOnboardingCycleLength(value)
}

func isValidOnboardingPeriodLength(value int) bool {
	return services.IsValidOnboardingPeriodLength(value)
}
