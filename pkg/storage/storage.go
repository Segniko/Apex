package storage

import (
	"time"

	apex "github.com/apex/monitor/proto"
)

// Project represents a workspace in Apex.
type Project struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	IngestKey string    `json:"ingest_key"`
	CreatedAt time.Time `json:"created_at"`
}

// Provider is the interface for all storage implementations.
type Provider interface {
	SaveReport(r *apex.CrashReport, projectID string) error
	GetReports(limit int, projectID string) ([]*apex.CrashReport, error)

	SaveProject(p *Project) error
	GetProjects(userID string) ([]*Project, error)
	ValidateKey(key string) (string, error)
}
