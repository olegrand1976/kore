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
	ErrRuleNotFound      = errors.New("rule not found")
	ErrInvalidFrequency  = errors.New("invalid frequency")
	ErrNoRecipients      = errors.New("no recipients")
)

type Frequency string

const (
	FrequencyImmediate        Frequency = "immediate"
	FrequencyMorning          Frequency = "morning"
	FrequencyMonday           Frequency = "monday"
	FrequencyFriday           Frequency = "friday"
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
	ID          uuid.UUID
	TenantID    kernel.TenantID
	RuleCode    string
	Recipients  []string
	Subject     string
	Body        string
	Attachments []Attachment
	Status      MessageStatus
	Attempts    int
	SentAt      *time.Time
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
