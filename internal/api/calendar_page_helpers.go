package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

func (handler *Handler) buildCalendarViewData(user *models.User, language string, messages map[string]string, now time.Time, monthStart time.Time, selectedDate string) (fiber.Map, string, error) {
	logRangeStart, logRangeEnd := calendarLogRange(monthStart)
	logs, err := handler.fetchLogsForUser(user.ID, logRangeStart, logRangeEnd)
	if err != nil {
		return nil, "failed to load calendar", err
	}

	stats, _, err := handler.buildCycleStatsForRange(user, now.AddDate(-2, 0, 0), now, now)
	if err != nil {
		return nil, "failed to load stats", err
	}

	days := handler.buildCalendarDays(monthStart, logs, stats, now)
	prevMonth, nextMonth := calendarAdjacentMonthValues(monthStart)

	data := fiber.Map{
		"Title":        localizedPageTitle(messages, "meta.title.calendar", "Lume | Calendar"),
		"CurrentUser":  user,
		"MonthLabel":   localizedMonthYear(language, monthStart),
		"MonthValue":   monthStart.Format("2006-01"),
		"PrevMonth":    prevMonth,
		"NextMonth":    nextMonth,
		"SelectedDate": selectedDate,
		"CalendarDays": days,
		"Today":        dateAtLocation(now, handler.location).Format("2006-01-02"),
		"Stats":        stats,
		"IsOwner":      isOwnerUser(user),
	}
	return data, "", nil
}
