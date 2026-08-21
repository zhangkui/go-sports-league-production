package verify

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/goxm2/sports-league/internal/models"
)

func TestBug014_BusinessRegression(t *testing.T) {
	runBug014BusinessRegression(t)
}

func runBug014BusinessRegression(t *testing.T) {
	f := newFixture(t)
	seasonID := f.seedSeason("bug014", models.SeasonStatusRegistration)
	fromTeam := f.seedTeam(seasonID, "FROM14")
	toTeam := f.seedTeam(seasonID, "TO14")
	playerID := f.seedPlayer(seasonID, fromTeam, "Transfer Player", 7)
	transfer, err := f.teams.CreateTransfer(f.ctx, models.CreateTransferRequest{PlayerID: playerID, ToTeamID: toTeam, Reason: "move"}, 1)
	if err != nil {
		t.Fatalf("create transfer: %v", err)
	}

	start := make(chan struct{})
	var successes atomic.Int32
	var wg sync.WaitGroup
	for reviewerID := int64(1); reviewerID <= 2; reviewerID++ {
		wg.Add(1)
		go func(reviewer int64) {
			defer wg.Done()
			<-start
			if _, reviewErr := f.teams.ReviewTransfer(f.ctx, transfer.ID, reviewer, models.TransferReview{Status: models.TransferStatusApproved}); reviewErr == nil {
				successes.Add(1)
			}
		}(reviewerID)
	}
	close(start)
	wg.Wait()

	if successes.Load() != 1 {
		t.Errorf("successful concurrent transfer approvals = %d, want 1", successes.Load())
	}
	var currentTeam int64
	if err := f.db.QueryRow(`SELECT team_id FROM players WHERE id=?`, playerID).Scan(&currentTeam); err != nil {
		t.Fatalf("read player team: %v", err)
	}
	if currentTeam != toTeam {
		t.Errorf("concurrent approval left player on team %d, want destination %d", currentTeam, toTeam)
	}
}
