package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/mikcompany/ravireader/db"
	"github.com/mikcompany/ravireader/updater"
)

func rootHandler(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := db.FetchAllData()
	if err != nil {
		http.Error(w, "Failed to fetch data from database: " + err.Error(), http.StatusInternalServerError)
		return
	}

	// db.FetchAllData returns []db.CRR, but template expects it directly
	tmpl.Execute(w, result)
}

func jsonHandler(w http.ResponseWriter, req *http.Request) {
	result, err := db.FetchAllData()
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
