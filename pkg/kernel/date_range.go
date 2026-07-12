package kernel

import (
	"fmt"
	"time"
)

type DateRange struct {
	From time.Time
	To   time.Time
}

func NewDateRange(from, to time.Time) (DateRange, error) {
	if to.Before(from) {
		return DateRange{}, fmt.Errorf("date range to before from")
	}
	return DateRange{From: from, To: to}, nil
}

func (d DateRange) Overlaps(other DateRange) bool {
	return !d.To.Before(other.From) && !other.To.Before(d.From)
}
