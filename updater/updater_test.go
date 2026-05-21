package updater

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/mikcompany/ravireader/db"
	"github.com/mikcompany/ravireader/veikkausapi"
)

func TestUpdater(t *testing.T) {
	// 1. Initialize in-memory DB
	os.Setenv("DUCKDB_QUACK_URL", ":memory:")
	defer os.Unsetenv("DUCKDB_QUACK_URL")
	err := db.Init(false)
	if err != nil {
		t.Fatalf("Failed to initialize db: %v", err)
	}

	// 2. Mock Veikkaus API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/toto-info/v1/cards/today":
			w.Write([]byte(`{"collection": [{"cardId": 1, "country": "FI", "trackName": "MockTrack", "meetDate": "2026-05-21"}]}`))
		case "/api/toto-info/v1/card/1/results":
			w.Write([]byte(`{"collection": [{"raceId": 101, "raceNumber": 1, "seriesSpecification": "MockSeries", "cancelledRace": false}]}`))
		case "/api/toto-info/v1/race/101/competition-results":
			w.Write([]byte(`{"raceId": 101, "cardId": 1, "raceNumber": 1, "distance": 2100, "startType": "Auto", "startTime": 12345, "runners": [{"position": 1, "startNumber": 1, "horseName": "MockHorse", "driverFirstName": "A", "driverLastName": "B", "distance": 2100, "startTrack": 1, "kmTime": "1.15,0"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Override BaseURL
	originalBaseURL := veikkausapi.BaseURL
	veikkausapi.BaseURL = server.URL
	defer func() { veikkausapi.BaseURL = originalBaseURL }()

	// 3. Run updater once
	// Since Start runs an infinite loop if we call it with a ticker, we just call update() directly.
	update()

	// Wait a tiny bit just in case, though update is synchronous.
	time.Sleep(100 * time.Millisecond)

	// 4. Verify data in DB
	data, err := db.FetchAllData()
	if err != nil {
		t.Fatalf("Failed to fetch data: %v", err)
	}

	if len(data) != 1 {
		t.Fatalf("Expected 1 card, got %d", len(data))
	}
	if data[0].Card.TrackName != "MockTrack" {
		t.Errorf("Expected track MockTrack, got %s", data[0].Card.TrackName)
	}
	if len(data[0].Races) != 1 {
		t.Errorf("Expected 1 race, got %d", len(data[0].Races))
	}
	if len(data[0].Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(data[0].Results))
	}
	if len(data[0].Results[0].Runners) != 1 {
		t.Errorf("Expected 1 runner, got %d", len(data[0].Results[0].Runners))
	}
	if data[0].Results[0].Runners[0].HorseName != "MockHorse" {
		t.Errorf("Expected horse MockHorse, got %s", data[0].Results[0].Runners[0].HorseName)
	}
}
