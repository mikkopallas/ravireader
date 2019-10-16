package main

import (
	"fmt"

	"github.com/mikcompany/ravireader/veikkausapi"
)

func main() {
	fmt.Println("Ravi Reader V0.1")

	cards := veikkausapi.FetchCards()

	for _, card := range cards {
		if card.Country == "FI" {
			fmt.Println(card)

			races := veikkausapi.FetchRaces(card.CardID)

			for _, race := range races {
				fmt.Println(veikkausapi.FetchResult(race.RaceID))
			}
		}
	}

}
