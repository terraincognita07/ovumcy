package api

import (
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
)

func (handler *Handler) applyUserCycleBaseline(user *models.User, logs []models.DailyLog, stats services.CycleStats, now time.Time) services.CycleStats {
	if user == nil || user.Role != models.RoleOwner {
		return stats
	}

	latestLoggedPeriodStart := time.Time{}
	detectedStarts := services.DetectCycleStarts(logs)
	if len(detectedStarts) > 0 {
		latestLoggedPeriodStart = dateAtLocation(detectedStarts[len(detectedStarts)-1], handler.location)
	}

	cycleLength := 0
	if isValidOnboardingCycleLength(user.CycleLength) {
		cycleLength = user.CycleLength
	}

	periodLength := 0
	if isValidOnboardingPeriodLength(user.PeriodLength) {
		periodLength = user.PeriodLength
	}

	reliableCycleData := len(services.CycleLengths(logs)) >= 2
	if !reliableCycleData {
		if cycleLength > 0 {
			stats.AverageCycleLength = float64(cycleLength)
			stats.MedianCycleLength = cycleLength
		}
		if periodLength > 0 {
			stats.AveragePeriodLength = float64(periodLength)
		}
		switch {
		case !latestLoggedPeriodStart.IsZero():
			stats.LastPeriodStart = latestLoggedPeriodStart
		case user.LastPeriodStart != nil:
			stats.LastPeriodStart = dateAtLocation(*user.LastPeriodStart, handler.location)
		}
	} else if !latestLoggedPeriodStart.IsZero() {
		stats.LastPeriodStart = latestLoggedPeriodStart
	}

	if !stats.LastPeriodStart.IsZero() && cycleLength > 0 && (!reliableCycleData || stats.NextPeriodStart.IsZero()) {
		stats.NextPeriodStart = dateAtLocation(stats.LastPeriodStart.AddDate(0, 0, cycleLength), handler.location)
		stats.OvulationDate = dateAtLocation(stats.NextPeriodStart.AddDate(0, 0, -handler.lutealPhaseDays), handler.location)
		stats.FertilityWindowStart = dateAtLocation(stats.OvulationDate.AddDate(0, 0, -5), handler.location)
		stats.FertilityWindowEnd = dateAtLocation(stats.OvulationDate.AddDate(0, 0, 1), handler.location)
	}

	today := dateAtLocation(now.In(handler.location), handler.location)
	if !stats.LastPeriodStart.IsZero() && !today.Before(stats.LastPeriodStart) {
		stats.CurrentCycleDay = int(today.Sub(stats.LastPeriodStart).Hours()/24) + 1
	} else {
		stats.CurrentCycleDay = 0
	}

	stats.CurrentPhase = handler.detectCurrentPhase(stats, logs, today)

	return stats
}

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
