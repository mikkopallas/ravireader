package db

import (
	"os"
	"testing"

	"github.com/mikcompany/ravireader/veikkausapi"
)

func TestDBInitAndSaveFetch(t *testing.T) {
	// Use in-memory db for testing
	os.Setenv("DUCKDB_QUACK_URL", ":memory:")
	defer os.Unsetenv("DUCKDB_QUACK_URL")

	err := Init(false)
	if err != nil {
		t.Fatalf("Failed to initialize db: %v", err)
	}

	// 1. Test SaveCards
	cards := []veikkausapi.Card{
		{CardID: 1, Country: "FI", TrackName: "Vermo", MeetDate: "2026-05-21"},
		{CardID: 2, Country: "SE", TrackName: "Solvalla", MeetDate: "2026-05-21"}, // Non-FI shouldn't be fetched later
	}
	err = SaveCards(cards)
	if err != nil {
		t.Fatalf("SaveCards failed: %v", err)
	}

	// 2. Test SaveRaces
	races := []veikkausapi.Race{
		{RaceID: 101, RaceNumber: 1, SeriesSpecification: "Kylmäveriset", CancelledRace: false},
		{RaceID: 102, RaceNumber: 2, SeriesSpecification: "Lämminveriset", CancelledRace: false},
	}
	err = SaveRaces(1, races)
	if err != nil {
		t.Fatalf("SaveRaces failed: %v", err)
	}

	// 3. Test SaveResult
	result101 := veikkausapi.Result{
		RaceID: 101, CardID: 1, RaceNumber: 1, Distance: 2120, StartType: "Auto", StartTime: 123456789,
		Runners: []veikkausapi.Runner{
			{Position: 1, StartNumber: 3, HorseName: "Nopea", DriverFirstName: "Matti", DriverLastName: "Meikäläinen", Distance: 2120, StartTrack: 3, KmTime: "1.15,0"},
		},
	}
	err = SaveResult(result101)
	if err != nil {
		t.Fatalf("SaveResult failed: %v", err)
	}

	// 4. Test FetchAllData (should only return FI cards)
	crrData, err := FetchAllData()
	if err != nil {
		t.Fatalf("FetchAllData failed: %v", err)
	}

	if len(crrData) != 1 {
		t.Fatalf("Expected 1 FI card, got %d", len(crrData))
	}

	if crrData[0].Card.TrackName != "Vermo" {
		t.Errorf("Expected track Vermo, got %s", crrData[0].Card.TrackName)
	}

	if len(crrData[0].Races) != 2 {
		t.Errorf("Expected 2 races, got %d", len(crrData[0].Races))
	}

	if len(crrData[0].Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(crrData[0].Results))
	}

	if len(crrData[0].Results[0].Runners) != 1 {
		t.Errorf("Expected 1 runner in result, got %d", len(crrData[0].Results[0].Runners))
	}
	
	if crrData[0].Results[0].Runners[0].HorseName != "Nopea" {
		t.Errorf("Expected horse Nopea, got %s", crrData[0].Results[0].Runners[0].HorseName)
	}
}
