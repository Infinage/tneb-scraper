package main

import (
	"log"
	"os"

	"github.com/infinage/tneb-scraper/internal"
)

func main() {
	// Warn the user in case dotenv load fails (non fatal)
	if err := internal.LoadEnv(); err != nil {
		log.Println("env load fail (non-fatal):", err)
	}

	loginURL := os.Getenv("TNEB_LOGIN_URL")
	username := os.Getenv("TNEB_USERNAME")
	password := os.Getenv("TNEB_PASSWORD")

	// Fail if username, password or login url is not set
	if username == "" || password == "" || loginURL == "" {
		log.Fatalln("Env: TNEB_USERNAME, TNEB_PASSWORD or TNEB_LOGIN_URL is not set")
	}

	// Load the mapping file
	mapping, err := internal.LoadConsumerMapping()
	if err != nil {
		log.Println("Consumer mapping load fail (non-fatal):", err)
		mapping = make(map[string]string)
	}

	bills, err := internal.ExtractBills(loginURL, username, password, mapping)
	if err != nil {
		log.Fatalln("Failed to extract EB bills:", err)
	}

	log.Println(bills)
	log.Println("Process finished successfully.")
}
