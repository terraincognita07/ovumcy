package services

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func ApplyUserCycleBaseline(user *models.User, logs []models.DailyLog, stats CycleStats, now time.Time, location *time.Location) CycleStats {
	if user == nil || user.Role != models.RoleOwner {
		return stats
	}
	if location == nil {
		location = time.UTC
	}

	latestLoggedPeriodStart := time.Time{}
	detectedStarts := DetectCycleStarts(logs)
	if len(detectedStarts) > 0 {
		latestLoggedPeriodStart = DateAtLocation(detectedStarts[len(detectedStarts)-1], location)
	}

	cycleLength := 0
	if IsValidOnboardingCycleLength(user.CycleLength) {
		cycleLength = user.CycleLength
	}

	periodLength := 0
	if IsValidOnboardingPeriodLength(user.PeriodLength) {
		periodLength = user.PeriodLength
	}
	if periodLength <= 0 {
		periodLength = models.DefaultPeriodLength
	}

	reliableCycleData := len(CycleLengths(logs)) >= 2
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
			stats.LastPeriodStart = DateAtLocation(*user.LastPeriodStart, location)
		}
	} else if !latestLoggedPeriodStart.IsZero() {
		stats.LastPeriodStart = latestLoggedPeriodStart
	}

	if !stats.LastPeriodStart.IsZero() && cycleLength > 0 && (!reliableCycleData || stats.NextPeriodStart.IsZero()) {
		stats.NextPeriodStart = DateAtLocation(stats.LastPeriodStart.AddDate(0, 0, cycleLength), location)
		predictedPeriodLength := int(stats.AveragePeriodLength + 0.5)
		if predictedPeriodLength <= 0 {
			predictedPeriodLength = periodLength
		}
		ovulationDate, fertilityWindowStart, fertilityWindowEnd, ovulationExact, ovulationCalculable := PredictCycleWindow(
			stats.LastPeriodStart,
			cycleLength,
			predictedPeriodLength,
		)
		if !ovulationCalculable {
			stats.OvulationDate = time.Time{}
			stats.OvulationExact = false
			stats.OvulationImpossible = true
			stats.FertilityWindowStart = time.Time{}
			stats.FertilityWindowEnd = time.Time{}
		} else {
			stats.OvulationDate = DateAtLocation(ovulationDate, location)
			stats.OvulationExact = ovulationExact
			stats.OvulationImpossible = false
			if fertilityWindowStart.IsZero() {
				stats.FertilityWindowStart = time.Time{}
			} else {
				stats.FertilityWindowStart = DateAtLocation(fertilityWindowStart, location)
			}
			if fertilityWindowEnd.IsZero() {
				stats.FertilityWindowEnd = time.Time{}
			} else {
				stats.FertilityWindowEnd = DateAtLocation(fertilityWindowEnd, location)
			}
		}
	}

	today := DateAtLocation(now.In(location), location)
	if _, projectedCycleDay, projectionOK := ProjectCycleStart(stats.LastPeriodStart, cycleLength, today); projectionOK {
		stats.CurrentCycleDay = projectedCycleDay
	} else if !stats.LastPeriodStart.IsZero() && !today.Before(stats.LastPeriodStart) {
		stats.CurrentCycleDay = int(today.Sub(stats.LastPeriodStart).Hours()/24) + 1
	} else {
		stats.CurrentCycleDay = 0
	}

	stats.CurrentPhase = DetectCurrentPhase(stats, logs, today, location)
	return stats
}

func DetectCurrentPhase(stats CycleStats, logs []models.DailyLog, today time.Time, location *time.Location) string {
	if location == nil {
		location = time.UTC
	}
	periodByDate := make(map[string]bool, len(logs))
	for _, logEntry := range logs {
		if logEntry.IsPeriod {
			periodByDate[DateAtLocation(logEntry.Date, location).Format("2006-01-02")] = true
		}
	}
	if periodByDate[today.Format("2006-01-02")] {
		return "menstrual"
	}

	periodLength := int(stats.AveragePeriodLength + 0.5)
	if periodLength <= 0 {
		periodLength = models.DefaultPeriodLength
	}
	if !stats.LastPeriodStart.IsZero() {
		periodEnd := DateAtLocation(stats.LastPeriodStart.AddDate(0, 0, periodLength-1), location)
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

func ProjectCycleStart(lastPeriodStart time.Time, cycleLength int, today time.Time) (time.Time, int, bool) {
	if lastPeriodStart.IsZero() || cycleLength <= 0 {
		return time.Time{}, 0, false
	}
	if today.Before(lastPeriodStart) {
		return lastPeriodStart, 0, true
	}

	elapsedDays := int(today.Sub(lastPeriodStart).Hours() / 24)
	cyclesElapsed := elapsedDays / cycleLength
	projectedStart := DateAtLocation(lastPeriodStart.AddDate(0, 0, cyclesElapsed*cycleLength), today.Location())
	projectedCycleDay := (elapsedDays % cycleLength) + 1
	return projectedStart, projectedCycleDay, true
}

func ShiftCycleStartToFutureOvulation(cycleStart time.Time, ovulationDate time.Time, cycleLength int, today time.Time) time.Time {
	if cycleLength <= 0 || !ovulationDate.Before(today) {
		return cycleStart
	}
	lagDays := int(today.Sub(ovulationDate).Hours() / 24)
	shiftCycles := lagDays/cycleLength + 1
	return DateAtLocation(cycleStart.AddDate(0, 0, shiftCycles*cycleLength), today.Location())
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
