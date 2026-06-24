package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	tacticalai "github.com/Segniko/Apex/pkg/ai"
	"github.com/Segniko/Apex/pkg/alert"
	"github.com/Segniko/Apex/pkg/limiter"
	"github.com/Segniko/Apex/pkg/receiver"
	"github.com/Segniko/Apex/pkg/storage"
	apex "github.com/Segniko/Apex/proto"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"google.golang.org/api/iterator"
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
	dbSaveErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "apex_db_save_errors_total",
		Help: "The total number of failed report persistence attempts",
	})
	aiEnrichErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "apex_ai_enrich_errors_total",
		Help: "The total number of dropped/failed AI enrichment jobs",
	})
)

// enrichJob is queued after a report is persisted so AI analysis runs off the
// hot ingest path.
type enrichJob struct {
	report      *apex.CrashReport
	projectID   string
	fingerprint string
}

type Server struct {
	store        storage.Provider
	rdb          *redis.Client
	recv         *receiver.Receiver
	ai           *tacticalai.TacticalAI
	limiter      *limiter.RateLimiter
	alerter      *alert.Notifier
	enrichCh     chan *enrichJob
	isPersistent bool
	fileCache    map[string][]string
	cacheOnce    sync.Once
}

func NewServer(store storage.Provider, rdb *redis.Client, geminiKey string) *Server {
	recv, _ := receiver.New()
	s := &Server{
		store:        store,
		rdb:          rdb,
		recv:         recv,
		ai:           tacticalai.NewTacticalAI(geminiKey),
		limiter:      limiter.NewRateLimiter(rdb),
		alerter:      alert.New(os.Getenv("APEX_WEBHOOK_URL"), os.Getenv("APEX_DASHBOARD_URL")),
		enrichCh:     make(chan *enrichJob, 2048),
		isPersistent: false, // Will be set in main
		fileCache:    make(map[string][]string),
	}

	if s.alerter.Enabled() {
		slog.Info("[APEX_SERVER] Crash alerting enabled (webhook configured)")
	}

	// Warm up the file cache in the background to avoid blocking startup
	go s.warmUpFileCache()

	// Start the ingest worker (fast persistence) and a small enrichment pool
	// (rate-limited Gemini analysis, off the hot path).
	go s.worker(0)
	for i := 0; i < 2; i++ {
		go s.enrichmentWorker(i)
	}

	return s
}

// normalizeErrorID guarantees a UUID id (the storage column is UUID-typed).
// Empty -> random; non-UUID -> deterministic UUIDv5 so duplicates stay stable.
func normalizeErrorID(id string) string {
	if _, err := uuid.Parse(id); err == nil {
		return id
	}
	if id == "" {
		return uuid.New().String()
	}
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(id)).String()
}

// fireAlert sends a webhook notification the first time a signature is seen
// (deduped per fingerprint+project for 1h).
func (s *Server) fireAlert(ctx context.Context, projectID, fingerprint string, report *apex.CrashReport) {
	if !s.alerter.Enabled() {
		return
	}
	dedupeKey := "alert:sent:" + projectID + ":" + fingerprint
	if ok, _ := s.rdb.SetNX(ctx, dedupeKey, 1, time.Hour).Result(); ok {
		go s.alerter.Notify(context.Background(), alert.Crash{
			ProjectID: projectID,
			ErrorID:   report.ErrorId,
			Message:   report.Message,
			Stack:     report.StackTrace,
			Insight:   report.AiInsight,
		})
	}
}

// enqueueEnrichment hands a saved report to the AI pool without blocking. If
// the queue is saturated the report simply keeps its heuristic insight.
func (s *Server) enqueueEnrichment(report *apex.CrashReport, projectID, fingerprint string) {
	select {
	case s.enrichCh <- &enrichJob{report: report, projectID: projectID, fingerprint: fingerprint}:
	default:
		aiEnrichErrors.Inc()
		slog.Warn("Enrichment queue full, keeping heuristic insight", "crash_id", report.ErrorId)
	}
}

// enrichmentWorker drains the AI queue, respecting cache + per-project quota,
// then writes the upgraded insight back onto the stored report.
func (s *Server) enrichmentWorker(id int) {
	ctx := context.Background()
	for job := range s.enrichCh {
		cacheKey := "ai_insight:" + job.fingerprint

		if cached, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
			if err := s.store.UpdateInsight(job.report.ErrorId, cached); err != nil {
				slog.Error("Failed to write cached insight", "error", err)
			}
			continue
		}

		// Rate Limit: 100 insights per hour per project.
		allowed, _ := s.limiter.Allow(ctx, job.projectID+":analysis", 100, 1*time.Hour)
		if !allowed {
			slog.Warn("AI Analysis rate limit exceeded", "project_id", job.projectID)
			continue
		}

		sourceContext := s.getSourceContext(job.report.StackTrace)
		similar, _ := s.store.GetSimilarReports(job.report.Message, 3, job.projectID)
		historicalContext := ""
		if len(similar) > 0 {
			historicalContext = "\nHISTORICAL_SIMILARITY_DATA:\n"
			for _, sr := range similar {
				if sr.ErrorId == job.report.ErrorId {
					continue
				}
				historicalContext += fmt.Sprintf("- Past Error: %s\n  Insight: %s\n", sr.Message, sr.AiInsight)
			}
		}

		insight := s.ai.AnalyzeReport(job.report.Message, job.report.StackTrace+historicalContext, sourceContext)
		if insight == "" {
			aiEnrichErrors.Inc()
			continue
		}

		s.rdb.Set(ctx, cacheKey, insight, 30*24*time.Hour)
		if err := s.store.UpdateInsight(job.report.ErrorId, insight); err != nil {
			slog.Error("Failed to persist AI insight", "worker_id", id, "error", err)
		} else {
			slog.Info("AI insight enriched", "worker_id", id, "crash_id", job.report.ErrorId)
		}

		// Quota safety pacing between live Gemini calls.
		time.Sleep(2 * time.Second)
	}
}

func (s *Server) warmUpFileCache() {
	// Source-context enrichment reads files from THIS server's filesystem, which
	// only makes sense for local/self-hosted dev where the app's source is present.
	// In a hosted receiver it's the Apex repo, not the user's code, so skip the walk.
	if os.Getenv("APEX_LOCAL_SOURCE") == "" {
		slog.Info("[APEX_SERVER] Local source context disabled (set APEX_LOCAL_SOURCE=1 to enable)")
		return
	}
	slog.Info("[APEX_SERVER] Warming up project file cache...")
	start := time.Now()

	cache := make(map[string][]string)
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			// Skip huge/irrelevant directories during general indexing
			if info != nil && info.IsDir() {
				if info.Name() == "node_modules" || info.Name() == ".next" || info.Name() == ".git" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		base := filepath.Base(path)
		cache[base] = append(cache[base], path)
		return nil
	})

	s.fileCache = cache
	slog.Info("[APEX_SERVER] File cache warmed up", "entries", len(cache), "duration", time.Since(start))
}

func (s *Server) worker(id int) {
	ctx := context.Background()
	streamName := "apex_reports"
	offsetKey := "apex:stream:offset"

	// 1. Load the last processed offset from Redis to prevent re-processing history on restart
	lastID, err := s.rdb.Get(ctx, offsetKey).Result()
	if err != nil || lastID == "" {
		lastID = "0" // Fallback to beginning only if no offset is saved
		slog.Info("No persistent offset found, starting from 0", "worker_id", id)
	} else {
		slog.Info("Resuming stream from persistent offset", "worker_id", id, "offset", lastID)
	}

	for {
		// Block-read from the Redis Stream
		streams, err := s.rdb.XRead(ctx, &redis.XReadArgs{
			Streams: []string{streamName, lastID},
			Block:   0, // Wait indefinitely
			Count:   1,
		}).Result()

		if err != nil {
			slog.Error("Redis Read error", "worker_id", id, "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				lastID = message.ID // Track progress

				dataVal, ok := message.Values["data"]
				if !ok {
					slog.Warn("Message missing data field, skipping", "msg_id", message.ID)
					continue
				}
				dataStr, ok := dataVal.(string)
				if !ok {
					slog.Warn("Message data is not a string, skipping", "msg_id", message.ID)
					continue
				}

				projectIDVal, ok := message.Values["project_id"]
				if !ok {
					slog.Warn("Message missing project_id field, skipping", "msg_id", message.ID)
					continue
				}
				projectID, ok := projectIDVal.(string)
				if !ok {
					slog.Warn("Message project_id is not a string, skipping", "msg_id", message.ID)
					continue
				}

				report := &apex.CrashReport{}
				if err := proto.Unmarshal([]byte(dataStr), report); err != nil {
					slog.Error("Failed to unmarshal from Redis", "worker_id", id, "error", err)
					continue
				}

				// Normalize the ID so non-UUID agents don't break UUID-typed storage.
				report.ErrorId = normalizeErrorID(report.ErrorId)

				fingerprint := generateFingerprint(report.Message, scrubStackTrace(report.StackTrace))

				// Seed a fast heuristic insight so the HUD is never blank; the
				// async enrichment pass upgrades it to full Gemini analysis.
				if report.AiInsight == "" {
					report.AiInsight = s.recv.Analyze(report)
				}

				// Persist immediately — visibility no longer waits on AI.
				if err := s.store.SaveReport(report, projectID, fingerprint); err != nil {
					slog.Error("Failed to save report", "worker_id", id, "error", err, "project_id", projectID)
					dbSaveErrors.Inc()
				} else {
					slog.Info("Report persisted", "worker_id", id, "crash_id", report.ErrorId, "project_id", projectID)
					s.fireAlert(ctx, projectID, fingerprint, report)
					s.enqueueEnrichment(report, projectID, fingerprint)
				}

				// Advance the durable offset after handling the message.
				s.rdb.Set(ctx, offsetKey, message.ID, 0)
			}
		}
	}
}

var hexRegex = regexp.MustCompile(`0x[0-9a-fA-F]+`)

func scrubStackTrace(stackTrace string) string {
	// Remove hex memory addresses to ensure stable fingerprints for caching
	return hexRegex.ReplaceAllString(stackTrace, "<addr>")
}

func generateFingerprint(message, stackTrace string) string {
	h := sha256.New()
	h.Write([]byte(message))
	h.Write([]byte(stackTrace))
	return hex.EncodeToString(h.Sum(nil))
}

// getSourceContext attempts to extract file paths from a stack trace and read the source code.
func (s *Server) getSourceContext(stackTrace string) map[string]string {
	context := make(map[string]string)
	// Only attempt local file reads when explicitly enabled (see warmUpFileCache).
	if os.Getenv("APEX_LOCAL_SOURCE") == "" {
		return context
	}
	lines := strings.Split(stackTrace, "\n")

	for _, line := range lines {
		if !strings.Contains(line, ".go") && !strings.Contains(line, ".ts") && !strings.Contains(line, ".js") {
			continue
		}

		parts := strings.Fields(line)
		for _, part := range parts {
			cleanPath := strings.Trim(part, "(),: ")

			// Try 1: Exact or relative path check (fast)
			if _, err := os.Stat(cleanPath); err == nil {
				if s.readAndStore(cleanPath, context) {
					continue
				}
			}

			// Try 2: Handle paths with line numbers (e.g., "pkg/ai/ai.go:45")
			pathOnly := cleanPath
			if lastColon := strings.LastIndex(cleanPath, ":"); lastColon != -1 {
				pathOnly = cleanPath[:lastColon]
			}

			if _, err := os.Stat(pathOnly); err == nil {
				if s.readAndStore(pathOnly, context) {
					continue
				}
			}

			// Try 3: Use the pre-warmed file cache (very fast)
			base := filepath.Base(pathOnly)
			if paths, ok := s.fileCache[base]; ok {
				for _, p := range paths {
					if s.readAndStore(p, context) {
						break // Found a readable version of this file
					}
				}
			}
		}
		if len(context) >= 3 {
			break
		}
	}
	return context
}

func (s *Server) readAndStore(path string, context map[string]string) bool {
	if _, ok := context[path]; ok {
		return true
	}
	data, err := os.ReadFile(path)
	if err == nil {
		content := string(data)
		if len(content) > 2048 {
			content = content[:2048]
		}
		context[path] = content
		return true
	}
	return false
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
	slog.Info("Batch received", "count", len(batch.Reports), "persistent", s.isPersistent)
	for _, report := range batch.Reports {
		reportsReceived.Inc()
		report.ErrorId = normalizeErrorID(report.ErrorId)
		fingerprint := generateFingerprint(report.Message, scrubStackTrace(report.StackTrace))
		if report.AiInsight == "" {
			report.AiInsight = s.recv.Analyze(report)
		}

		if !s.isPersistent {
			// Direct Save in Memory Mode (Bypass Redis/Workers for quick local feedback)
			if err := s.store.SaveReport(report, projectID, fingerprint); err != nil {
				slog.Error("Failed to save report (Memory Mode)", "error", err, "crash_id", report.ErrorId)
				dbSaveErrors.Inc()
			} else {
				slog.Info("Report saved directly (Memory Mode)", "crash_id", report.ErrorId)
				s.enqueueEnrichment(report, projectID, fingerprint)
			}
			continue
		}

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
			// Last resort: if Redis fails and we have a store, try saving directly
			if err := s.store.SaveReport(report, projectID, fingerprint); err != nil {
				dbSaveErrors.Inc()
			} else {
				s.fireAlert(ctx, projectID, fingerprint, report)
				s.enqueueEnrichment(report, projectID, fingerprint)
			}
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

	// If the TUI connects using an API Key instead of a query parameter
	if projectID == "" {
		key := r.Header.Get("X-Apex-API-Key")
		if key != "" {
			resolved, err := s.store.ValidateKey(key)
			if err == nil && resolved != "" {
				projectID = resolved
				slog.Info("[APEX_DEBUG] Resolved project from key", "project_id", projectID)
			} else {
				slog.Warn("[APEX_DEBUG] Key validation failed or no project found", "error", err)
			}
		}
	}

	reports, err := s.store.GetReports(50, projectID)
	if err != nil {
		slog.Error("Failed to fetch reports", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	slog.Info("[APEX_DEBUG] Reports fetched", "count", len(reports), "project_id", projectID)

	w.Header().Set("Content-Type", "application/json")
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

	// Rate Limit: 10 chats per hour per report_id (proxy for project/user)
	allowed, _ := s.limiter.Allow(r.Context(), req.ReportID+":chat", 10, 1*time.Hour)
	if !allowed {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{"response": "RATE_LIMIT: Quota exceeded for this session."})
		return
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	// Get context for the report if ID is provided
	sourceContext := make(map[string]string)
	if req.ReportID != "" {
		reports, err := s.store.GetReports(100, "")
		if err == nil {
			for _, r := range reports {
				if r.ErrorId == req.ReportID {
					sourceContext = s.getSourceContext(r.StackTrace)
					break
				}
			}
		}
	}

	iter, err := s.ai.ChatStream(ctx, req.Message, req.ReportID, sourceContext)
	if err != nil {
		slog.Error("Failed to start AI stream", "error", err)
		fmt.Fprintf(w, "data: SIGNAL_LOSS: %v\n\n", err)
		flusher.Flush()
		return
	}

	slog.Info("[APEX_DEBUG] Starting AI stream generation", "report_id", req.ReportID)

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			slog.Error("Stream iteration error", "error", err)
			fmt.Fprintf(w, "data: SIGNAL_NOISE: %v\n\n", err)
			break
		}

		if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
			for _, part := range resp.Candidates[0].Content.Parts {
				text := fmt.Sprintf("%v", part)
				// Format as SSE data
				fmt.Fprintf(w, "data: %s\n\n", text)
				flusher.Flush()
			}
		}
	}

	// Signal end of stream
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
	slog.Info("[APEX_DEBUG] AI stream generation complete", "report_id", req.ReportID)
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
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
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

func (s *Server) handleResolveReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing report ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Resolved bool `json:"resolved"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := s.store.ResolveReport(id, req.Resolved); err != nil {
		slog.Error("Failed to resolve report", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing project ID", http.StatusBadRequest)
		return
	}

	if err := s.store.DeleteProject(id); err != nil {
		slog.Error("Failed to delete project", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"persistent": s.isPersistent,
		"status":     "operational",
		"timestamp":  time.Now().Unix(),
	})
}

// handleGetIssues returns deduplicated crash groups (server-side grouping) with
// occurrence counts, first/last-seen, and pagination.
func (s *Server) handleGetIssues(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		if key := r.Header.Get("X-Apex-API-Key"); key != "" {
			if resolved, err := s.store.ValidateKey(key); err == nil && resolved != "" {
				projectID = resolved
			}
		}
	}

	limit := atoiDefault(r.URL.Query().Get("limit"), 50)
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := atoiDefault(r.URL.Query().Get("offset"), 0)
	if offset < 0 {
		offset = 0
	}

	issues, err := s.store.GetIssues(projectID, limit, offset)
	if err != nil {
		slog.Error("Failed to fetch issues", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issues)
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// seedDemo populates a read-only demo project so first-time visitors see a
// live-looking HUD (grouped issues, sparkline, AI panels) instead of emptiness.
// Idempotent: skips if the demo project already exists. Enable with APEX_SEED_DEMO=1.
func (s *Server) seedDemo() {
	if os.Getenv("APEX_SEED_DEMO") != "1" {
		return
	}
	const demoUser = "demo"
	const demoKey = "apex_demo-public-ingest-key"

	if existing, _ := s.store.GetProjects(demoUser); len(existing) > 0 {
		for _, p := range existing {
			if p.IngestKey == demoKey {
				slog.Info("[APEX_SERVER] Demo project already present", "id", p.ID)
				return
			}
		}
	}

	p := &storage.Project{
		ID: uuid.New().String(), UserID: demoUser, Name: "Demo · Payments API",
		IngestKey: demoKey, CreatedAt: time.Now(),
	}
	if err := s.store.SaveProject(p); err != nil {
		slog.Error("Demo seed: project save failed", "error", err)
		return
	}

	now := time.Now().Unix()
	type sample struct {
		msg, stack, os, arch string
		ageSec               int64
		dupes                int
	}
	samples := []sample{
		{"runtime error: invalid memory address or nil pointer dereference", "main.(*Handler).Serve(0x0)\n\thandler.go:88 +0x1f\nnet/http.(*conn).serve()\n\tserver.go:1995", "linux", "amd64", 120, 3},
		{"runtime error: index out of range [5] with length 3", "pkg/engine/processor.go:77\n\tpkg/engine/worker.go:12\nmain.go:104", "darwin", "arm64", 900, 1},
		{"context deadline exceeded: upstream payments-gateway", "client.(*Conn).roundTrip()\n\tclient.go:240\nbilling/charge.go:51", "linux", "amd64", 3600, 2},
	}
	for _, sm := range samples {
		for i := 0; i < sm.dupes; i++ {
			r := &apex.CrashReport{
				ErrorId:    uuid.New().String(),
				Message:    sm.msg,
				StackTrace: sm.stack,
				Timestamp:  now - sm.ageSec - int64(i*30),
				Context:    &apex.DeviceContext{Os: sm.os, Arch: sm.arch, TotalMemory: 16 << 30, FreeMemory: 4 << 30, BatteryLevel: 76},
			}
			r.AiInsight = s.recv.Analyze(r)
			fp := generateFingerprint(r.Message, scrubStackTrace(r.StackTrace))
			if err := s.store.SaveReport(r, p.ID, fp); err != nil {
				slog.Error("Demo seed: report save failed", "error", err)
			}
		}
	}
	slog.Info("[APEX_SERVER] Demo project seeded", "id", p.ID, "ingest_key", demoKey)
}

// startRetention periodically prunes reports older than APEX_RETENTION_DAYS
// (default 30) to keep storage bounded on free tiers.
func (s *Server) startRetention() {
	days := atoiDefault(os.Getenv("APEX_RETENTION_DAYS"), 30)
	if days <= 0 {
		slog.Info("[APEX_SERVER] Retention disabled (APEX_RETENTION_DAYS<=0)")
		return
	}
	prune := func() {
		cutoff := time.Now().AddDate(0, 0, -days)
		if removed, err := s.store.PruneReports(cutoff); err != nil {
			slog.Error("Retention prune failed", "error", err)
		} else if removed > 0 {
			slog.Info("Retention prune complete", "removed", removed, "older_than_days", days)
		}
	}
	prune() // run once at boot
	ticker := time.NewTicker(6 * time.Hour)
	for range ticker.C {
		prune()
	}
}

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load configuration
	godotenv.Load() // Root .env
	envLocalPath := filepath.Join("dashboard", ".env.local")
	if _, err := os.Stat(envLocalPath); err == nil {
		if err := godotenv.Load(envLocalPath); err == nil {
			slog.Info("Loaded configuration from dashboard/.env.local")
		}
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// CockroachDB default insecure connection
		connStr = "postgresql://root@127.0.0.1:5433/defaultdb?sslmode=disable"
	}

	var store storage.Provider
	var err error

	// Production Resilience: Retry connecting to DB
	slog.Info("Attempting to connect to database...", "url_set", os.Getenv("DATABASE_URL") != "")
	for i := 0; i < 5; i++ {
		pgStore, pgErr := storage.NewPostgres(connStr)
		if pgErr == nil {
			if err = pgStore.Initialize(); err == nil {
				store = pgStore
				slog.Info("Database initialized successfully")
				break
			} else {
				slog.Error("Database initialization failed", "attempt", i+1, "error", err)
			}
		} else {
			slog.Warn("Database connection failed", "attempt", i+1, "error", pgErr)
		}
		time.Sleep(2 * time.Second)
	}

	if store == nil {
		if os.Getenv("DATABASE_URL") != "" {
			slog.Error("CRITICAL: DATABASE_URL is set but connection FAILED. Falling back to Memory Mode.")
		} else {
			slog.Info("DATABASE_URL not set, using default Memory Mode.")
		}
		store = storage.NewMemoryStore()
	} else {
		slog.Info("Persistent Vault Storage Online")
	}

	// Initialize Redis for Ingest Buffering
	rdbAddr := os.Getenv("REDIS_URL")
	if rdbAddr == "" {
		rdbAddr = "127.0.0.1:6379"
	}

	var rdb *redis.Client
	if len(rdbAddr) > 8 && (rdbAddr[:8] == "redis://" || rdbAddr[:9] == "rediss://") {
		opt, err := redis.ParseURL(rdbAddr)
		if err != nil {
			slog.Error("Failed to parse Redis URL", "url", rdbAddr, "error", err)
			rdb = redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
		} else {
			rdb = redis.NewClient(opt)
			slog.Info("Redis connection initialized via URL", "addr", opt.Addr)
		}
	} else {
		rdb = redis.NewClient(&redis.Options{
			Addr: rdbAddr,
		})
		slog.Info("Redis connection initialized via Addr", "addr", rdbAddr)
	}

	// Verify Redis connection
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		slog.Error("Redis connection verification failed", "error", err)
	} else {
		slog.Info("Redis connection verified")
	}

	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		slog.Warn("GEMINI_API_KEY not set. AI unit will be in degraded mode.")
	}

	srv := NewServer(store, rdb, geminiKey)
	if store != nil {
		// Only mark as persistent if it's not the internal MemoryStore
		if _, ok := store.(*storage.MemoryStore); !ok {
			srv.isPersistent = true
		}
	}

	// Background retention to keep storage bounded (esp. on free DB tiers).
	go srv.startRetention()

	// Optional demo data so a fresh deployment isn't an empty HUD.
	srv.seedDemo()

	// Route definitions with CORS
	http.HandleFunc("/ingest", corsMiddleware(srv.handleIngest))
	http.HandleFunc("/reports", corsMiddleware(srv.handleGetReports))
	http.HandleFunc("/api/reports", corsMiddleware(srv.internalGate(srv.handleGetReportsJSON)))
	http.HandleFunc("/api/issues", corsMiddleware(srv.internalGate(srv.handleGetIssues)))
	http.HandleFunc("/api/chat", corsMiddleware(srv.handleChat))
	http.HandleFunc("/api/projects", corsMiddleware(srv.internalGate(srv.handleGetProjects)))
	http.HandleFunc("/api/projects/create", corsMiddleware(dashboardGate(srv.handleCreateProject)))
	http.HandleFunc("/api/projects/delete", corsMiddleware(dashboardGate(srv.handleDeleteProject)))
	http.HandleFunc("/api/reports/resolve", corsMiddleware(dashboardGate(srv.handleResolveReport)))
	http.HandleFunc("/api/status", corsMiddleware(srv.handleStatus))
	http.Handle("/metrics", promhttp.Handler())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	slog.Info("Apex Production Receiver starting", "port", port)
	http.ListenAndServe(":"+port, nil)
}

// resolveAllowedOrigin echoes the request Origin when it is permitted by
// APEX_ALLOWED_ORIGINS (comma-separated). When that env is empty or "*",
// every origin is allowed (backward-compatible default for the public demo).
func resolveAllowedOrigin(reqOrigin string) string {
	allowed := strings.TrimSpace(os.Getenv("APEX_ALLOWED_ORIGINS"))
	if allowed == "" || allowed == "*" {
		return "*"
	}
	for _, o := range strings.Split(allowed, ",") {
		if strings.EqualFold(strings.TrimSpace(o), reqOrigin) {
			return reqOrigin
		}
	}
	return "" // not allowed
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := resolveAllowedOrigin(r.Header.Get("Origin"))
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if origin != "*" {
				w.Header().Set("Vary", "Origin")
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Apex-API-Key, X-Apex-Dashboard-Key")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// internalGate optionally protects read endpoints. When APEX_INTERNAL_SECRET is
// set, callers must present it via X-Apex-Internal-Key (the Next.js BFF injects
// it server-side) OR a valid ingest key (for the TUI/agents). When unset, reads
// stay open so the existing demo deployment keeps working unchanged.
func (s *Server) internalGate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		secret := strings.TrimSpace(os.Getenv("APEX_INTERNAL_SECRET"))
		if secret == "" || r.Method == http.MethodOptions {
			next(w, r)
			return
		}
		if r.Header.Get("X-Apex-Internal-Key") == secret {
			next(w, r)
			return
		}
		if key := r.Header.Get("X-Apex-API-Key"); key != "" {
			if pid, err := s.store.ValidateKey(key); err == nil && pid != "" {
				next(w, r)
				return
			}
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
}

// dashboardGate optionally protects mutating dashboard endpoints. When
// APEX_DASHBOARD_SECRET is set, callers must present it via the
// X-Apex-Dashboard-Key header. When unset, the gate is a no-op so existing
// open/demo deployments keep working unchanged.
func dashboardGate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		secret := strings.TrimSpace(os.Getenv("APEX_DASHBOARD_SECRET"))
		if secret != "" && r.Method != http.MethodOptions {
			if r.Header.Get("X-Apex-Dashboard-Key") != secret {
				http.Error(w, "Unauthorized: invalid dashboard key", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}
