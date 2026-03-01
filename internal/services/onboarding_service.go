package services

import (
	"errors"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

var (
	ErrOnboardingStepsRequired       = errors.New("complete onboarding steps first")
	ErrOnboardingStartDateRequired   = errors.New("onboarding start date is required")
	ErrOnboardingStartDateOutOfRange = errors.New("onboarding start date out of range")
)

type OnboardingUserRepository interface {
	FindByID(userID uint) (models.User, error)
	SaveOnboardingStep1(userID uint, start time.Time) error
	SaveOnboardingStep2(userID uint, cycleLength int, periodLength int, autoPeriodFill bool) error
	CompleteOnboarding(userID uint, startDay time.Time, periodLength int) error
}

type OnboardingService struct {
	users OnboardingUserRepository
}

func NewOnboardingService(users OnboardingUserRepository) *OnboardingService {
	return &OnboardingService{users: users}
}

func (service *OnboardingService) SaveStep1(userID uint, valuesStart time.Time) error {
	return service.users.SaveOnboardingStep1(userID, valuesStart)
}

func (service *OnboardingService) ValidateStep1StartDate(start time.Time, now time.Time, location *time.Location) error {
	if start.IsZero() {
		return ErrOnboardingStartDateRequired
	}

	minDate, maxDate := OnboardingDateBounds(now, location)
	day := DateAtLocation(start, location)
	if day.Before(minDate) || day.After(maxDate) {
		return ErrOnboardingStartDateOutOfRange
	}

	return nil
}

func (service *OnboardingService) SaveStep2(userID uint, cycleLength int, periodLength int, autoPeriodFill bool) (int, int, error) {
	safeCycleLength, safePeriodLength := SanitizeOnboardingCycleAndPeriod(cycleLength, periodLength)
	if err := service.users.SaveOnboardingStep2(userID, safeCycleLength, safePeriodLength, autoPeriodFill); err != nil {
		return 0, 0, err
	}
	return safeCycleLength, safePeriodLength, nil
}

func (service *OnboardingService) CompleteOnboardingForUser(userID uint, location *time.Location) (time.Time, error) {
	current, err := service.users.FindByID(userID)
	if err != nil {
		return time.Time{}, err
	}
	if current.LastPeriodStart == nil {
		return time.Time{}, ErrOnboardingStepsRequired
	}

	startDay := DateAtLocation(*current.LastPeriodStart, location)
	_, periodLength := SanitizeOnboardingCycleAndPeriod(current.CycleLength, current.PeriodLength)
	if err := service.users.CompleteOnboarding(userID, startDay, periodLength); err != nil {
		return time.Time{}, err
	}
	return startDay, nil
}

func SanitizeOnboardingCycleAndPeriod(cycleLength int, periodLength int) (int, int) {
	safeCycleLength := ClampOnboardingCycleLength(cycleLength)
	safePeriodLength := ClampOnboardingPeriodLength(periodLength)

	if safeCycleLength-safePeriodLength < 8 {
		safePeriodLength = safeCycleLength - 8
		if safePeriodLength < 1 {
			safePeriodLength = 1
		}
	}

	return safeCycleLength, safePeriodLength
}

func ClampOnboardingCycleLength(value int) int {
	if value < 15 {
		return 15
	}
	if value > 90 {
		return 90
	}
	return value
}

func ClampOnboardingPeriodLength(value int) int {
	if value < 1 {
		return 1
	}
	if value > 14 {
		return 14
	}
	return value
}

func IsValidOnboardingCycleLength(value int) bool {
	return value >= 15 && value <= 90
}

func IsValidOnboardingPeriodLength(value int) bool {
	return value >= 1 && value <= 14
}

func ResolveCycleAndPeriodDefaults(cycleLength int, periodLength int) (int, int) {
	resolvedCycleLength := cycleLength
	if !IsValidOnboardingCycleLength(resolvedCycleLength) {
		resolvedCycleLength = models.DefaultCycleLength
	}

	resolvedPeriodLength := periodLength
	if !IsValidOnboardingPeriodLength(resolvedPeriodLength) {
		resolvedPeriodLength = models.DefaultPeriodLength
	}

	return resolvedCycleLength, resolvedPeriodLength
}

func OnboardingDateBounds(now time.Time, location *time.Location) (time.Time, time.Time) {
	if location == nil {
		location = time.UTC
	}

	today := DateAtLocation(now.In(location), location)
	minDate := time.Date(today.Year(), time.January, 1, 0, 0, 0, 0, location)
	sixtyDaysAgo := today.AddDate(0, 0, -60)
	if sixtyDaysAgo.After(minDate) {
		minDate = sixtyDaysAgo
	}
	return minDate, today
}
