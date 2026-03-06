package agent

import (
	"fmt"
	"runtime"
	"time"

	"github.com/apex/monitor/pkg/syphon"
	"github.com/apex/monitor/pkg/vault"
	apex "github.com/apex/monitor/proto"
	"github.com/google/uuid"
)

// Agent is the main entry point for the Apex monitoring system.
type Agent struct {
	vault       *vault.Vault
	syphon      *syphon.Syphon
	config      Config
	stackBuffer []byte
	stopChan    chan struct{}
}

// New creates a new instance of the Apex Agent and starts the background sync.
func New(v *vault.Vault, s *syphon.Syphon, cfg Config) *Agent {
	a := &Agent{
		vault:       v,
		syphon:      s,
		config:      cfg,
		stackBuffer: make([]byte, 1024*8),
		stopChan:    make(chan struct{}),
	}

	go a.syncLoop()

	return a
}

// CapturePanic is a helper that should be used with 'defer'.
// It intercepts any crash and prepares it for the Vault.
func (a *Agent) CapturePanic() {
	if r := recover(); r != nil {
		n := runtime.Stack(a.stackBuffer, false)
		stackTrace := string(a.stackBuffer[:n])

		report := &apex.CrashReport{
			ErrorId:    uuid.NewString(),
			Message:    fmt.Sprintf("%v", r),
			StackTrace: stackTrace,
			Timestamp:  time.Now().Unix(),
			Context:    a.collectContext(),
		}

		if a.vault != nil {
			a.vault.Save(report)
		}
	}
}

// collectContext gathers telemetry about the system.
func (a *Agent) collectContext() *apex.DeviceContext {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &apex.DeviceContext{
		Os:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		TotalMemory:  int64(m.Sys),
		FreeMemory:   int64(m.HeapIdle),
		BatteryLevel: 1.0,
	}
}

// Stop gracefully shuts down the Agent's background processes.
func (a *Agent) Stop() {
	close(a.stopChan)
}

// syncLoop runs in the background and periodically attempts to sync data via the Syphon.
func (a *Agent) syncLoop() {
	ticker := time.NewTicker(a.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.attemptSync()
		case <-a.stopChan:
			return
		}
	}
}

func (a *Agent) attemptSync() {
	if a.syphon == nil || a.vault == nil {
		return
	}

	if !a.syphon.ShouldSync() {
		return
	}

	reports, err := a.vault.FetchAll()
	if err != nil || len(reports) == 0 {
		return
	}

	// Limit reports to batch size
	if len(reports) > a.config.BatchSize {
		reports = reports[:a.config.BatchSize]
	}

	err = a.syphon.SendBatch(reports, a.config.IngestURL, a.config.APIKey)
	if err == nil {
		// Cleanup old reports on success
		maxTs := reports[len(reports)-1].Timestamp
		a.vault.Cleanup(maxTs + 1)
	}
}
