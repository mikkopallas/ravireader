package updater

import (
	"log"
	"time"

	"github.com/mikcompany/ravireader/db"
	"github.com/mikcompany/ravireader/veikkausapi"
)

// Start begins the background loop to fetch data from Veikkaus API and save it to the DB.
func Start(interval time.Duration) {
	log.Println("Starting background updater...")

	// Run once immediately
	update()

	ticker := time.NewTicker(interval)
	for range ticker.C {
		update()
	}
}

func update() {
	err := FetchAndSaveForDate("today")
	if err != nil {
		log.Printf("Updater: Background update failed: %v\n", err)
	} else {
		log.Println("Updater: Finished updating all data for today.")
	}
}

// FetchAndSaveForDate fetches and saves data for a given date.
func FetchAndSaveForDate(date string) error {
	log.Printf("Updater: Fetching cards for %s...\n", date)
	cards := veikkausapi.FetchCards(date)
	
	err := db.SaveCards(cards)
	if err != nil {
		return err
	}

	for _, card := range cards {
		if card.Country == "FI" {
			// Fetch races for this card
			races := veikkausapi.FetchRaces(card.CardID)
			
			err = db.SaveRaces(card.CardID, races)
			if err != nil {
				log.Printf("Updater: Failed to save races for card %d: %v\n", card.CardID, err)
				continue
			}

			// Fetch results for each race
			for _, race := range races {
				result := veikkausapi.FetchResult(race.RaceID)
				
				err = db.SaveResult(result)
				if err != nil {
					log.Printf("Updater: Failed to save result for race %d: %v\n", race.RaceID, err)
					continue
				}
			}
		}
	}
	
	return nil
}
