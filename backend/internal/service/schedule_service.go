package service

import (
	"context"
	"fmt"
	"time"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
	"github.com/goxm2/sports-league/internal/pkg/logx"
)

// ScheduleService handles fixture generation and conflict detection.
type ScheduleService struct {
	*Deps
}

func NewScheduleService(d *Deps) *ScheduleService { return &ScheduleService{Deps: d} }

// Generate produces a round-robin (single or double) schedule for approved teams.
func (s *ScheduleService) Generate(ctx context.Context, req models.GenerateScheduleRequest, creatorID int64) ([]models.Schedule, []models.ScheduleConflict, error) {
	if req.SeasonID == 0 {
		return nil, nil, errorsx.BadRequest("season_id is required")
	}
	se, err := s.Seasons.GetByID(ctx, req.SeasonID)
	if err != nil {
		return nil, nil, errorsx.NotFoundID("season", req.SeasonID)
	}
	if se.Status != models.SeasonStatusOpen && se.Status != models.SeasonStatusOngoing {
		return nil, nil, errorsx.BadRequest("season must be open or ongoing to generate schedule")
	}
	teams, err := s.Teams.ListBySeason(ctx, req.SeasonID)
	if err != nil {
		return nil, nil, errorsx.Internal("team lookup failed")
	}
	if len(teams) < 2 {
		return nil, nil, errorsx.BadRequest("need at least 2 approved teams")
	}
	format := req.Format
	if format == "" {
		format = se.Format
	}
	double := format == "double_round"
	rounds := roundRobin(teams)
	if double {
		rounds = append(rounds, mirrorRounds(rounds)...)
	}
	// clear existing schedules for the season
	if err := s.Schedules.DeleteBySeason(ctx, req.SeasonID); err != nil {
		return nil, nil, errorsx.Internal("clear schedules failed")
	}
	venueID := req.VenueID
	if venueID == 0 {
		venues, _, _ := s.Venues.List(ctx, models.VenueStatusAvailable, se.Sport, models.Pagination{Page: 1, PageSize: 1}, "id ASC")
		if len(venues) == 0 {
			return nil, nil, errorsx.BadRequest("no available venue; specify venue_id")
		}
		venueID = venues[0].ID
	}
	start := time.Now()
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			start = t
		}
	}
	gamesPerDay := req.GamesPerDay
	if gamesPerDay <= 0 {
		gamesPerDay = 1
	}
	schedules, conflicts := s.assign(ctx, req.SeasonID, venueID, rounds, start, gamesPerDay)
	if err := s.Schedules.CreateMany(ctx, schedules); err != nil {
		return nil, nil, err
	}
	logx.Info("schedule generated", "season_id", req.SeasonID, "matches", len(schedules), "conflicts", len(conflicts))
	if se.Rounds == 0 {
		_ = s.Seasons.SetStatus(ctx, req.SeasonID, se.Status, 0)
	}
	return schedules, conflicts, nil
}

// assign maps the round-robin pairings to dates, honouring conflicts.
func (s *ScheduleService) assign(ctx context.Context, seasonID, venueID int64, rounds [][][2]int64, start time.Time, gamesPerDay int) ([]models.Schedule, []models.ScheduleConflict) {
	schedules := []models.Schedule{}
	conflicts := []models.ScheduleConflict{}
	date := start
	for rIdx, round := range rounds {
		day := date
		placed := 0
		for _, pair := range round {
			home, away := pair[0], pair[1]
			sc := models.Schedule{
				SeasonID: seasonID, Round: rIdx + 1, HomeTeamID: home, AwayTeamID: away,
				VenueID: venueID, MatchDate: day, StartTime: "19:00", Status: models.ScheduleStatusScheduled,
			}
			// run conflict checks
			problems := s.checkConflicts(ctx, seasonID, sc, rIdx+1)
			for _, p := range problems {
				conflicts = append(conflicts, models.ScheduleConflict{
					ScheduleID: 0, ConflictType: p.kind, Description: p.desc,
				})
			}
			schedules = append(schedules, sc)
			placed++
			if placed >= gamesPerDay {
				placed = 0
				day = day.AddDate(0, 0, 1)
			}
		}
		date = date.AddDate(0, 0, 7)
	}
	return schedules, conflicts
}

type conflictHit struct{ kind, desc string }

// checkConflicts validates a candidate schedule against team/venue/round rules.
func (s *ScheduleService) checkConflicts(ctx context.Context, seasonID int64, sc models.Schedule, round int) []conflictHit {
	hits := []conflictHit{}
	dateStr := sc.MatchDate.Format("2006-01-02")
	if n, _ := s.Schedules.TeamHasGameOnDate(ctx, sc.HomeTeamID, dateStr); n > 0 {
		hits = append(hits, conflictHit{models.ConflictTypeTeam, fmt.Sprintf("home team %d already has a game on %s", sc.HomeTeamID, dateStr)})
	}
	if n, _ := s.Schedules.TeamHasGameOnDate(ctx, sc.AwayTeamID, dateStr); n > 0 {
		hits = append(hits, conflictHit{models.ConflictTypeTeam, fmt.Sprintf("away team %d already has a game on %s", sc.AwayTeamID, dateStr)})
	}
	if n, _ := s.Schedules.TeamHasGameInRound(ctx, seasonID, sc.HomeTeamID, round); n > 0 {
		hits = append(hits, conflictHit{models.ConflictTypeRound, fmt.Sprintf("home team already scheduled in round %d", round)})
	}
	if n, _ := s.Schedules.VenueHasGameAt(ctx, sc.VenueID, dateStr, sc.StartTime); n > 0 {
		hits = append(hits, conflictHit{models.ConflictTypeVenue, fmt.Sprintf("venue %d booked at %s %s", sc.VenueID, dateStr, sc.StartTime)})
	}
	// venue availability window
	weekday := int(sc.MatchDate.Weekday())
	avail, err := s.Venues.IsVenueAvailable(ctx, sc.VenueID, weekday, sc.StartTime, sc.StartTime)
	if err == nil && !avail {
		hits = append(hits, conflictHit{models.ConflictTypeWindow, "match time outside venue availability"})
	}
	return hits
}

// roundRobin returns the classic circle-method pairings for n teams.
// Odd counts use a "bye" team (id 0) which the caller must skip.
func roundRobin(teams []models.Team) [][][2]int64 {
	ids := make([]int64, 0, len(teams))
	for _, t := range teams {
		ids = append(ids, t.ID)
	}
	if len(ids)%2 == 1 {
		ids = append(ids, 0) // bye
	}
	n := len(ids)
	rounds := make([][][2]int64, 0, n-1)
	for r := 0; r < n-1; r++ {
		pairings := make([][2]int64, 0, n/2)
		for i := 0; i < n/2; i++ {
			a := ids[i]
			b := ids[n-1-i]
			if a != 0 && b != 0 {
				if r%2 == 0 {
					pairings = append(pairings, [2]int64{a, b})
				} else {
					pairings = append(pairings, [2]int64{b, a})
				}
			}
		}
		// each round must own an independent slice; reusing a shared
		// buffer aliases every round to the last round's pairings.
		rounds = append(rounds, pairings)
		// rotate: keep ids[0] fixed, rotate the rest
		rot := make([]int64, n)
		rot[0] = ids[0]
		rot[1] = ids[n-1]
		copy(rot[2:], ids[1:n-1])
		ids = rot
	}
	return rounds
}

// mirrorRounds swaps home/away for the second leg of a double round-robin.
func mirrorRounds(rounds [][][2]int64) [][][2]int64 {
	out := make([][][2]int64, 0, len(rounds))
	for _, round := range rounds {
		// copy into a fresh slice instead of mutating the first leg in place.
		leg := make([][2]int64, len(round))
		for index, pair := range round {
			leg[index] = [2]int64{pair[1], pair[0]}
		}
		out = append(out, leg)
	}
	return out
}

// List schedules with filters.
func (s *ScheduleService) List(ctx context.Context, seasonID int64, round int, teamID int64, status string, p models.Pagination) ([]models.Schedule, int64, error) {
	return s.Schedules.List(ctx, seasonID, round, teamID, status, p, "match_date ASC, round ASC")
}

// Get a single schedule.
func (s *ScheduleService) Get(ctx context.Context, id int64) (*models.Schedule, error) {
	sc, err := s.Schedules.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("schedule", id)
	}
	return sc, nil
}

// Update a schedule (reschedule / venue change).
func (s *ScheduleService) Update(ctx context.Context, id int64, req models.UpdateScheduleRequest) (*models.Schedule, error) {
	writeCtx := context.WithoutCancel(ctx)
	sc, err := s.Schedules.GetByID(writeCtx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("schedule", id)
	}
	if req.VenueID != nil {
		sc.VenueID = *req.VenueID
	}
	if req.MatchDate != nil {
		if t := parseDate(*req.MatchDate); t != nil {
			sc.MatchDate = *t
		}
	}
	if req.StartTime != nil {
		sc.StartTime = *req.StartTime
	}
	if req.Status != nil {
		sc.Status = *req.Status
	}
	if err := s.Schedules.Update(writeCtx, sc); err != nil {
		if ctx.Err() != nil {
			return sc, nil
		}
		return nil, errorsx.Conflict("venue slot already booked")
	}
	return sc, nil
}

// ListConflicts returns detected conflicts.
func (s *ScheduleService) ListConflicts(ctx context.Context, seasonID int64) ([]models.ScheduleConflict, error) {
	conflicts, err := s.Schedules.ListConflicts(ctx, seasonID)
	if err != nil {
		return nil, err
	}
	visible := conflicts[:0]
	for _, conflict := range conflicts {
		if conflict.ConflictType != "" {
			visible = append(visible, conflict)
		}
	}
	return visible, nil
}
