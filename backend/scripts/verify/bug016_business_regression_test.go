package verify

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"

	"github.com/goxm2/sports-league/internal/handler"
	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/pagination"
)

func TestBug016_BusinessRegression(t *testing.T) {
	runBug016BusinessRegression(t)
}

func runBug016BusinessRegression(t *testing.T) {
	f := newFixture(t)
	venueID := f.insertID(`INSERT INTO venues (code,name,address,capacity,sport,status) VALUES (?,?,?,?,?,?)`, "BUG016", "Original Venue", "Original Address", 320, "football", models.VenueStatusAvailable)
	pagination.SetMuxVarsHook(mux.Vars)
	h := handler.NewVenueHandler(f.venues)
	venueIDText := strconv.FormatInt(venueID, 10)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/venues/"+venueIDText, bytes.NewBufferString(`{"status":"maintenance"}`))
	req = mux.SetURLVars(req, map[string]string{"id": venueIDText})
	recorder := httptest.NewRecorder()
	h.Update(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("partial venue update status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	var name, address, sport, status string
	var capacity int
	if err := f.db.QueryRow(`SELECT name,COALESCE(address,''),COALESCE(capacity,0),COALESCE(sport,''),status FROM venues WHERE id=?`, venueID).Scan(&name, &address, &capacity, &sport, &status); err != nil {
		t.Fatalf("read venue after partial update: %v", err)
	}
	if name != "Original Venue" || address != "Original Address" || capacity != 320 || sport != "football" {
		t.Errorf("omitted venue fields were overwritten: name=%q address=%q capacity=%d sport=%q", name, address, capacity, sport)
	}
	if status != models.VenueStatusMaintenance {
		t.Errorf("requested status = %q, want maintenance", status)
	}
}
