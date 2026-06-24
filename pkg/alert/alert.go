// Package alert delivers crash notifications to an external webhook
// (Slack, Discord, or any generic JSON endpoint). It is intentionally
// dependency-free and fails open: a misconfigured or down webhook never
// blocks the ingest pipeline.
package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Notifier posts crash alerts to a configured webhook URL.
type Notifier struct {
	webhookURL   string
	dashboardURL string
	client       *http.Client
}

// New builds a Notifier. If webhookURL is empty the notifier is disabled
// and Notify becomes a no-op.
func New(webhookURL, dashboardURL string) *Notifier {
	if dashboardURL == "" {
		dashboardURL = "https://apex-addis.vercel.app"
	}
	return &Notifier{
		webhookURL:   strings.TrimSpace(webhookURL),
		dashboardURL: strings.TrimRight(dashboardURL, "/"),
		client:       &http.Client{Timeout: 8 * time.Second},
	}
}

// Enabled reports whether a webhook destination is configured.
func (n *Notifier) Enabled() bool { return n != nil && n.webhookURL != "" }

// Crash carries the minimal fields needed to render an alert.
type Crash struct {
	ProjectID string
	ErrorID   string
	Message   string
	Stack     string
	Insight   string
}

// Notify sends a single crash alert. Safe to call in a goroutine; errors are
// logged, not returned, so callers never have to handle webhook failures.
func (n *Notifier) Notify(ctx context.Context, c Crash) {
	if !n.Enabled() {
		return
	}

	firstFrame := c.Message
	if line := firstLine(c.Stack); line != "" {
		firstFrame = line
	}

	link := fmt.Sprintf("%s/dashboard/projects/%s", n.dashboardURL, c.ProjectID)
	text := fmt.Sprintf("🚨 *Apex crash captured*\n*%s*\n`%s`\n<%s|Open forensics HUD>",
		truncate(c.Message, 200), truncate(firstFrame, 200), link)

	// Slack uses "text", Discord uses "content". Sending both keys keeps the
	// same payload compatible with either destination (extra keys are ignored).
	payload := map[string]any{
		"text":    text,
		"content": fmt.Sprintf("🚨 Apex crash: %s\n%s\n%s", truncate(c.Message, 200), truncate(firstFrame, 200), link),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("alert: marshal failed", "error", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, bytes.NewReader(body))
	if err != nil {
		slog.Error("alert: request build failed", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		slog.Warn("alert: webhook delivery failed", "error", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		slog.Warn("alert: webhook returned non-2xx", "status", resp.StatusCode)
		return
	}
	slog.Info("alert: crash notification delivered", "project_id", c.ProjectID, "error_id", c.ErrorID)
}

func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			return t
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
