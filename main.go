package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mikcompany/ravireader/db"
	"github.com/mikcompany/ravireader/updater"
)

type PageData struct {
	SelectedDate string
	Data         []db.CRR
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.New("index.html").Funcs(template.FuncMap{
		"formatTime": func(ms int) string {
			if ms == 0 {
				return "-"
			}
			return time.UnixMilli(int64(ms)).Format("15:04")
		},
		"formatPosition": func(pos int) string {
			if pos == 0 {
				return "-"
			}
			if pos <= 100 {
				return strconv.Itoa(pos)
			}
			switch pos {
			case 101:
				return "hll"
			case 102:
				return "hpl"
			case 103:
				return "hva"
			case 104:
				return "k"
			case 105:
				return "hyl"
			case 106:
				return "0"
			case 107:
				return "poissa"
			default:
				return "-"
			}
		},
	}).ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	date := req.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	hasData, err := db.HasDataForDate(date)
	if err == nil && !hasData {
		// Fetch on-demand
		updater.FetchAndSaveForDate(date)
	}

	result, err := db.FetchAllData(date)
	if err != nil {
		http.Error(w, "Failed to fetch data from database: " + err.Error(), http.StatusInternalServerError)
		return
	}

	pageData := PageData{
		SelectedDate: date,
		Data:         result,
	}

	tmpl.Execute(w, pageData)
}

func jsonHandler(w http.ResponseWriter, req *http.Request) {
	date := req.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	hasData, err := db.HasDataForDate(date)
	if err == nil && !hasData {
		updater.FetchAndSaveForDate(date)
	}

	result, err := db.FetchAllData(date)
	if err != nil {
		http.Error(w, "Failed to fetch data from database: " + err.Error(), http.StatusInternalServerError)
		return
	}

	jData, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

func main() {
	fmt.Println("Ravi Reader V0.1")

	serveQuack := flag.Bool("serve-quack", false, "Start the DuckDB Quack server locally")
	flag.Parse()

	// Initialize database
	if err := db.Init(*serveQuack); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Start background updater to fetch data every 5 minutes
	go updater.Start(5 * time.Minute)

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/json", jsonHandler)
	
	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
