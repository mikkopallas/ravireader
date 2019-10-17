package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/mikcompany/ravireader/veikkausapi"
)

func rootHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "ravireader")
}

// CRR is Card Races Results for json response.
type CRR struct {
	Card    veikkausapi.Card
	Races   []veikkausapi.Race
	Results []veikkausapi.Result
}

func jsonHandler(w http.ResponseWriter, req *http.Request) {
	result := CRR{}
	cards := veikkausapi.FetchCards()

	for _, card := range cards {
		if card.Country == "FI" {
			result.Card = card
			result.Races = veikkausapi.FetchRaces(card.CardID)

			for _, race := range result.Races {
				result.Results = append(result.Results, veikkausapi.FetchResult(race.RaceID))
			}
		}
	}

	jData, err := json.Marshal(result)
	if err != nil {
		// handle error
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

func main() {
	fmt.Println("Ravi Reader V0.1")

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/json", jsonHandler)
	log.Fatal(http.ListenAndServe(":80", nil))
}
