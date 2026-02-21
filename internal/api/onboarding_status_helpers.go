package api

import "strings"

func normalizeOnboardingPeriodStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case onboardingPeriodStatusOngoing:
		return onboardingPeriodStatusOngoing
	case onboardingPeriodStatusFinished:
		return onboardingPeriodStatusFinished
	default:
		return ""
	}
}
