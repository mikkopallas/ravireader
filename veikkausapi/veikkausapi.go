package veikkausapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

// Card ...
type Card struct {
	CardID    int
	Country   string
	TrackName string
	MeetDate  string
}

// CardCollection ...
type CardCollection struct {
	Collection []Card
}

// Race ...
type Race struct {
	RaceID              int
	RaceNumber          int
	SeriesSpecification string
	CancelledRace       bool
}

// RaceCollection ...
type RaceCollection struct {
	Collection []Race
}

// Runner ...
type Runner struct {
	Position        int
	StartNumber     int
	HorseName       string
	DriverFirstName string
	DriverLastName  string
	Distance        int
	StartTrack      int
	KmTime          string
}

// Result ...
type Result struct {
	RaceID     int
	CardID     int
	RaceNumber int
	Distance   int
	StartType  string
	StartTime  int
	Runners    []Runner
}

// FetchCards ...
func FetchCards() []Card {
	resp, err := http.Get("https://www.veikkaus.fi/api/toto-info/v1/cards/today")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var collection CardCollection
	err = json.Unmarshal(body, &collection)
	if err != nil {
		panic(err)
	}

	return collection.Collection
}

// FetchRaces ...
func FetchRaces(cardID int) []Race {
	resp, err := http.Get("https://www.veikkaus.fi/api/toto-info/v1/card/" + strconv.Itoa(cardID) + "/results")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var collection RaceCollection
	err = json.Unmarshal(body, &collection)
	if err != nil {
		panic(err)
	}

	return collection.Collection
}

// FetchResults ...
func FetchResult(raceID int) Result {
	resp, err := http.Get("https://www.veikkaus.fi/api/toto-info/v1/race/" + strconv.Itoa(raceID) + "/competition-results")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var result Result
	err = json.Unmarshal(body, &result)
	if err != nil {
		panic(err)
	}

	return result
}
