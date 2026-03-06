package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/apex/monitor/pkg/receiver"
	"github.com/apex/monitor/pkg/storage"
	apex "github.com/apex/monitor/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	reportsReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "apex_reports_received_total",
		Help: "The total number of crash reports received",
	})
	ingestErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "apex_ingest_errors_total",
		Help: "The total number of ingestion errors",
	})
	ingestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "apex_ingest_duration_seconds",
		Help:    "Time spent unpacking and routing batches",
		Buckets: prometheus.DefBuckets,
	})
)

type Server struct {
	recv    *receiver.Receiver
	store   storage.Provider
	jobChan chan *apex.CrashReport
}

func NewServer(store storage.Provider) *Server {
	recv, _ := receiver.New()
	s := &Server{
		recv:    recv,
		store:   store,
		jobChan: make(chan *apex.CrashReport, 1000),
	}

	// Start 5 workers to handle database ingestion concurrently
	for i := 0; i < 5; i++ {
		go s.worker(i)
	}

	return s
}

func (s *Server) worker(id int) {
	for report := range s.jobChan {
		err := s.store.SaveReport(report)
		if err != nil {
			slog.Error("Failed to save report", "worker_id", id, "error", err)
		} else {
			slog.Info("Report saved successfully", "worker_id", id, "crash_id", report.ErrorId)
		}
	}
}

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Apex Ingest Endpoint\nStatus: Online\n\nUsage: This endpoint only accepts POST requests from the Apex Agent.\nTo view the dashboard, please visit /reports")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	timer := prometheus.NewTimer(ingestDuration)
	defer timer.ObserveDuration()

	// Simple API Key check
	apiKey := r.Header.Get("X-Apex-API-Key")
	if apiKey != os.Getenv("APEX_API_KEY") && os.Getenv("APEX_API_KEY") != "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Read error", http.StatusInternalServerError)
		return
	}

	batch, err := s.recv.Unpack(body)
	if err != nil {
		slog.Error("Failed to unpack batch", "error", err)
		ingestErrors.Inc()
		http.Error(w, "Unpack error", http.StatusBadRequest)
		return
	}

	// Offload to workers
	slog.Info("Batch received", "count", len(batch.Reports))
	for _, report := range batch.Reports {
		reportsReceived.Inc()
		s.jobChan <- report
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Batch accepted: %d reports", len(batch.Reports))
}

func (s *Server) handleGetReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reports, err := s.store.GetReports(50)
	if err != nil {
		slog.Error("Failed to fetch reports", "error", err)
	}

	renderDashboard(w, reports)
}

func (s *Server) handleGetReportsJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reports, err := s.store.GetReports(50)
	if err != nil {
		slog.Error("Failed to fetch reports", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Allow CORS for the dashboard
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(reports)
}

func renderDashboard(w http.ResponseWriter, reports []*apex.CrashReport) {
	tmpl, err := template.ParseFiles("templates/dashboard.html")
	if err != nil {
		slog.Error("Template error", "error", err)
		http.Error(w, "UI rendering error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Reports []*apex.CrashReport
	}{
		Reports: reports,
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "dashboard", data)
}

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Use port 5433 to avoid host conflicts with local Postgres
		connStr = "postgres://postgres:postgres@localhost:5433/apex?sslmode=disable"
	}

	var store storage.Provider
	var err error

	// Production Resilience: Retry connecting to DB
	slog.Info("Attempting to connect to database...")
	for i := 0; i < 5; i++ {
		pgStore, pgErr := storage.NewPostgres(connStr)
		if pgErr == nil {
			if err = pgStore.Initialize(); err == nil {
				store = pgStore
				break
			}
		}
		slog.Warn("Database not ready, retrying in 2s...", "attempt", i+1, "error", pgErr)
		time.Sleep(2 * time.Second)
	}

	if store == nil {
		slog.Error("Database connection failed, running in MEMORY mode (Reports will NOT persist after restart)")
		store = storage.NewMemoryStore()
	} else {
		slog.Info("Database connection established")
	}

	srv := NewServer(store)

	http.HandleFunc("/ingest", srv.handleIngest)
	http.HandleFunc("/reports", srv.handleGetReports)
	http.HandleFunc("/api/reports", srv.handleGetReportsJSON)
	http.Handle("/metrics", promhttp.Handler())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	slog.Info("Apex Production Receiver starting", "port", port)
	http.ListenAndServe(":"+port, nil)
}
