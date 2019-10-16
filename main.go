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

func jsonHandler(w http.ResponseWriter, req *http.Request) {

	var result []interface{}

	cards := veikkausapi.FetchCards()

	for _, card := range cards {
		if card.Country == "FI" {
			fmt.Println(card)

			races := veikkausapi.FetchRaces(card.CardID)

			for _, race := range races {
				//fmt.Println(veikkausapi.FetchResult(race.RaceID))
				result = append(result, veikkausapi.FetchResult(race.RaceID))
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
