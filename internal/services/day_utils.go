package services

import (
	"strings"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func DateAtLocation(value time.Time, location *time.Location) time.Time {
	if location == nil {
		location = time.UTC
	}
	localized := value.In(location)
	year, month, day := localized.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, location)
}

func DayRange(value time.Time, location *time.Location) (time.Time, time.Time) {
	start := DateAtLocation(value, location)
	return start, start.AddDate(0, 0, 1)
}

func DayHasData(entry models.DailyLog) bool {
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

func RemoveUint(values []uint, needle uint) []uint {
	filtered := make([]uint, 0, len(values))
	for _, value := range values {
		if value != needle {
			filtered = append(filtered, value)
		}
	}
	return filtered
}
