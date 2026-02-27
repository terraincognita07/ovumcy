package services

import (
	"errors"

	"github.com/terraincognita07/ovumcy/internal/models"
)

const MaxDayNotesLength = 2000

var ErrInvalidDayFlow = errors.New("invalid day flow")

func NormalizeDayEntryInput(input DayEntryInput) (DayEntryInput, error) {
	if !IsValidDayFlow(input.Flow) {
		return input, ErrInvalidDayFlow
	}
	if !input.IsPeriod {
		input.Flow = models.FlowNone
		input.SymptomIDs = []uint{}
	}
	input.Notes = TrimDayNotes(input.Notes)
	return input, nil
}

func IsValidDayFlow(flow string) bool {
	switch flow {
	case models.FlowNone, models.FlowLight, models.FlowMedium, models.FlowHeavy:
		return true
	default:
		return false
	}
}

func TrimDayNotes(value string) string {
	if len(value) <= MaxDayNotesLength {
		return value
	}
	return value[:MaxDayNotesLength]
}
