package calendar

import "time"

// PublicHolidayDates returns fixed and Easter-based public holidays for a country/year.
func PublicHolidayDates(year int, countryCode string) map[time.Time]struct{} {
	if countryCode == "" {
		countryCode = "FR"
	}
	out := make(map[time.Time]struct{})
	switch countryCode {
	case "FR":
		add(out, year, 1, 1)
		add(out, year, 5, 1)
		add(out, year, 5, 8)
		add(out, year, 7, 14)
		add(out, year, 8, 15)
		add(out, year, 11, 1)
		add(out, year, 11, 11)
		add(out, year, 12, 25)
		easter := easterSunday(year)
		addDate(out, easter.AddDate(0, 0, 1))  // Lundi de Pâques
		addDate(out, easter.AddDate(0, 0, 39)) // Ascension
		addDate(out, easter.AddDate(0, 0, 50)) // Lundi de Pentecôte
	default:
	}
	return out
}

func IsPublicHoliday(day time.Time, countryCode string) bool {
	y, m, d := day.Date()
	normalized := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	holidays := PublicHolidayDates(y, countryCode)
	_, ok := holidays[normalized]
	return ok
}

func add(out map[time.Time]struct{}, year, month, day int) {
	addDate(out, time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC))
}

func addDate(out map[time.Time]struct{}, day time.Time) {
	y, m, d := day.Date()
	out[time.Date(y, m, d, 0, 0, 0, 0, time.UTC)] = struct{}{}
}

// easterSunday — algorithme de Meeus/Jones/Butcher (Gregorian).
func easterSunday(year int) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
