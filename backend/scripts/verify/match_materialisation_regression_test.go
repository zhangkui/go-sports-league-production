package verify

import (
	"sync"
	"testing"

	"github.com/goxm2/sports-league/internal/models"
)

func runBug001BusinessRegression(t *testing.T) {
	f := newFixture(t)
	seasonID := f.seedSeason("bug001", models.SeasonStatusOpen)
	home := f.seedTeam(seasonID, "M001H")
	away := f.seedTeam(seasonID, "M001A")
	venueID := f.seedVenue("Materialise Venue")
	scheduleID := f.insertID(`INSERT INTO schedules (season_id,round,home_team_id,away_team_id,venue_id,match_date,start_time,status) VALUES (?,?,?,?,?,?,?,?)`, seasonID, 1, home, away, venueID, "2026-09-01", "10:00", models.ScheduleStatusScheduled)

	results := make([]*models.Match, 2)
	errs := make([]error, 2)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range results {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			results[index], errs[index] = f.matches.GetOrMaterialise(f.ctx, scheduleID)
		}(i)
	}
	close(start)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("concurrent materialisation %d returned error: %v", i, err)
		}
		if results[i] == nil || results[i].ID == 0 {
			t.Errorf("concurrent materialisation %d returned an unpersisted match: %#v", i, results[i])
		}
	}
	if results[0] != nil && results[1] != nil && results[0].ID != results[1].ID {
		t.Errorf("concurrent callers observed different match ids: %d and %d", results[0].ID, results[1].ID)
	}
	if count := f.count(`SELECT COUNT(*) FROM matches WHERE schedule_id=?`, scheduleID); count != 1 {
		t.Errorf("materialised match rows for one schedule = %d, want 1", count)
	}
	if count := f.count(`SELECT COUNT(*) FROM matches WHERE schedule_id=? AND id=0`, scheduleID); count != 0 {
		t.Errorf("found %d zero-id materialised rows", count)
	}
}
