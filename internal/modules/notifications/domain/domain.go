package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrRuleNotFound     = errors.New("rule not found")
	ErrInvalidFrequency = errors.New("invalid frequency")
	ErrNoRecipients     = errors.New("no recipients")
)

type Frequency string

const (
	FrequencyImmediate         Frequency = "immediate"
	FrequencyMorning           Frequency = "morning"
	FrequencyMonday            Frequency = "monday"
	FrequencyFriday            Frequency = "friday"
	FrequencyLastMondayOfMonth Frequency = "last_monday_of_month"
)

func ParseFrequency(raw string) (Frequency, error) {
	f := Frequency(strings.ToLower(strings.TrimSpace(raw)))
	switch f {
	case FrequencyImmediate, FrequencyMorning, FrequencyMonday, FrequencyFriday, FrequencyLastMondayOfMonth:
		return f, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidFrequency, raw)
	}
}

func (f Frequency) IsImmediate() bool {
	return f == FrequencyImmediate
}

const scheduledSendHour = 8

// NextRun computes the next dispatch time for a non-immediate frequency,
// strictly after now. Sends are scheduled at 08:00 local time.
func NextRun(freq Frequency, now time.Time) time.Time {
	switch freq {
	case FrequencyImmediate:
		return now
	case FrequencyMorning:
		candidate := atHour(now, scheduledSendHour)
		if !candidate.After(now) {
			candidate = candidate.AddDate(0, 0, 1)
		}
		return candidate
	case FrequencyMonday:
		return nextWeekday(now, time.Monday)
	case FrequencyFriday:
		return nextWeekday(now, time.Friday)
	case FrequencyLastMondayOfMonth:
		return nextLastMondayOfMonth(now)
	default:
		return now
	}
}

func atHour(t time.Time, hour int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location())
}

func nextWeekday(now time.Time, target time.Weekday) time.Time {
	candidate := atHour(now, scheduledSendHour)
	for i := 0; i < 8; i++ {
		if candidate.Weekday() == target && candidate.After(now) {
			return candidate
		}
		candidate = candidate.AddDate(0, 0, 1)
	}
	return candidate
}

func nextLastMondayOfMonth(now time.Time) time.Time {
	candidate := lastMondayOfMonth(now.Year(), now.Month(), now.Location())
	if !candidate.After(now) {
		year, month := now.Year(), now.Month()
		if month == time.December {
			year++
			month = time.January
		} else {
			month++
		}
		candidate = lastMondayOfMonth(year, month, now.Location())
	}
	return candidate
}

func lastMondayOfMonth(year int, month time.Month, loc *time.Location) time.Time {
	lastDay := time.Date(year, month, 1, scheduledSendHour, 0, 0, 0, loc).AddDate(0, 1, -1)
	offset := (int(lastDay.Weekday()) - int(time.Monday) + 7) % 7
	return lastDay.AddDate(0, 0, -offset)
}

type Channel string

const (
	ChannelEmail        Channel = "email"
	ChannelEmailWithPDF Channel = "email_with_pdf"
)

type RecipientPolicy struct {
	UserIDs       []uuid.UUID `json:"userIds,omitempty"`
	ServiceID     *uuid.UUID  `json:"serviceId,omitempty"`
	EquipeID      *uuid.UUID  `json:"equipeId,omitempty"`
	ApplicationID *uuid.UUID  `json:"applicationId,omitempty"`
}

type NotificationRule struct {
	ID               uuid.UUID
	TenantID         kernel.TenantID
	Code             string
	Trigger          string
	Frequency        Frequency
	RecipientsPolicy RecipientPolicy
	Template         string
	AttachPDF        bool
}

type MessageStatus string

const (
	MessageStatusPending MessageStatus = "pending"
	MessageStatusSent    MessageStatus = "sent"
	MessageStatusFailed  MessageStatus = "failed"
)

type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Content     []byte `json:"content"`
}

type NotificationMessage struct {
	ID           uuid.UUID
	TenantID     kernel.TenantID
	RuleCode     string
	Recipients   []string
	Subject      string
	Body         string
	Attachments  []Attachment
	Status       MessageStatus
	Attempts     int
	SentAt       *time.Time
	ScheduledFor *time.Time
}

func ApplyTemplate(tpl string, vars map[string]string) string {
	out := tpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}

func DefaultSignature(companyName, tenantURL string) string {
	lines := []string{"Cordialement"}
	if companyName != "" {
		lines = append(lines, companyName)
	}
	if tenantURL != "" {
		lines = append(lines, tenantURL)
	}
	return strings.Join(lines, "\n")
}

func WithSignature(body, signature string) string {
	body = strings.TrimSpace(body)
	signature = strings.TrimSpace(signature)
	if signature == "" {
		return body
	}
	if body == "" {
		return signature
	}
	return body + "\n\n" + signature
}
