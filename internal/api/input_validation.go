package api

import (
	"errors"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

func parseCredentials(c *fiber.Ctx) (credentialsInput, error) {
	credentials := credentialsInput{}
	if err := c.BodyParser(&credentials); err != nil {
		return credentialsInput{}, err
	}

	credentials.Email = strings.ToLower(strings.TrimSpace(credentials.Email))
	credentials.Password = strings.TrimSpace(credentials.Password)
	credentials.ConfirmPassword = strings.TrimSpace(credentials.ConfirmPassword)
	credentials.RememberMe = credentials.RememberMe || parseBoolValue(c.FormValue("remember_me"))

	if credentials.Email == "" || credentials.Password == "" {
		return credentialsInput{}, errors.New("missing credentials")
	}
	if _, err := mail.ParseAddress(credentials.Email); err != nil {
		return credentialsInput{}, errors.New("invalid email")
	}

	return credentials, nil
}

func isValidOnboardingCycleLength(value int) bool {
	return value >= 15 && value <= 90
}

func isValidOnboardingPeriodLength(value int) bool {
	return value >= 1 && value <= 10
}

func validatePasswordStrength(password string) error {
	if !passwordLengthRegex.MatchString(password) {
		return errors.New("password too short")
	}

	if passwordUpperRegex.MatchString(password) &&
		passwordLowerRegex.MatchString(password) &&
		passwordDigitRegex.MatchString(password) {
		return nil
	}
	return errors.New("weak password")
}

func parseDayPayload(c *fiber.Ctx) (dayPayload, error) {
	payload := dayPayload{Flow: models.FlowNone, SymptomIDs: []uint{}}
	contentType := strings.ToLower(c.Get("Content-Type"))

	if strings.Contains(contentType, "application/json") {
		if err := c.BodyParser(&payload); err != nil {
			return payload, err
		}
	} else {
		payload.IsPeriod = parseBoolValue(c.FormValue("is_period"))
		payload.Flow = strings.ToLower(strings.TrimSpace(c.FormValue("flow")))
		payload.Notes = strings.TrimSpace(c.FormValue("notes"))

		symptomRaw := c.Context().PostArgs().PeekMulti("symptom_ids")
		for _, value := range symptomRaw {
			parsed, err := strconv.ParseUint(string(value), 10, 64)
			if err == nil {
				payload.SymptomIDs = append(payload.SymptomIDs, uint(parsed))
			}
		}
	}

	payload.Flow = strings.ToLower(strings.TrimSpace(payload.Flow))
	if payload.Flow == "" {
		payload.Flow = models.FlowNone
	}
	payload.Notes = strings.TrimSpace(payload.Notes)

	return payload, nil
}

func isValidFlow(flow string) bool {
	switch flow {
	case models.FlowNone, models.FlowLight, models.FlowMedium, models.FlowHeavy:
		return true
	default:
		return false
	}
}

func parseBoolValue(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return normalized == "1" || normalized == "true" || normalized == "on" || normalized == "yes"
}

func parseDayParam(raw string, location *time.Location) (time.Time, error) {
	if raw == "" {
		return time.Time{}, errors.New("date is required")
	}
	parsed, err := time.ParseInLocation("2006-01-02", raw, location)
	if err != nil {
		return time.Time{}, err
	}
	return dateAtLocation(parsed, location), nil
}

func parseMonthQuery(raw string, now time.Time, location *time.Location) (time.Time, error) {
	if raw == "" {
		current := dateAtLocation(now, location)
		return time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, location), nil
	}
	parsed, err := time.ParseInLocation("2006-01", raw, location)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, location), nil
}
