package api

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func (handler *Handler) detectCurrentPhase(stats services.CycleStats, logs []models.DailyLog, today time.Time) string {
	periodByDate := make(map[string]bool, len(logs))
	for _, logEntry := range logs {
		if logEntry.IsPeriod {
			periodByDate[dateAtLocation(logEntry.Date, handler.location).Format("2006-01-02")] = true
		}
	}
	if periodByDate[today.Format("2006-01-02")] {
		return "menstrual"
	}

	periodLength := int(stats.AveragePeriodLength + 0.5)
	if periodLength <= 0 {
		periodLength = 5
	}
	if !stats.LastPeriodStart.IsZero() {
		periodEnd := dateAtLocation(stats.LastPeriodStart.AddDate(0, 0, periodLength-1), handler.location)
		if betweenCalendarDaysInclusive(today, stats.LastPeriodStart, periodEnd) {
			return "menstrual"
		}
	}

	if stats.OvulationImpossible {
		return "unknown"
	}

	if !stats.OvulationDate.IsZero() {
		switch {
		case sameCalendarDay(today, stats.OvulationDate):
			return "ovulation"
		case betweenCalendarDaysInclusive(today, stats.FertilityWindowStart, stats.FertilityWindowEnd):
			return "fertile"
		case today.Before(stats.OvulationDate):
			return "follicular"
		default:
			return "luteal"
		}
	}

	return "unknown"
}
