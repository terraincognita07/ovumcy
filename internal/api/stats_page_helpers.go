package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

const maxStatsTrendPoints = 12

func trimTrailingCycleTrendLengths(lengths []int, maxPoints int) []int {
	if maxPoints <= 0 || len(lengths) <= maxPoints {
		return lengths
	}
	return lengths[len(lengths)-maxPoints:]
}

func buildStatsChartData(messages map[string]string, lengths []int, baselineCycleLength int) fiber.Map {
	chartPayload := fiber.Map{
		"labels": buildCycleTrendLabels(messages, len(lengths)),
		"values": lengths,
	}
	if baselineCycleLength > 0 {
		chartPayload["baseline"] = baselineCycleLength
	}
	return chartPayload
}

func (handler *Handler) buildStatsTrendView(user *models.User, logs []models.DailyLog, now time.Time, messages map[string]string) (fiber.Map, int, int) {
	lengths := handler.completedCycleTrendLengths(logs, now)
	lengths = trimTrailingCycleTrendLengths(lengths, maxStatsTrendPoints)

	baselineCycleLength := ownerBaselineCycleLength(user)
	chartPayload := buildStatsChartData(messages, lengths, baselineCycleLength)
	return chartPayload, baselineCycleLength, len(lengths)
}

func (handler *Handler) buildStatsSymptomCounts(user *models.User, language string) ([]SymptomCount, string, error) {
	if !isOwnerUser(user) {
		return []SymptomCount{}, "", nil
	}

	symptomLogs, err := handler.fetchAllLogsForUser(user.ID)
	if err != nil {
		return nil, "failed to load symptom logs", err
	}

	symptomCounts, err := handler.calculateSymptomFrequencies(user.ID, symptomLogs)
	if err != nil {
		return nil, "failed to load symptom stats", err
	}
	localizeSymptomFrequencySummaries(language, symptomCounts)
	return symptomCounts, "", nil
}
