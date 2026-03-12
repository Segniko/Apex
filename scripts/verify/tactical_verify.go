package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/klauspost/compress/zstd"
	apex "github.com/apex/monitor/proto"
	"google.golang.org/protobuf/proto"
)

func main() {
	// 1. Check for Ingest Key
	key := os.Getenv("X_APEX_KEY")
	if key == "" {
		fmt.Println("❌ ERROR: X_APEX_KEY environment variable is not set.")
		fmt.Println("Usage (CMD): set X_APEX_KEY=your_key && go run scripts/verify/tactical_verify.go")
		return
	}

	fmt.Printf("📡 Initializing Tactical Ingest [Key: %s...]\n", key[:8])

	// 2. Build a Forensic Report with a VALID UUID
	report := &apex.CrashReport{
		ErrorId:    uuid.New().String(), // MUST BE A VALID UUID
		Message:    "Manual Tactical Verification: Runtime Error in Null Pointer Dereference",
		StackTrace: "main.go:42\nruntime.go:101\nkernel.go:05",
		Timestamp:  time.Now().Unix(),
		Context: &apex.DeviceContext{
			Os:           "windows",
			Arch:         "amd64",
			BatteryLevel: 99,
			FreeMemory:   8192,
		},
	}

	batch := &apex.BatchReport{
		Reports: []*apex.CrashReport{report},
	}

	// 3. Serialize and Compress (Zstd)
	data, err := proto.Marshal(batch)
	if err != nil {
		fmt.Printf("❌ Failed to marshal: %v\n", err)
		return
	}

	var buf bytes.Buffer
	zw, _ := zstd.NewWriter(&buf)
	zw.Write(data)
	zw.Close()

	// 4. Transmit to Local Dockerized Receiver
	url := "http://localhost:8081/ingest"
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("X-Apex-API-Key", key)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Transmission failed: %v\n", err)
		fmt.Println("Check if Docker containers are running (docker-compose ps)")
		return
	}
	defer resp.Body.Close()

	// 5. Audit Response
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusAccepted {
		fmt.Println("✅ SUCCESS: Tactical Batch Accepted!")
		fmt.Printf("Server Response: %s\n", string(body))
		fmt.Println("Go to http://localhost:3000 to view your AI Forensic HUD.")
	} else {
		fmt.Printf("❌ FAILED: Status %d\n", resp.StatusCode)
		fmt.Printf("Server Response: %s\n", string(body))
	}
}
