package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Segniko/Apex/pkg/agent"
	"github.com/Segniko/Apex/pkg/syphon"
	"github.com/Segniko/Apex/pkg/vault"
)

type MockNetwork struct {
	Type syphon.NetworkType
}

func (m *MockNetwork) CurrentType() syphon.NetworkType {
	return m.Type
}

func main() {
	key := []byte("this-is-a-32-byte-secret-key-!!!")

	v, err := vault.New("apex.db", key)
	if err != nil {
		fmt.Printf("Vault error: %v\n", err)
		return
	}
	defer v.Close()

	net := &MockNetwork{Type: syphon.NetworkWifi}
	s, _ := syphon.New(net)

	cfg := agent.DefaultConfig()
	cfg.IngestURL = "http://localhost:8081/ingest"
	cfg.APIKey = "apex-prod-key-12345"
	cfg.SyncInterval = 5 * time.Second

	a := agent.New(v, s, cfg)
	defer a.Stop()

	fmt.Println("=== Apex Production Agent Running ===")
	fmt.Printf("Monitoring started. Crashes will sync every %v to %s\n", cfg.SyncInterval, cfg.IngestURL)

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nApex Agent shutting down...")
}

func simulateCrash(a *agent.Agent, msg string) {
	defer a.CapturePanic()
	panic(msg)
}
