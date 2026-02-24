package api

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
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
	if periodLength <= 0 {
		periodLength = models.DefaultPeriodLength
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
		predictedPeriodLength := int(stats.AveragePeriodLength + 0.5)
		if predictedPeriodLength <= 0 {
			predictedPeriodLength = periodLength
		}
		ovulationDate, fertilityWindowStart, fertilityWindowEnd, ovulationExact, ovulationCalculable := services.PredictCycleWindow(
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
			stats.OvulationDate = dateAtLocation(ovulationDate, handler.location)
			stats.OvulationExact = ovulationExact
			stats.OvulationImpossible = false
			if fertilityWindowStart.IsZero() {
				stats.FertilityWindowStart = time.Time{}
			} else {
				stats.FertilityWindowStart = dateAtLocation(fertilityWindowStart, handler.location)
			}
			if fertilityWindowEnd.IsZero() {
				stats.FertilityWindowEnd = time.Time{}
			} else {
				stats.FertilityWindowEnd = dateAtLocation(fertilityWindowEnd, handler.location)
			}
		}
	}

	today := dateAtLocation(now.In(handler.location), handler.location)
	if _, projectedCycleDay, projectionOK := projectCycleStart(stats.LastPeriodStart, cycleLength, today); projectionOK {
		stats.CurrentCycleDay = projectedCycleDay
	} else if !stats.LastPeriodStart.IsZero() && !today.Before(stats.LastPeriodStart) {
		stats.CurrentCycleDay = int(today.Sub(stats.LastPeriodStart).Hours()/24) + 1
	} else {
		stats.CurrentCycleDay = 0
	}

	stats.CurrentPhase = handler.detectCurrentPhase(stats, logs, today)

	return stats
}

func projectCycleStart(lastPeriodStart time.Time, cycleLength int, today time.Time) (time.Time, int, bool) {
	if lastPeriodStart.IsZero() || cycleLength <= 0 {
		return time.Time{}, 0, false
	}
	if today.Before(lastPeriodStart) {
		return lastPeriodStart, 0, true
	}

	elapsedDays := int(today.Sub(lastPeriodStart).Hours() / 24)
	cyclesElapsed := elapsedDays / cycleLength
	projectedStart := dateAtLocation(lastPeriodStart.AddDate(0, 0, cyclesElapsed*cycleLength), today.Location())
	projectedCycleDay := (elapsedDays % cycleLength) + 1
	return projectedStart, projectedCycleDay, true
}

func shiftCycleStartToFutureOvulation(cycleStart time.Time, ovulationDate time.Time, cycleLength int, today time.Time) time.Time {
	if cycleLength <= 0 || !ovulationDate.Before(today) {
		return cycleStart
	}
	lagDays := int(today.Sub(ovulationDate).Hours() / 24)
	shiftCycles := lagDays/cycleLength + 1
	return dateAtLocation(cycleStart.AddDate(0, 0, shiftCycles*cycleLength), today.Location())
}
