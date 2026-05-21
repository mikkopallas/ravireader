package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/mikcompany/ravireader/veikkausapi"
)

var DB *sql.DB
var dbMutex sync.Mutex

// Init initializes the DuckDB connection and creates schemas.
func Init(serveQuack bool) error {
	var err error

	dsn := os.Getenv("DUCKDB_QUACK_URL")
	if dsn == "" {
		dsn = "data.duckdb"
	}

	DB, err = sql.Open("duckdb", dsn)
	if err != nil {
		return fmt.Errorf("failed to open duckdb: %w", err)
	}

	if dsn == "data.duckdb" {
		if serveQuack {
			// Try to load quack extension and start server in background, 
			// errors are just logged as this might fail if extension is missing locally.
			log.Println("Starting Quack server on localhost:9494 with token 'ravitoken'...")
			go func() {
				_, err := DB.Exec("INSTALL quack; LOAD quack; CALL quack_serve('quack:localhost:9494', token := 'ravitoken');")
				if err != nil {
					log.Printf("Failed to start quack server (you may need to install the extension manually): %v\n", err)
				}
			}()
		}
	}

	// Create schemas for local files or memory DBs
	if dsn == "data.duckdb" || dsn == ":memory:" {
		err = createSchemas()
		if err != nil {
			return fmt.Errorf("failed to create schemas: %w", err)
		}
	}

	return nil
}

func createSchemas() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS cards (
			card_id BIGINT PRIMARY KEY,
			country VARCHAR,
			track_name VARCHAR,
			meet_date VARCHAR,
			updated_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS races (
			race_id BIGINT PRIMARY KEY,
			card_id BIGINT,
			race_number INTEGER,
			series_specification VARCHAR,
			cancelled_race BOOLEAN,
			updated_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS results (
			race_id BIGINT PRIMARY KEY,
			card_id BIGINT,
			race_number INTEGER,
			distance INTEGER,
			start_type VARCHAR,
			start_time BIGINT,
			updated_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS runners (
			id BIGINT PRIMARY KEY, -- DuckDB can autoincrement but we will just generate an ID or use sequence
			race_id BIGINT,
			position INTEGER,
			start_number INTEGER,
			horse_name VARCHAR,
			driver_first_name VARCHAR,
			driver_last_name VARCHAR,
			distance INTEGER,
			start_track INTEGER,
			km_time VARCHAR
		)`,
		`CREATE SEQUENCE IF NOT EXISTS runner_seq`,
	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

// SaveCards saves an array of cards to the database.
func SaveCards(cards []veikkausapi.Card) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO cards (card_id, country, track_name, meet_date, updated_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for _, card := range cards {
		_, err = stmt.Exec(card.CardID, card.Country, card.TrackName, card.MeetDate, now)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// SaveRaces saves races to the DB
func SaveRaces(cardID int, races []veikkausapi.Race) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO races (race_id, card_id, race_number, series_specification, cancelled_race, updated_at) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for _, race := range races {
		_, err = stmt.Exec(race.RaceID, cardID, race.RaceNumber, race.SeriesSpecification, race.CancelledRace, now)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// SaveResult saves a result to the DB
func SaveResult(result veikkausapi.Result) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert result
	rStmt, err := tx.Prepare("INSERT OR REPLACE INTO results (race_id, card_id, race_number, distance, start_type, start_time, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer rStmt.Close()

	now := time.Now()
	_, err = rStmt.Exec(result.RaceID, result.CardID, result.RaceNumber, result.Distance, result.StartType, result.StartTime, now)
	if err != nil {
		return err
	}

	// Delete existing runners for this race to avoid duplicates
	_, err = tx.Exec("DELETE FROM runners WHERE race_id = ?", result.RaceID)
	if err != nil {
		return err
	}

	// Insert runners
	rnStmt, err := tx.Prepare("INSERT INTO runners (id, race_id, position, start_number, horse_name, driver_first_name, driver_last_name, distance, start_track, km_time) VALUES (nextval('runner_seq'), ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer rnStmt.Close()

	for _, runner := range result.Runners {
		_, err = rnStmt.Exec(result.RaceID, runner.Position, runner.StartNumber, runner.HorseName, runner.DriverFirstName, runner.DriverLastName, runner.Distance, runner.StartTrack, runner.KmTime)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// CRR is Card Races Results for json response, defined here so handlers can use it.
type CRR struct {
	Card    veikkausapi.Card
	Races   []veikkausapi.Race
	Results []veikkausapi.Result
}

// FetchAllData returns all current FI cards and their races/results from the database.
func FetchAllData() ([]CRR, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	var result []CRR

	// Get FI Cards
	cardRows, err := DB.Query("SELECT card_id, country, track_name, meet_date FROM cards WHERE country = 'FI'")
	if err != nil {
		return nil, err
	}
	defer cardRows.Close()

	for cardRows.Next() {
		var c veikkausapi.Card
		if err := cardRows.Scan(&c.CardID, &c.Country, &c.TrackName, &c.MeetDate); err != nil {
			return nil, err
		}

		crr := CRR{Card: c}

		// Get Races for this card
		raceRows, err := DB.Query("SELECT race_id, race_number, series_specification, cancelled_race FROM races WHERE card_id = ? ORDER BY race_number", c.CardID)
		if err != nil {
			return nil, err
		}
		
		for raceRows.Next() {
			var r veikkausapi.Race
			if err := raceRows.Scan(&r.RaceID, &r.RaceNumber, &r.SeriesSpecification, &r.CancelledRace); err != nil {
				raceRows.Close()
				return nil, err
			}
			crr.Races = append(crr.Races, r)

			// Get Results for this race
			resRow := DB.QueryRow("SELECT distance, start_type, start_time FROM results WHERE race_id = ?", r.RaceID)
			var res veikkausapi.Result
			res.RaceID = r.RaceID
			res.CardID = c.CardID
			res.RaceNumber = r.RaceNumber
			err := resRow.Scan(&res.Distance, &res.StartType, &res.StartTime)
			if err != nil && err != sql.ErrNoRows {
				raceRows.Close()
				return nil, err
			}

			if err != sql.ErrNoRows {
				// Fetch runners for this result
				runnerRows, err := DB.Query("SELECT position, start_number, horse_name, driver_first_name, driver_last_name, distance, start_track, km_time FROM runners WHERE race_id = ? ORDER BY position", r.RaceID)
				if err != nil {
					raceRows.Close()
					return nil, err
				}
				for runnerRows.Next() {
					var rn veikkausapi.Runner
					if err := runnerRows.Scan(&rn.Position, &rn.StartNumber, &rn.HorseName, &rn.DriverFirstName, &rn.DriverLastName, &rn.Distance, &rn.StartTrack, &rn.KmTime); err != nil {
						runnerRows.Close()
						raceRows.Close()
						return nil, err
					}
					res.Runners = append(res.Runners, rn)
				}
				runnerRows.Close()
				crr.Results = append(crr.Results, res)
			}
		}
		raceRows.Close()
		result = append(result, crr)
	}

	return result, nil
}
