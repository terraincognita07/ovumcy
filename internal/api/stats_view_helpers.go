package api

import (
	"fmt"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func ownerBaselineCycleLength(user *models.User) int {
	if !isOwnerUser(user) || !isValidOnboardingCycleLength(user.CycleLength) {
		return 0
	}
	return user.CycleLength
}

func buildCycleTrendLabels(messages map[string]string, pointCount int) []string {
	if pointCount <= 0 {
		return []string{}
	}

	cycleLabelPattern := translateMessage(messages, "stats.cycle_label")
	if cycleLabelPattern == "stats.cycle_label" {
		cycleLabelPattern = "Cycle %d"
	}

	labels := make([]string, 0, pointCount)
	for index := 0; index < pointCount; index++ {
		labels = append(labels, fmt.Sprintf(cycleLabelPattern, index+1))
	}
	return labels
}

func localizeSymptomFrequencySummaries(language string, counts []SymptomCount) {
	for index := range counts {
		counts[index].FrequencySummary = localizedSymptomFrequencySummary(language, counts[index].Count, counts[index].TotalDays)
	}
}
