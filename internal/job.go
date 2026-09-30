package internal

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"time"
)

// runJob executes the tneb-scraper job once
func runJob(rc runConfig) ([]EBBill, error) {
	log.Println("Starting a new job")
	bills, err := extractBills(rc)
	if err != nil {
		err = fmt.Errorf("failed to extract eb bills: %w", err)
	}

	if !rc.skipEmail {
		mailErr := SendMail(rc.guser, rc.gpass, rc.gtarget, bills, err)
		if mailErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to send email: %w", mailErr))
		}
	}

	return bills, err
}

// ScraperScheduler wraps the periodic scraping logic.
type ScraperScheduler struct {
	rc   runConfig     // Config to determine scraper behaviour
	days []int         // List of dates each month when scraper should run
	ch   chan struct{} // Signal ScraperScheduler to shutdown
}

// NewScraperScheduler returns a ready to use Scheduler. Days denotes the list
// of dates each month when the scraper should run.
func NewScraperScheduler(rc runConfig, days ...int) ScraperScheduler {
	return ScraperScheduler{rc: rc, days: days, ch: make(chan struct{})}
}

// Start runs an infinte loop until until Shutdown() is called. On startup goes
// to sleep until next start of day at which point it starts a periodic ticker
// that ticks every 24 hours.
func (s ScraperScheduler) Start() {
	// Get the duration between now and tomorrow and sleep for the duration
	now := time.Now()
	y, m, d := now.Date()
	tomorrow := time.Date(y, m, d+1, 9, 0, 0, 0, now.Location())
	duration := tomorrow.Sub(now)
	log.Printf("Sleeping for %v until tomorrow.", duration)
	select {
	case <-time.After(duration):
	case <-s.ch:
		log.Println("Cronjob shutdown complete")
		return
	}

	// Start ticking
	log.Println("Starting cron")
	ticker := time.NewTicker(time.Hour * 24)
	defer ticker.Stop()

	for {
		if day := time.Now().Day(); slices.Contains(s.days, day) {
			// Execute the job
			if _, err := runJob(s.rc); err != nil {
				log.Println("Process finished with errors:", err)
			} else {
				log.Println("Process finished successfully.")
			}
		}

		// Block until next tick
		select {
		case <-ticker.C:
		case <-s.ch:
			log.Println("Cronjob shutdown complete")
			return
		}
	}
}

// Shutdown signals the scheduler to exit.
func (s ScraperScheduler) Shutdown() {
	close(s.ch)
}
