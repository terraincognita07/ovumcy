package api

import (
	"fmt"
	"strings"
	"time"
)

var monthNames = map[string][]string{
	"en": {"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"},
	"ru": {"Январь", "Февраль", "Март", "Апрель", "Май", "Июнь", "Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь"},
}

var monthLongNames = map[string][]string{
	"en": {"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"},
	"ru": {"января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря"},
}

var weekdayShortNames = map[string][]string{
	"en": {"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
	"ru": {"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"},
}

var weekdayLongNames = map[string][]string{
	"en": {"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"},
	"ru": {"воскресенье", "понедельник", "вторник", "среда", "четверг", "пятница", "суббота"},
}

var monthShortNames = map[string][]string{
	"en": {"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
	"ru": {"Янв", "Фев", "Мар", "Апр", "Май", "Июн", "Июл", "Авг", "Сен", "Окт", "Ноя", "Дек"},
}

func localizedMonthYear(language string, value time.Time) string {
	names, ok := monthNames[strings.ToLower(language)]
	if !ok || len(names) < 12 {
		return value.Format("January 2006")
	}
	monthIndex := int(value.Month()) - 1
	if monthIndex < 0 || monthIndex >= len(names) {
		return value.Format("January 2006")
	}
	return fmt.Sprintf("%s %d", names[monthIndex], value.Year())
}

func localizedDateLabel(language string, value time.Time) string {
	lang := strings.ToLower(strings.TrimSpace(language))
	weekdays, weekdaysOK := weekdayShortNames[lang]
	months, monthsOK := monthShortNames[lang]
	if !weekdaysOK || !monthsOK {
		return value.Format("Mon, Jan 2")
	}
	monthIndex := int(value.Month()) - 1
	if monthIndex < 0 || monthIndex >= len(months) {
		return value.Format("Mon, Jan 2")
	}

	weekday := weekdays[int(value.Weekday())]
	month := months[monthIndex]
	return fmt.Sprintf("%s, %s %d", weekday, month, value.Day())
}

func localizedDashboardDate(language string, value time.Time) string {
	lang := strings.ToLower(strings.TrimSpace(language))
	weekdays, weekdaysOK := weekdayLongNames[lang]
	months, monthsOK := monthLongNames[lang]
	if !weekdaysOK || !monthsOK {
		return value.Format("January 2, 2006, Monday")
	}
	monthIndex := int(value.Month()) - 1
	if monthIndex < 0 || monthIndex >= len(months) {
		return value.Format("January 2, 2006, Monday")
	}

	weekday := weekdays[int(value.Weekday())]
	month := months[monthIndex]
	if lang == "ru" {
		return fmt.Sprintf("%d %s %d, %s", value.Day(), month, value.Year(), weekday)
	}
	return fmt.Sprintf("%s %d, %d, %s", month, value.Day(), value.Year(), weekday)
}
