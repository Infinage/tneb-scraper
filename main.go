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

	gmailSender := os.Getenv("GMAIL_FROM_ADDRESS")
	gmailPassword := os.Getenv("GMAIL_APP_PWD")
	gmailTarget := os.Getenv("GMAIL_TO_ADDRESS")

	// Skip sending out email if mail related env variables are not set
	skipEmail := false
	if gmailSender == "" || gmailPassword == "" || gmailTarget == "" {
		skipEmail = true
		log.Println("Env: GMAIL_FROM_ADDRESS, GMAIL_TO_ADDRESS or GMAIL_APP_PWD " +
			"is not set (mail step skipped)")
	}

	// Load the mapping file
	mapping, err := internal.LoadConsumerMapping()
	if err != nil {
		log.Println("Consumer mapping load fail (non-fatal):", err)
		mapping = make(map[string]string)
	}

	bills, err := internal.ExtractBills(loginURL, username, password, mapping)
	if err != nil {
		log.Println("Failed to extract EB bills:", err)
	} else {
		log.Println("EB Bill details:", bills)
	}

	if !skipEmail {
		err = internal.SendMail(gmailSender, gmailPassword, gmailTarget, bills, err)
		if err != nil {
			log.Fatalln("Failed to send email:", err)
		}
	}

	if err != nil {
		log.Println("Process finished with errors.")
		return
	}

	log.Println("Process finished successfully.")
}
