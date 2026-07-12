package kernel

import (
	"fmt"
	"time"
)

type Duration struct {
	Minutes int
}

func NewDuration(minutes int) (Duration, error) {
	if minutes < 0 {
		return Duration{}, fmt.Errorf("duration cannot be negative")
	}
	return Duration{Minutes: minutes}, nil
}

func DurationFromTime(d time.Duration) Duration {
	return Duration{Minutes: int(d.Minutes())}
}

func (d Duration) ToTimeDuration() time.Duration {
	return time.Duration(d.Minutes) * time.Minute
}

func (d Duration) String() string {
	h := d.Minutes / 60
	m := d.Minutes % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
