package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
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
	cycleStaleAnchor := dashboardCycleStaleAnchor(user, stats, handler.location)
	cycleDataStale := dashboardCycleDataLooksStale(cycleStaleAnchor, today, cycleDayReference)
	displayNextPeriodStart, displayOvulationDate, displayOvulationExact, displayOvulationImpossible := dashboardUpcomingPredictions(stats, user, today, cycleDayReference)
	nextPeriodInPast := !displayNextPeriodStart.IsZero() && displayNextPeriodStart.Before(today)
	ovulationInPast := !displayOvulationImpossible && !displayOvulationDate.IsZero() && displayOvulationDate.Before(today)

	data := fiber.Map{
		"Title":                      localizedPageTitle(messages, "meta.title.dashboard", "Ovumcy | Dashboard"),
		"CurrentUser":                user,
		"Stats":                      stats,
		"CycleDayReference":          cycleDayReference,
		"CycleDayWarning":            cycleDayWarning,
		"CycleDataStale":             cycleDataStale,
		"DisplayNextPeriodStart":     displayNextPeriodStart,
		"DisplayOvulationDate":       displayOvulationDate,
		"DisplayOvulationExact":      displayOvulationExact,
		"DisplayOvulationImpossible": displayOvulationImpossible,
		"NextPeriodInPast":           nextPeriodInPast,
		"OvulationInPast":            ovulationInPast,
		"Today":                      today.Format("2006-01-02"),
		"FormattedDate":              localizedDashboardDate(language, today),
		"TodayEntry":                 todayLog,
		"TodayLog":                   todayLog,
		"TodayHasData":               dayHasData(todayLog),
		"Symptoms":                   symptoms,
		"SelectedSymptomID":          symptomIDSet(todayLog.SymptomIDs),
		"IsOwner":                    isOwnerUser(user),
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

func dashboardCycleDataLooksStale(lastPeriodStart time.Time, today time.Time, referenceLength int) bool {
	if lastPeriodStart.IsZero() || referenceLength <= 0 || today.Before(lastPeriodStart) {
		return false
	}
	rawCycleDay := int(today.Sub(lastPeriodStart).Hours()/24) + 1
	return rawCycleDay > referenceLength
}

func dashboardCycleStaleAnchor(user *models.User, stats services.CycleStats, location *time.Location) time.Time {
	if user == nil || user.LastPeriodStart == nil || user.LastPeriodStart.IsZero() {
		return stats.LastPeriodStart
	}
	if location == nil {
		location = time.UTC
	}
	return dateAtLocation(*user.LastPeriodStart, location)
}

func dashboardUpcomingPredictions(stats services.CycleStats, user *models.User, today time.Time, cycleLength int) (time.Time, time.Time, bool, bool) {
	nextPeriodStart := stats.NextPeriodStart
	ovulationDate := stats.OvulationDate
	ovulationExact := stats.OvulationExact
	ovulationImpossible := stats.OvulationImpossible

	if stats.LastPeriodStart.IsZero() || cycleLength <= 0 {
		return nextPeriodStart, ovulationDate, ovulationExact, ovulationImpossible
	}

	cycleStart, _, projectionOK := projectCycleStart(stats.LastPeriodStart, cycleLength, today)
	if !projectionOK {
		return nextPeriodStart, ovulationDate, ovulationExact, ovulationImpossible
	}

	nextPeriodStart = dateAtLocation(cycleStart.AddDate(0, 0, cycleLength), today.Location())
	predictedPeriodLength := dashboardPredictedPeriodLength(user, stats)
	ovulationDate, _, _, ovulationExact, ovulationCalculable := services.PredictCycleWindow(
		cycleStart,
		cycleLength,
		predictedPeriodLength,
	)
	if ovulationCalculable && ovulationDate.Before(today) {
		cycleStart = shiftCycleStartToFutureOvulation(cycleStart, ovulationDate, cycleLength, today)
		nextPeriodStart = dateAtLocation(cycleStart.AddDate(0, 0, cycleLength), today.Location())
		ovulationDate, _, _, ovulationExact, ovulationCalculable = services.PredictCycleWindow(
			cycleStart,
			cycleLength,
			predictedPeriodLength,
		)
	}
	if !ovulationCalculable {
		return nextPeriodStart, time.Time{}, false, true
	}
	return nextPeriodStart, ovulationDate, ovulationExact, false
}

func dashboardPredictedPeriodLength(user *models.User, stats services.CycleStats) int {
	if user != nil && isValidOnboardingPeriodLength(user.PeriodLength) {
		return user.PeriodLength
	}
	predictedPeriodLength := int(stats.AveragePeriodLength + 0.5)
	if predictedPeriodLength > 0 {
		return predictedPeriodLength
	}
	return models.DefaultPeriodLength
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
