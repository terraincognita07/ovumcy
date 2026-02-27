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

	cycleDayReference := services.DashboardCycleReferenceLength(user, stats)
	cycleDayWarning := services.DashboardCycleDayLooksLong(stats.CurrentCycleDay, cycleDayReference)
	cycleStaleAnchor := services.DashboardCycleStaleAnchor(user, stats, handler.location)
	cycleDataStale := services.DashboardCycleDataLooksStale(cycleStaleAnchor, today, cycleDayReference)
	displayNextPeriodStart, displayOvulationDate, displayOvulationExact, displayOvulationImpossible := services.DashboardUpcomingPredictions(stats, user, today, cycleDayReference)
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
