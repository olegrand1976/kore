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
	case "MA", "TN", "MG", "MD":
		return []LeaveTypeTemplate{
			{Code: "conges_payes", Label: "Congés payés", TracksBalance: true, SortOrder: 1},
			{Code: "maladie", Label: "Maladie", TracksBalance: false, SortOrder: 2},
			{Code: "conge_exceptionnel", Label: "Congé exceptionnel", TracksBalance: false, SortOrder: 3},
		}, nil
	case "CA":
		return []LeaveTypeTemplate{
			{Code: "conges_annuels", Label: "Congés annuels", TracksBalance: true, SortOrder: 1},
			{Code: "maladie", Label: "Maladie", TracksBalance: false, SortOrder: 2},
			{Code: "personnel", Label: "Congé personnel", TracksBalance: true, SortOrder: 3},
		}, nil
	default:
		return nil, ErrUnsupportedCountry
	}
}

// StandardLeaveTypeCodes returns the union of all country default leave type codes.
func StandardLeaveTypeCodes() map[string]struct{} {
	out := make(map[string]struct{})
	for _, country := range []string{"FR", "BE", "MA", "TN", "MG", "CA"} {
		templates, err := DefaultLeaveTypesForCountry(country)
		if err != nil {
			continue
		}
		for _, tpl := range templates {
			out[tpl.Code] = struct{}{}
		}
	}
	return out
}

// DefaultLeaveTypeCodeSet returns the default leave type codes for a country.
func DefaultLeaveTypeCodeSet(country string) (map[string]struct{}, error) {
	templates, err := DefaultLeaveTypesForCountry(country)
	if err != nil {
		return nil, err
	}
	out := make(map[string]struct{}, len(templates))
	for _, tpl := range templates {
		out[tpl.Code] = struct{}{}
	}
	return out, nil
}
