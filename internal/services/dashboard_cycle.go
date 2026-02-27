package services

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func DashboardCycleReferenceLength(user *models.User, stats CycleStats) int {
	if user != nil && IsValidOnboardingCycleLength(user.CycleLength) {
		return user.CycleLength
	}
	if stats.MedianCycleLength > 0 {
		return stats.MedianCycleLength
	}
	if stats.AverageCycleLength > 0 {
		return int(stats.AverageCycleLength + 0.5)
	}
	return models.DefaultCycleLength
}

func DashboardCycleDayLooksLong(currentDay int, referenceLength int) bool {
	if currentDay <= 0 || referenceLength <= 0 {
		return false
	}
	return currentDay > referenceLength+7
}

func DashboardCycleDataLooksStale(lastPeriodStart time.Time, today time.Time, referenceLength int) bool {
	if lastPeriodStart.IsZero() || referenceLength <= 0 || today.Before(lastPeriodStart) {
		return false
	}
	rawCycleDay := int(today.Sub(lastPeriodStart).Hours()/24) + 1
	return rawCycleDay > referenceLength
}

func DashboardCycleStaleAnchor(user *models.User, stats CycleStats, location *time.Location) time.Time {
	if user == nil || user.LastPeriodStart == nil || user.LastPeriodStart.IsZero() {
		return stats.LastPeriodStart
	}
	return DateAtLocation(*user.LastPeriodStart, location)
}

func DashboardUpcomingPredictions(stats CycleStats, user *models.User, today time.Time, cycleLength int) (time.Time, time.Time, bool, bool) {
	nextPeriodStart := stats.NextPeriodStart
	ovulationDate := stats.OvulationDate
	ovulationExact := stats.OvulationExact
	ovulationImpossible := stats.OvulationImpossible

	if stats.LastPeriodStart.IsZero() || cycleLength <= 0 {
		return nextPeriodStart, ovulationDate, ovulationExact, ovulationImpossible
	}

	cycleStart, _, projectionOK := ProjectCycleStart(stats.LastPeriodStart, cycleLength, today)
	if !projectionOK {
		return nextPeriodStart, ovulationDate, ovulationExact, ovulationImpossible
	}

	nextPeriodStart = DateAtLocation(cycleStart.AddDate(0, 0, cycleLength), today.Location())
	predictedPeriodLength := DashboardPredictedPeriodLength(user, stats)
	ovulationDate, _, _, ovulationExact, ovulationCalculable := PredictCycleWindow(
		cycleStart,
		cycleLength,
		predictedPeriodLength,
	)
	if ovulationCalculable && ovulationDate.Before(today) {
		cycleStart = ShiftCycleStartToFutureOvulation(cycleStart, ovulationDate, cycleLength, today)
		nextPeriodStart = DateAtLocation(cycleStart.AddDate(0, 0, cycleLength), today.Location())
		ovulationDate, _, _, ovulationExact, ovulationCalculable = PredictCycleWindow(
			cycleStart,
			cycleLength,
			predictedPeriodLength,
		)
	}
	if !ovulationCalculable {
		return nextPeriodStart, time.Time{}, false, true
	}
	return nextPeriodStart, ovulationDate, ovulationExact, false
}

func DashboardPredictedPeriodLength(user *models.User, stats CycleStats) int {
	if user != nil && IsValidOnboardingPeriodLength(user.PeriodLength) {
		return user.PeriodLength
	}
	predictedPeriodLength := int(stats.AveragePeriodLength + 0.5)
	if predictedPeriodLength > 0 {
		return predictedPeriodLength
	}
	return models.DefaultPeriodLength
}

func CompletedCycleTrendLengths(logs []models.DailyLog, now time.Time, location *time.Location) []int {
	starts := DetectCycleStarts(logs)
	if len(starts) < 2 {
		return nil
	}

	today := DateAtLocation(now, location)
	lengths := make([]int, 0, len(starts)-1)
	for index := 1; index < len(starts); index++ {
		previousStart := DateAtLocation(starts[index-1], location)
		currentStart := DateAtLocation(starts[index], location)
		if !currentStart.Before(today) {
			break
		}
		lengths = append(lengths, int(currentStart.Sub(previousStart).Hours()/24))
	}
	return lengths
}
