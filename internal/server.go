package internal

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// endpoint listens on '/' and triggers the job run when hit. It
// requires 'secret' query value set (env: API_SECRET_KEY).
func endpoint(rc runConfig) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		if r.URL.Query().Get("secret") != rc.apiSecret {
			http.Error(w, "Unauthorised", http.StatusUnauthorized)
			return
		}

		bills, err := runJob(rc)
		if err != nil {
			http.Error(w, "Job failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Write the bill details as a JSON reponse
		log.Println("Job completed successfully")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(bills); err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
			return
		}
	}
}

// ScraperServer listens on '/' endpoint and triggers the scraper to run.
type ScraperServer struct {
	addr string
	srv  *http.Server
}

// NewScraperServer creates a ready to go server.
func NewScraperServer(addr string, rc runConfig) ScraperServer {
	srv := http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(endpoint(rc)),
	}

	return ScraperServer{addr: addr, srv: &srv}
}

// Start listens on the addr provided and blocks until Shutdown is called.
func (s ScraperServer) Start() {
	log.Printf("Starting server on %s", s.addr)
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

// Shutdown signals the server to cleanup. timeoutSec is the grace duration
// for the processes to complete.
func (s ScraperServer) Shutdown(timeoutSec int) {
	log.Println("Shutdown signal received")

	// Grace period for the server shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(),
		time.Second*time.Duration(timeoutSec))
	defer shutdownCancel()

	s.srv.Shutdown(shutdownCtx)
	log.Println("Sever shutdown complete")
}
