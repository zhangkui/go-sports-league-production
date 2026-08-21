package verify

import (
	"testing"

	"github.com/goxm2/sports-league/internal/models"
)

func TestBug015_BusinessRegression(t *testing.T) {
	runBug015BusinessRegression(t)
}

func runBug015BusinessRegression(t *testing.T) {
	f := newFixture(t)
	venueID := f.seedVenue("Availability Venue")
	slots := []models.VenueAvailabilityInput{
		{Weekday: 1, StartTime: "09:00", EndTime: "12:00"},
		{Weekday: 3, StartTime: "18:00", EndTime: "21:00"},
	}
	if err := f.venues.SetAvailability(f.ctx, venueID, slots); err != nil {
		t.Fatalf("set availability: %v", err)
	}
	rows, err := f.venues.ListAvailability(f.ctx, venueID)
	if err != nil {
		t.Fatalf("list availability: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("availability row count = %d, want 2", len(rows))
	}
	if len(rows) == 2 && (rows[0].Weekday == rows[1].Weekday || rows[0].StartTime == rows[1].StartTime) {
		t.Errorf("distinct availability windows collapsed to %#v", rows)
	}
}
