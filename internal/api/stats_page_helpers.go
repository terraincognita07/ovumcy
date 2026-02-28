package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
)

const maxStatsTrendPoints = 12

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
	handler.ensureDependencies()
	lengths, baselineCycleLength := handler.statsService.BuildTrend(user, logs, now, handler.location, maxStatsTrendPoints)
	chartPayload := buildStatsChartData(messages, lengths, baselineCycleLength)
	return chartPayload, baselineCycleLength, len(lengths)
}

func (handler *Handler) buildStatsSymptomCounts(user *models.User, language string) ([]SymptomCount, string, error) {
	handler.ensureDependencies()
	frequencies, err := handler.statsService.BuildSymptomFrequenciesForUser(user)
	if err != nil {
		return nil, "failed to load symptom stats", err
	}

	symptomCounts := make([]SymptomCount, 0, len(frequencies))
	for _, item := range frequencies {
		symptomCounts = append(symptomCounts, SymptomCount{
			Name:      item.Name,
			Icon:      item.Icon,
			Count:     item.Count,
			TotalDays: item.TotalDays,
		})
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
	handler.ensureDependencies()
	flags := handler.statsService.BuildFlags(user, logs, stats, now, handler.location, trendPointCount)
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
		"HasObservedCycleData": flags.HasObservedCycleData,
		"HasTrendData":         flags.HasTrendData,
		"HasReliableTrend":     flags.HasReliableTrend,
		"CycleDataStale":       flags.CycleDataStale,
		"SymptomCounts":        symptomCounts,
		"IsOwner":              isOwnerUser(user),
	}
	return data, "", nil
}
