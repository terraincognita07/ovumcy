package api

import (
	"sort"
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
)

func (handler *Handler) buildCalendarDays(monthStart time.Time, logs []models.DailyLog, stats services.CycleStats, now time.Time) []CalendarDay {
	monthEnd := monthStart.AddDate(0, 1, -1)
	gridStart := monthStart.AddDate(0, 0, -int(monthStart.Weekday()))
	gridEnd := monthEnd.AddDate(0, 0, 6-int(monthEnd.Weekday()))

	latestLogByDate := make(map[string]models.DailyLog)
	hasDataMap := make(map[string]bool)
	for _, logEntry := range logs {
		key := dateAtLocation(logEntry.Date, handler.location).Format("2006-01-02")
		existing, exists := latestLogByDate[key]
		if !exists || logEntry.Date.After(existing.Date) || (logEntry.Date.Equal(existing.Date) && logEntry.ID > existing.ID) {
			latestLogByDate[key] = logEntry
		}
		hasDataMap[key] = hasDataMap[key] || dayHasData(logEntry)
	}

	predictedPeriodMap := make(map[string]bool)
	predictedPeriodLength := int(stats.AveragePeriodLength + 0.5)
	if predictedPeriodLength <= 0 {
		predictedPeriodLength = 5
	}
	if !stats.NextPeriodStart.IsZero() {
		for offset := 0; offset < predictedPeriodLength; offset++ {
			day := stats.NextPeriodStart.AddDate(0, 0, offset)
			predictedPeriodMap[day.Format("2006-01-02")] = true
		}
	}

	fertilityMap := make(map[string]bool)
	if !stats.FertilityWindowStart.IsZero() && !stats.FertilityWindowEnd.IsZero() {
		for day := stats.FertilityWindowStart; !day.After(stats.FertilityWindowEnd); day = day.AddDate(0, 0, 1) {
			fertilityMap[day.Format("2006-01-02")] = true
		}
	}

	todayKey := dateAtLocation(now, handler.location).Format("2006-01-02")
	ovulationKey := stats.OvulationDate.Format("2006-01-02")

	days := make([]CalendarDay, 0, 42)
	for day := gridStart; !day.After(gridEnd); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		inMonth := day.Month() == monthStart.Month()
		entry, hasEntry := latestLogByDate[key]
		isPeriod := hasEntry && entry.IsPeriod
		isPredicted := predictedPeriodMap[key]
		isFertility := fertilityMap[key]
		isToday := key == todayKey
		isOvulation := key == ovulationKey
		hasData := hasDataMap[key]

		cellClass := "calendar-cell"
		textClass := "calendar-day-number"
		badgeClass := "calendar-tag"
		if isPeriod {
			cellClass += " calendar-cell-period"
			badgeClass += " calendar-tag-period"
		} else if isPredicted {
			cellClass += " calendar-cell-predicted"
			badgeClass += " calendar-tag-predicted"
		} else if isFertility {
			cellClass += " calendar-cell-fertile"
			badgeClass += " calendar-tag-fertile"
		}
		if !inMonth {
			cellClass += " calendar-cell-out"
			textClass += " calendar-day-out"
		}
		if isToday {
			cellClass += " calendar-cell-today"
		}

		days = append(days, CalendarDay{
			Date:         day,
			DateString:   key,
			Day:          day.Day(),
			InMonth:      inMonth,
			IsToday:      isToday,
			IsPeriod:     isPeriod,
			IsPredicted:  isPredicted,
			IsFertility:  isFertility,
			IsOvulation:  isOvulation,
			HasData:      hasData,
			CellClass:    cellClass,
			TextClass:    textClass,
			BadgeClass:   badgeClass,
			OvulationDot: isOvulation,
		})
	}
	return days
}

func (handler *Handler) calculateSymptomFrequencies(userID uint, logs []models.DailyLog) ([]SymptomCount, error) {
	if len(logs) == 0 {
		return []SymptomCount{}, nil
	}
	totalDays := len(logs)

	counts := make(map[uint]int)
	for _, logEntry := range logs {
		for _, id := range logEntry.SymptomIDs {
			counts[id]++
		}
	}
	if len(counts) == 0 {
		return []SymptomCount{}, nil
	}

	symptoms, err := handler.fetchSymptoms(userID)
	if err != nil {
		return nil, err
	}

	symptomByID := make(map[uint]models.SymptomType, len(symptoms))
	for _, symptom := range symptoms {
		symptomByID[symptom.ID] = symptom
	}

	result := make([]SymptomCount, 0, len(counts))
	for id, count := range counts {
		if symptom, ok := symptomByID[id]; ok {
			result = append(result, SymptomCount{Name: symptom.Name, Icon: symptom.Icon, Count: count, TotalDays: totalDays})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Name < result[j].Name
		}
		return result[i].Count > result[j].Count
	})

	return result, nil
}

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
