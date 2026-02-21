package api

import (
	"strings"
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func dayHasData(entry models.DailyLog) bool {
	if entry.IsPeriod {
		return true
	}
	if len(entry.SymptomIDs) > 0 {
		return true
	}
	if strings.TrimSpace(entry.Notes) != "" {
		return true
	}
	return strings.TrimSpace(entry.Flow) != "" && entry.Flow != models.FlowNone
}

func sameCalendarDay(a time.Time, b time.Time) bool {
	return a.Format("2006-01-02") == b.Format("2006-01-02")
}

func betweenCalendarDaysInclusive(day time.Time, start time.Time, end time.Time) bool {
	if start.IsZero() || end.IsZero() {
		return false
	}
	return (day.Equal(start) || day.After(start)) && (day.Equal(end) || day.Before(end))
}

func sanitizeLogForPartner(entry models.DailyLog) models.DailyLog {
	entry.Notes = ""
	entry.SymptomIDs = []uint{}
	return entry
}

func dateAtLocation(value time.Time, location *time.Location) time.Time {
	localized := value.In(location)
	year, month, day := localized.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, location)
}

func dayRange(value time.Time, location *time.Location) (time.Time, time.Time) {
	start := dateAtLocation(value, location)
	return start, start.AddDate(0, 0, 1)
}

func dayStorageKey(value time.Time, location *time.Location) string {
	return dateAtLocation(value, location).Format("2006-01-02")
}

func nextDayStorageKey(value time.Time, location *time.Location) string {
	return dateAtLocation(value, location).AddDate(0, 0, 1).Format("2006-01-02")
}

func symptomIDSet(ids []uint) map[uint]bool {
	set := make(map[uint]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}

func removeUint(values []uint, needle uint) []uint {
	filtered := make([]uint, 0, len(values))
	for _, value := range values {
		if value != needle {
			filtered = append(filtered, value)
		}
	}
	return filtered
}
