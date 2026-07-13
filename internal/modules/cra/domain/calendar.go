package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DefaultWeekStartDay = 1 // Monday (time.Weekday / JS getDay)

// WeekRangeStart returns the UTC date of the first day of week N in the month grid.
func WeekRangeStart(year int, month time.Month, weekNumber WeekNumber, weekStartDay int) time.Time {
	if weekStartDay < 0 || weekStartDay > 6 {
		weekStartDay = DefaultWeekStartDay
	}
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	for int(start.Weekday()) != weekStartDay {
		start = start.AddDate(0, 0, -1)
	}
	offset := int(weekNumber) - 1
	if offset > 0 {
		start = start.AddDate(0, 0, offset*7)
	}
	return start
}

// MonthWeekCount returns how many week tabs are needed for a month.
func MonthWeekCount(year int, month time.Month, weekStartDay int) int {
	if weekStartDay < 0 || weekStartDay > 6 {
		weekStartDay = DefaultWeekStartDay
	}
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	start := WeekRangeStart(year, month, 1, weekStartDay)
	count := 0
	for {
		hasDay := false
		for i := 0; i < 7; i++ {
			d := start.AddDate(0, 0, i)
			if d.Month() == month && d.Day() <= lastDay {
				hasDay = true
				break
			}
		}
		if !hasDay {
			break
		}
		count++
		start = start.AddDate(0, 0, 7)
		if count >= 6 {
			break
		}
	}
	if count == 0 {
		return 1
	}
	return count
}

// WeekDaysInMonth returns in-month dates (YYYY-MM-DD) for weekNumber (up to 7 days).
func WeekDaysInMonth(month Month, weekNumber WeekNumber, weekStartDay int) ([]string, error) {
	year, m, err := parseMonthParts(month)
	if err != nil {
		return nil, err
	}
	start := WeekRangeStart(year, m, weekNumber, weekStartDay)
	days := make([]string, 0, 7)
	for i := 0; i < 7; i++ {
		d := start.AddDate(0, 0, i)
		if d.Month() != m {
			continue
		}
		days = append(days, formatDateUTC(d))
	}
	return days, nil
}

// MonthWeekNumber maps a day to the week tab number within the month grid.
func MonthWeekNumber(day time.Time, month Month, weekStartDay int) (WeekNumber, error) {
	year, m, err := parseMonthParts(month)
	if err != nil {
		return 0, err
	}
	if day.Month() != m || day.Year() != year {
		return 0, fmt.Errorf("day %s outside month %s", day.Format("2006-01-02"), month)
	}
	count := MonthWeekCount(year, m, weekStartDay)
	for w := WeekNumber(1); w <= WeekNumber(count); w++ {
		days, err := WeekDaysInMonth(month, w, weekStartDay)
		if err != nil {
			return 0, err
		}
		key := day.Format("2006-01-02")
		for _, d := range days {
			if d == key {
				return w, nil
			}
		}
	}
	return WeekNumber(1), nil
}

func parseMonthParts(month Month) (int, time.Month, error) {
	parts := strings.Split(string(month), "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid month format")
	}
	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	monthNum, err := strconv.Atoi(parts[1])
	if err != nil || monthNum < 1 || monthNum > 12 {
		return 0, 0, fmt.Errorf("invalid month")
	}
	return year, time.Month(monthNum), nil
}

func formatDateUTC(d time.Time) string {
	return d.Format("2006-01-02")
}
