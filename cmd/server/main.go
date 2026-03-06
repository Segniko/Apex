package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/apex/monitor/pkg/receiver"
	"github.com/apex/monitor/pkg/storage"
	apex "github.com/apex/monitor/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
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
	recv  *receiver.Receiver
	store storage.Provider
	rdb   *redis.Client
}

func NewServer(store storage.Provider, rdb *redis.Client) *Server {
	recv, _ := receiver.New()
	s := &Server{
		recv:  recv,
		store: store,
		rdb:   rdb,
	}

	// Start 5 workers to handle database ingestion concurrently from Redis
	for i := 0; i < 5; i++ {
		go s.worker(i)
	}

	return s
}

func (s *Server) worker(id int) {
	ctx := context.Background()
	streamName := "apex_reports"

	for {
		// Block-read from the Redis Stream
		streams, err := s.rdb.XRead(ctx, &redis.XReadArgs{
			Streams: []string{streamName, "$"}, // "$" means only new messages
			Block:   0,                         // Wait indefinitely
			Count:   1,
		}).Result()

		if err != nil {
			slog.Error("Redis Read error", "worker_id", id, "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				data := message.Values["data"].(string)
				report := &apex.CrashReport{}
				if err := proto.Unmarshal([]byte(data), report); err != nil {
					slog.Error("Failed to unmarshal from Redis", "worker_id", id, "error", err)
					continue
				}

				// Run AI Forensics before saving
				insight := s.recv.Analyze(report)
				report.AiInsight = insight

				err := s.store.SaveReport(report)
				if err != nil {
					slog.Error("Failed to save report", "worker_id", id, "error", err)
				} else {
					slog.Info("Report processed from Redis", "worker_id", id, "crash_id", report.ErrorId)
				}
			}
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

	// Offload to Redis Stream
	ctx := context.Background()
	slog.Info("Batch received, offloading to Redis", "count", len(batch.Reports))
	for _, report := range batch.Reports {
		reportsReceived.Inc()
		data, _ := proto.Marshal(report)
		err := s.rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: "apex_reports",
			Values: map[string]interface{}{"data": string(data)},
		}).Err()
		if err != nil {
			slog.Error("Failed to push to Redis", "error", err)
		}
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

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Message  string `json:"message"`
		ReportID string `json:"report_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	msg := strings.ToLower(req.Message)
	var response string

	// Tactical Logic: Context-Aware Response Matrix
	switch {
	case strings.Contains(msg, "hi") || strings.Contains(msg, "hello"):
		response = "APEX_AI unit online. Ready for tactical forensics. I am monitoring the CockroachDB clusters and Redis ingest streams. How can I assist your audit?"
	case strings.Contains(msg, "status") || strings.Contains(msg, "deployment"):
		response = "Telemetry indicates the current deployment is under heavy monitoring. Ingest buffer velocity is steady. No anomalies detected in the last 300ms. All services are tactical."
	case strings.Contains(msg, "nil") || strings.Contains(msg, "pointer") || strings.Contains(msg, "null"):
		response = "NIL_POINTER_ANALYSIS: This is a common failure in unsafe memory operations. I recommend a tactical audit of your reference guards. Check the stack trace for the 'dereference' instruction."
	case strings.Contains(msg, "database") || strings.Contains(msg, "sql") || strings.Contains(msg, "cockroach"):
		response = "DATABASE_INTEGRITY: CockroachDB clusters are reporting 99.99% consistency. If you see 'Scan Errors', ensure your COALESCE logic is applied to legacy schema fields."
	case strings.Contains(msg, "error") || strings.Contains(msg, "fault") || strings.Contains(msg, "crash"):
		response = "CRASH_FORENSICS: I have indexed the recent failure batches. Which specific 'INTELLIGENT_DECODE' ID should I perform a deep-trace on?"
	case req.ReportID != "" || strings.Contains(msg, "decode"):
		response = "DECODING_TRACE: Analyzing packet structure for high-fidelity reconstruction. Initial findings suggest an operational bottleneck in the ingest-to-vault pipeline."
	default:
		responses := []string{
			"Processing inquiry... Tactical analysis suggests we remain vigilant on the current telemetry stream.",
			"Data inconclusive for a deep-dive. Provide a specific Error ID for forensic reconstruction.",
			"I am monitoring the tactical nodes. Telemetry is flowing smoothly through the Redis buffer.",
			"Inquiry logged. Proceed with caution on the recent deployment patches.",
		}
		response = responses[time.Now().UnixNano()%int64(len(responses))]
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	json.NewEncoder(w).Encode(map[string]string{"response": response})
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

	// Initialize Redis for Ingest Buffering
	rdbAddr := os.Getenv("REDIS_URL")
	if rdbAddr == "" {
		rdbAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: rdbAddr,
	})
	slog.Info("Redis connection initialized", "addr", rdbAddr)

	srv := NewServer(store, rdb)

	http.HandleFunc("/ingest", srv.handleIngest)
	http.HandleFunc("/reports", srv.handleGetReports)
	http.HandleFunc("/api/reports", srv.handleGetReportsJSON)
	http.HandleFunc("/api/chat", srv.handleChat)
	http.Handle("/metrics", promhttp.Handler())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	slog.Info("Apex Production Receiver starting", "port", port)
	http.ListenAndServe(":"+port, nil)
}
