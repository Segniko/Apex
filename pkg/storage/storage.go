package storage

import (
	"time"

	apex "github.com/Segniko/Apex/proto"
)

// Issue is a deduplicated group of crash events sharing a fingerprint.
type Issue struct {
	Fingerprint string `json:"fingerprint"`
	ProjectID   string `json:"project_id"`
	Message     string `json:"message"`
	StackTrace  string `json:"stack_trace"`
	AiInsight   string `json:"ai_insight"`
	Count       int64  `json:"count"`
	FirstSeen   int64  `json:"first_seen"`
	LastSeen    int64  `json:"last_seen"`
	Resolved    bool   `json:"resolved"`
	LatestID    string `json:"error_id"`
}

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
	SaveReport(r *apex.CrashReport, projectID string, fingerprint string) error
	UpdateInsight(reportID string, insight string) error
	GetReports(limit int, projectID string) ([]*apex.CrashReport, error)
	GetIssues(projectID string, limit int, offset int) ([]*Issue, error)
	PruneReports(olderThan time.Time) (int64, error)

	SaveProject(p *Project) error
	GetProjects(userID string) ([]*Project, error)
	ValidateKey(key string) (string, error)
	GetSimilarReports(message string, limit int, projectID string) ([]*apex.CrashReport, error)
	ResolveReport(id string, resolved bool) error
	DeleteProject(id string) error
}
