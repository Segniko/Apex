package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"time"

	apex "github.com/apex/monitor/proto"
	"google.golang.org/protobuf/proto"
)

func main() {
	report := &apex.CrashReport{
		ErrorId:    "test-error-123",
		Message:    "Manual Test Error",
		StackTrace: "main.go:10\nmain.go:20",
		Timestamp:  time.Now().Unix(),
		Context: &apex.Context{
			Os:           "windows",
			Arch:         "amd64",
			BatteryLevel: 85,
			FreeMemory:   4096,
		},
	}

	batch := &apex.CrashBatch{
		Reports: []*apex.CrashReport{report},
	}

	data, _ := proto.Marshal(batch)
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write(data)
	gw.Close()

	req, _ := http.NewRequest("POST", "http://localhost:8081/ingest", &buf)
	req.Header.Set("X-Apex-API-Key", "test-key") // Note: Replace with actual key if needed
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %s\nBody: %s\n", resp.Status, string(body))
}
