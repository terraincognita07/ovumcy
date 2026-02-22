package api

import "github.com/terraincognita07/lume/internal/services"

func isValidOnboardingCycleLength(value int) bool {
	return value >= 15 && value <= 90
}

func isValidOnboardingPeriodLength(value int) bool {
	return value >= 1 && value <= 14
}

func canEstimateOvulation(cycleLength int, periodLength int) bool {
	day, _ := services.CalcOvulationDay(cycleLength, periodLength)
	return day > 0
}
