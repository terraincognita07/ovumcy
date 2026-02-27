package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
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
	lengths := services.CompletedCycleTrendLengths(logs, now, handler.location)
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

func (handler *Handler) buildStatsPageData(user *models.User, language string, messages map[string]string, now time.Time) (fiber.Map, string, error) {
	stats, logs, err := handler.buildCycleStatsForRange(user, now.AddDate(-2, 0, 0), now, now)
	if err != nil {
		return nil, "failed to load stats", err
	}

	chartPayload, baselineCycleLength, trendPointCount := handler.buildStatsTrendView(user, logs, now, messages)
	observedCycleCount := len(services.CycleLengths(logs))
	hasReliableTrend := trendPointCount >= 3
	today := dateAtLocation(now, handler.location)
	cycleDayReference := services.DashboardCycleReferenceLength(user, stats)
	cycleStaleAnchor := services.DashboardCycleStaleAnchor(user, stats, handler.location)
	cycleDataStale := services.DashboardCycleDataLooksStale(cycleStaleAnchor, today, cycleDayReference)
	symptomCounts, symptomErrorMessage, err := handler.buildStatsSymptomCounts(user, language)
	if err != nil {
		return nil, symptomErrorMessage, err
	}

	data := fiber.Map{
		"Title":                localizedPageTitle(messages, "meta.title.stats", "Ovumcy | Stats"),
		"CurrentUser":          user,
		"Stats":                stats,
		"ChartData":            chartPayload,
		"ChartBaseline":        baselineCycleLength,
		"TrendPointCount":      trendPointCount,
		"HasObservedCycleData": observedCycleCount > 0,
		"HasTrendData":         trendPointCount > 0,
		"HasReliableTrend":     hasReliableTrend,
		"CycleDataStale":       cycleDataStale,
		"SymptomCounts":        symptomCounts,
		"IsOwner":              isOwnerUser(user),
	}
	return data, "", nil
}
