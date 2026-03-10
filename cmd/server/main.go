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
	"time"

	tacticalai "github.com/apex/monitor/pkg/ai"
	"github.com/apex/monitor/pkg/receiver"
	"github.com/apex/monitor/pkg/storage"
	apex "github.com/apex/monitor/proto"
	"github.com/google/uuid"
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
	store        storage.Provider
	rdb          *redis.Client
	recv         *receiver.Receiver
	ai           *tacticalai.TacticalAI
	isPersistent bool
}

func NewServer(store storage.Provider, rdb *redis.Client) *Server {
	recv, _ := receiver.New()
	s := &Server{
		store:        store,
		rdb:          rdb,
		recv:         recv,
		ai:           tacticalai.NewTacticalAI(),
		isPersistent: false, // Will be set in main
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

				projectID := message.Values["project_id"].(string)

				err := s.store.SaveReport(report, projectID)
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

	// Validate Ingest Key (API Key)
	key := r.Header.Get("X-Apex-API-Key")
	if key == "" {
		http.Error(w, "Unauthorized: Missing X-Apex-API-Key", http.StatusUnauthorized)
		return
	}

	projectID, err := s.store.ValidateKey(key)
	if err != nil {
		slog.Error("Failed to validate ingest key", "error", err)
	}
	if projectID == "" {
		// Fallback for demo: accept default key if env is set, but warn
		defaultKey := os.Getenv("APEX_API_KEY")
		if defaultKey == "" || key != defaultKey {
			http.Error(w, "Unauthorized: Invalid Ingest Key", http.StatusUnauthorized)
			return
		}
		projectID = "00000000-0000-0000-0000-000000000000" // Default Demo Project
	}

	compressed, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Read error", http.StatusInternalServerError)
		return
	}

	batch, err := s.recv.Unpack(compressed)
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
			Values: map[string]interface{}{
				"data":       string(data),
				"project_id": projectID,
			},
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

	projectID := r.URL.Query().Get("project_id")
	reports, err := s.store.GetReports(50, projectID)
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

	projectID := r.URL.Query().Get("project_id")
	reports, err := s.store.GetReports(50, projectID)
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

	response := s.ai.Chat(req.Message, req.ReportID)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	json.NewEncoder(w).Encode(map[string]string{"response": response})
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
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
		UserID string `json:"user_id"`
		Name   string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Generate a secure Ingest Key
	ingestKey := "apex_" + uuid.New().String()
	if req.UserID == "" {
		req.UserID = "anonymous"
	}

	p := &storage.Project{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Name:      req.Name,
		IngestKey: ingestKey,
		CreatedAt: time.Now(),
	}

	slog.Info("Attempting to save project", "id", p.ID, "name", p.Name, "userId", p.UserID)

	if err := s.store.SaveProject(p); err != nil {
		slog.Error("Failed to save project", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (s *Server) handleGetProjects(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	projects, err := s.store.GetProjects(userID)
	if err != nil {
		slog.Error("Failed to fetch projects", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"persistent": s.isPersistent,
		"status":     "operational",
		"timestamp":  time.Now().Unix(),
	})
}

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// CockroachDB default insecure connection
		connStr = "postgresql://root@127.0.0.1:5433/defaultdb?sslmode=disable"
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
		rdbAddr = "127.0.0.1:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: rdbAddr,
	})
	slog.Info("Redis connection initialized", "addr", rdbAddr)

	srv := NewServer(store, rdb)
	if store != nil {
		srv.isPersistent = true
	}

	http.HandleFunc("/ingest", srv.handleIngest)
	http.HandleFunc("/reports", srv.handleGetReports)
	http.HandleFunc("/api/reports", srv.handleGetReportsJSON)
	http.HandleFunc("/api/chat", srv.handleChat)
	http.HandleFunc("/api/projects", srv.handleGetProjects)
	http.HandleFunc("/api/projects/create", srv.handleCreateProject)
	http.HandleFunc("/api/status", srv.handleStatus)
	http.Handle("/metrics", promhttp.Handler())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	slog.Info("Apex Production Receiver starting", "port", port)
	http.ListenAndServe(":"+port, nil)
}
