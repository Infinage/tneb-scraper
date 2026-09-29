package internal

import (
	"fmt"
	"log"
	"os"
)

// runConfig is an internal struct containing secrets and
// values that determine scraping behaviour.
type runConfig struct {
	ebURL, ebUser, ebPass string
	guser, gpass, gtarget string

	mapping   map[string]string // EB Consumer No -> Consumer Name
	apiSecret string            // ScraperServer endpoint is authenticated with this key
	skipEmail bool              // Emails are skipped if 'guser', 'gpass', 'gtarget' is not set
}

// NewRunConfig loads its values from env. Looks for the following:
//   - TNEB_USERNAME (REQUIRED)
//   - TNEB_PASSWORD (REQUIRED)
//   - TNEB_LOGIN_URL (REQUIRED)
//   - GMAIL_FROM_ADDRESS
//   - GMAIL_TO_ADDRESS
//   - GMAIL_APP_PWD
//   - API_SECRET_KEY
func NewRunConfig() (runConfig, error) {
	// Warn the user in case dotenv load fails (non fatal)
	if err := loadEnv(); err != nil {
		log.Println("env load fail (non-fatal):", err)
	}

	var rc runConfig
	rc.ebURL = os.Getenv("TNEB_LOGIN_URL")
	rc.ebUser = os.Getenv("TNEB_USERNAME")
	rc.ebPass = os.Getenv("TNEB_PASSWORD")

	// Fail if username, password or login url is not set
	if rc.ebUser == "" || rc.ebPass == "" || rc.ebURL == "" {
		return runConfig{}, fmt.Errorf("env: TNEB_USERNAME,TNEB_PASSWORD or " +
			"TNEB_LOGIN_URL not set")
	}

	rc.guser = os.Getenv("GMAIL_FROM_ADDRESS")
	rc.gpass = os.Getenv("GMAIL_APP_PWD")
	rc.gtarget = os.Getenv("GMAIL_TO_ADDRESS")

	// Load the mapping file
	var err error
	rc.mapping, err = loadConsumerMapping()
	if err != nil {
		log.Println("Consumer mapping load fail (non-fatal):", err)
		rc.mapping = make(map[string]string)
	}

	// Skip sending out email if mail related env variables are not set
	if rc.guser == "" || rc.gpass == "" || rc.gtarget == "" {
		rc.skipEmail = true
		log.Println("Env: GMAIL_FROM_ADDRESS, GMAIL_TO_ADDRESS or GMAIL_APP_PWD " +
			"is not set (mail step skipped)")
	}

	rc.apiSecret = os.Getenv("API_SECRET_KEY")
	if rc.apiSecret == "" {
		log.Println("Env: API_SECRET_KEY not set, endpoint not secure.")
	}

	return rc, nil
}
