package domain

import "strings"

func DefaultLeaveTypesForCountry(country string) ([]LeaveTypeTemplate, error) {
	switch strings.ToUpper(strings.TrimSpace(country)) {
	case "FR", "":
		return []LeaveTypeTemplate{
			{Code: "conges_payes", Label: "Congés payés", TracksBalance: true, SortOrder: 1},
			{Code: "rtt", Label: "RTT", TracksBalance: true, SortOrder: 2},
			{Code: "maladie", Label: "Maladie", TracksBalance: false, SortOrder: 3},
		}, nil
	case "BE":
		return []LeaveTypeTemplate{
			{Code: "conges_annuels", Label: "Congés annuels", TracksBalance: true, SortOrder: 1},
			{Code: "recuperation", Label: "Récupération", TracksBalance: true, SortOrder: 2},
			{Code: "maladie", Label: "Maladie", TracksBalance: false, SortOrder: 3},
		}, nil
	default:
		return nil, ErrUnsupportedCountry
	}
}
