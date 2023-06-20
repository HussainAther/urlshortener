package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func main() {
	// Open the database connection
	var err error
	db, err = sql.Open("sqlite3", "./shortener.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the URLs table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			long_url TEXT NOT NULL,
			short_code TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", createURL).Methods("POST")
	r.HandleFunc("/{code}", redirectToURL).Methods("GET")

	http.Handle("/", r)

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createURL(w http.ResponseWriter, r *http.Request) {
	longURL := r.FormValue("url")

	// Validate the long URL
	if longURL == "" {
		http.Error(w, "URL cannot be empty", http.StatusBadRequest)
		return
	}

	// Generate a unique short code for the URL
	shortCode := generateShortCode(longURL)

	// Insert the URL into the database
	_, err := db.Exec("INSERT INTO urls (long_url, short_code) VALUES (?, ?)", longURL, shortCode)
	if err != nil {
		http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
		return
	}

	// Generate the short URL
	shortURL := fmt.Sprintf("http://localhost:8080/%s", shortCode)

	// Return the short URL to the user
	w.Write([]byte(shortURL))
}

func redirectToURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["code"]

	// Retrieve the long URL from the database
	var longURL string
	err := db.QueryRow("SELECT long_url FROM urls WHERE short_code = ?", shortCode).Scan(&longURL)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Redirect the user to the original URL
	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

func generateShortCode(url string) string {
	hash := sha1.Sum([]byte(url + time.Now().String()))
	return base64.URLEncoding.EncodeToString(hash[:])[:8]
}

