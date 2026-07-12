package kernel

import (
	"fmt"
	"time"
)

type Period struct {
	Start time.Time
	End   time.Time
}

func NewPeriod(start, end time.Time) (Period, error) {
	if end.Before(start) {
		return Period{}, fmt.Errorf("period end before start")
	}
	return Period{Start: start, End: end}, nil
}

func (p Period) Contains(t time.Time) bool {
	return !t.Before(p.Start) && !t.After(p.End)
}
