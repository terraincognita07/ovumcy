package api

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func dayHasData(entry models.DailyLog) bool {
	return services.DayHasData(entry)
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
	return services.DateAtLocation(value, location)
}

func dayRange(value time.Time, location *time.Location) (time.Time, time.Time) {
	return services.DayRange(value, location)
}

func symptomIDSet(ids []uint) map[uint]bool {
	set := make(map[uint]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}
