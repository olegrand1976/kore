package domain

import (
	"testing"
	"time"
)

func TestNewLeadRequiresConsent(t *testing.T) {
	_, err := NewLead("a@b.c", "ACME", "10", "demo", "", false, time.Now())
	if err != ErrConsentRequired {
		t.Fatalf("expected consent error, got %v", err)
	}
}

func TestBookingSlotIsBookable(t *testing.T) {
	now := time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC)
	slot := BookingSlot{
		Status:    SlotStatusFree,
		SlotStart: now.Add(time.Hour),
	}
	if !slot.IsBookable(now) {
		t.Fatal("future free slot should be bookable")
	}
	slot.SlotStart = now.Add(-time.Hour)
	if slot.IsBookable(now) {
		t.Fatal("past slot must not be bookable")
	}
}
