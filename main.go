package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/infinage/tneb-scraper/internal"
)

func main() {
	// Load the config settings for the scraper
	rc, err := internal.NewRunConfig()
	if err != nil {
		log.Fatalf("Failed to init: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt,
		syscall.SIGTERM)
	defer cancel()

	// Listen on endpoint "/?secret=<rc.apiSecret>"
	server := internal.NewScraperServer(":8080", rc)
	defer server.Shutdown(5)
	go server.Start()

	// Run on 10th and 20th of each month
	scheduler := internal.NewScraperScheduler(rc, 10, 20)
	defer scheduler.Shutdown()
	go scheduler.Start()

	// block main thread until interupted, after which defer blocks
	// execute and the goroutines are shutdown
	<-ctx.Done()
}
