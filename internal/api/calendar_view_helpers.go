package api

import (
	"strings"
	"time"
)

func resolveCalendarMonthAndSelectedDate(monthQueryRaw string, selectedDayRaw string, now time.Time, location *time.Location) (time.Time, string, error) {
	monthQuery := strings.TrimSpace(monthQueryRaw)
	activeMonth, err := parseMonthQuery(monthQuery, now, location)
	if err != nil {
		return time.Time{}, "", err
	}

	selectedDate := ""
	selectedDayRaw = strings.TrimSpace(selectedDayRaw)
	if selectedDayRaw != "" {
		if selectedDay, parseErr := parseDayParam(selectedDayRaw, location); parseErr == nil {
			selectedDate = selectedDay.Format("2006-01-02")
			if monthQuery == "" {
				activeMonth = time.Date(selectedDay.Year(), selectedDay.Month(), 1, 0, 0, 0, 0, location)
			}
		}
	}

	return activeMonth, selectedDate, nil
}

func calendarLogRange(monthStart time.Time) (time.Time, time.Time) {
	monthEnd := monthStart.AddDate(0, 1, -1)
	return monthStart.AddDate(0, 0, -70), monthEnd.AddDate(0, 0, 70)
}

func calendarAdjacentMonthValues(monthStart time.Time) (string, string) {
	return monthStart.AddDate(0, -1, 0).Format("2006-01"), monthStart.AddDate(0, 1, 0).Format("2006-01")
}
