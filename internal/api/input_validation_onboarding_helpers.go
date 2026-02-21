package api

func isValidOnboardingCycleLength(value int) bool {
	return value >= 15 && value <= 90
}

func isValidOnboardingPeriodLength(value int) bool {
	return value >= 1 && value <= 10
}
