package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/klauspost/compress/zstd"
)

type SampleReport struct {
	ErrorID    string            `json:"error_id"`
	Message    string            `json:"message"`
	StackTrace string            `json:"stack_trace"`
	Timestamp  int64             `json:"timestamp"`
	Context    map[string]string `json:"context"`
	OrgID      string            `json:"org_id"`
	ProjectID  string            `json:"project_id"`
}

func main() {
	url := "http://localhost:8081/ingest"

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".apex_config.json")
	apiKey := "apex-demo-key-999"
	if data, err := os.ReadFile(configPath); err == nil {
		var cfg struct {
			APIKey string `json:"api_key"`
		}
		json.Unmarshal(data, &cfg)
		if cfg.APIKey != "" {
			apiKey = cfg.APIKey
		}
	}

	samples := []SampleReport{
		{
			ErrorID:    "e101-db-sync",
			Message:    "database connection refused: sentinel node unavailable",
			StackTrace: "main.go:42\nstorage/postgres.go:128\nruntime/proc.go:250",
			Timestamp:  time.Now().Unix(),
			Context:    map[string]string{"env": "production", "service": "billing-api", "version": "v1.2.4"},
			OrgID:      "apex-global",
			ProjectID:  "main-cluster",
		},
		{
			ErrorID:    "e202-null-ptr",
			Message:    "runtime error: invalid memory address or nil pointer dereference",
			StackTrace: "pkg/engine/processor.go:77\npkg/engine/worker.go:12\nmain.go:104",
			Timestamp:  time.Now().Unix(),
			Context:    map[string]string{"env": "staging", "service": "data-transformer", "arch": "arm64"},
			OrgID:      "apex-global",
			ProjectID:  "staging-box",
		},
		{
			ErrorID:    "e303-auth-fail",
			Message:    "security breach: invalid token signature detected on auth-v3",
			StackTrace: "middleware/auth.go:88\nrouter/router.go:50\nmain.go:22",
			Timestamp:  time.Now().Unix(),
			Context:    map[string]string{"env": "production", "service": "auth-node", "ip": "192.168.1.104"},
			OrgID:      "apex-global",
			ProjectID:  "main-cluster",
		},
	}

	fmt.Println("Starting Apex Discovery Simulation...")

	client := &http.Client{Timeout: 5 * time.Second}

	// Wrap in BatchReport structure
	batch := map[string]interface{}{
		"reports": samples,
	}

	payload, _ := json.Marshal(batch)

	// Compress with ZSTD (Tactical Requirement)
	var buf bytes.Buffer
	zw, _ := zstd.NewWriter(&buf)
	zw.Write(payload)
	zw.Close()

	req, _ := http.NewRequest("POST", url, &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Apex-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("FAILED_SYNC: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("SUCCESS: Simulation batch synced using Key: %s...\n", apiKey[:8])
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("REJECTED: Server returned %d - %s\n", resp.StatusCode, string(body))
	}

	fmt.Println("\n Simulation Complete. Check your dashboard at http://localhost:3000/dashboard")
}
