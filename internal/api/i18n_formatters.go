package api

import (
	"fmt"
	"strings"
)

func localizedSymptomFrequencySummary(language string, count int, days int) string {
	lang := strings.ToLower(strings.TrimSpace(language))
	if lang == "ru" {
		return fmt.Sprintf("%d %s (за %d %s)",
			count,
			russianPluralForm(count, "раз", "раза", "раз"),
			days,
			russianPluralForm(days, "день", "дня", "дней"),
		)
	}

	countWord := "times"
	if count == 1 {
		countWord = "time"
	}
	dayWord := "days"
	if days == 1 {
		dayWord = "day"
	}
	return fmt.Sprintf("%d %s (in %d %s)", count, countWord, days, dayWord)
}

func russianPluralForm(value int, one string, few string, many string) string {
	absolute := value
	if absolute < 0 {
		absolute = -absolute
	}
	lastTwoDigits := absolute % 100
	if lastTwoDigits >= 11 && lastTwoDigits <= 14 {
		return many
	}

	lastDigit := absolute % 10
	switch {
	case lastDigit == 1:
		return one
	case lastDigit >= 2 && lastDigit <= 4:
		return few
	default:
		return many
	}
}
