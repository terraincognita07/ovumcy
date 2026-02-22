package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
)

func (handler *Handler) buildDashboardViewData(user *models.User, language string, messages map[string]string, now time.Time) (fiber.Map, string, error) {
	today := dateAtLocation(now, handler.location)

	stats, _, err := handler.buildCycleStatsForRange(user, today.AddDate(-2, 0, 0), today, now)
	if err != nil {
		return nil, "failed to load logs", err
	}

	todayLog, symptoms, err := handler.fetchDayLogForViewer(user, today)
	if err != nil {
		return nil, "failed to load today log", err
	}

	cycleDayReference := dashboardCycleReferenceLength(user, stats)
	cycleDayWarning := dashboardCycleDayLooksLong(stats.CurrentCycleDay, cycleDayReference)
	nextPeriodInPast := !stats.NextPeriodStart.IsZero() && stats.NextPeriodStart.Before(today)
	ovulationInPast := !stats.OvulationDate.IsZero() && stats.OvulationDate.Before(today)

	data := fiber.Map{
		"Title":             localizedPageTitle(messages, "meta.title.dashboard", "Lume | Dashboard"),
		"CurrentUser":       user,
		"Stats":             stats,
		"CycleDayReference": cycleDayReference,
		"CycleDayWarning":   cycleDayWarning,
		"NextPeriodInPast":  nextPeriodInPast,
		"OvulationInPast":   ovulationInPast,
		"Today":             today.Format("2006-01-02"),
		"FormattedDate":     localizedDashboardDate(language, today),
		"TodayLog":          todayLog,
		"TodayHasData":      dayHasData(todayLog),
		"Symptoms":          symptoms,
		"SelectedSymptomID": symptomIDSet(todayLog.SymptomIDs),
		"IsOwner":           isOwnerUser(user),
	}
	return data, "", nil
}

func dashboardCycleReferenceLength(user *models.User, stats services.CycleStats) int {
	if user != nil && isValidOnboardingCycleLength(user.CycleLength) {
		return user.CycleLength
	}
	if stats.MedianCycleLength > 0 {
		return stats.MedianCycleLength
	}
	if stats.AverageCycleLength > 0 {
		return int(stats.AverageCycleLength + 0.5)
	}
	return models.DefaultCycleLength
}

func dashboardCycleDayLooksLong(currentDay int, referenceLength int) bool {
	if currentDay <= 0 || referenceLength <= 0 {
		return false
	}
	return currentDay > referenceLength+7
}

func (handler *Handler) buildDayEditorPartialData(user *models.User, language string, messages map[string]string, day time.Time, now time.Time) (fiber.Map, string, error) {
	hasDayData, err := handler.dayHasDataForDate(user.ID, day)
	if err != nil {
		return nil, "failed to load day state", err
	}

	logEntry, symptoms, err := handler.fetchDayLogForViewer(user, day)
	if err != nil {
		return nil, "failed to load day", err
	}

	payload := fiber.Map{
		"Date":              day,
		"DateString":        day.Format("2006-01-02"),
		"DateLabel":         localizedDateLabel(language, day),
		"IsFutureDate":      day.After(dateAtLocation(now.In(handler.location), handler.location)),
		"NoDataLabel":       translateMessage(messages, "common.not_available"),
		"Log":               logEntry,
		"Symptoms":          symptoms,
		"SelectedSymptomID": symptomIDSet(logEntry.SymptomIDs),
		"HasDayData":        hasDayData,
		"IsOwner":           isOwnerUser(user),
	}
	return payload, "", nil
}
