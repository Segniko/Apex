package storage

import (
	"database/sql"
	"time"

	apex "github.com/Segniko/Apex/proto"
	_ "github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

func NewPostgres(connStr string) (*Store, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Connection Pooling for Stability
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Initialize() error {
	// 1. Create base tables
	tables := `
	CREATE TABLE IF NOT EXISTS crash_reports (
		id UUID PRIMARY KEY,
		message TEXT,
		stack_trace TEXT,
		os TEXT,
		arch TEXT,
		total_memory BIGINT,
		free_memory BIGINT,
		battery_level FLOAT,
		ai_insight TEXT,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS projects (
		id UUID PRIMARY KEY,
		user_id TEXT,
		name TEXT,
		ingest_key TEXT UNIQUE,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := s.db.Exec(tables); err != nil {
		return err
	}

	// 2. Ensure project_id and resolved columns exist (Manual Migration)
	if _, err := s.db.Exec("ALTER TABLE crash_reports ADD COLUMN IF NOT EXISTS project_id UUID"); err != nil {
		return err
	}
	if _, err := s.db.Exec("ALTER TABLE crash_reports ADD COLUMN IF NOT EXISTS resolved BOOLEAN DEFAULT FALSE"); err != nil {
		return err
	}
	// Stable signature used to group identical crashes into issues.
	if _, err := s.db.Exec("ALTER TABLE crash_reports ADD COLUMN IF NOT EXISTS fingerprint TEXT"); err != nil {
		return err
	}

	// 3. Create indices
	indices := `
	CREATE INDEX IF NOT EXISTS idx_reports_created_at ON crash_reports(created_at);
	CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);
	CREATE INDEX IF NOT EXISTS idx_reports_project_id ON crash_reports(project_id);
	CREATE INDEX IF NOT EXISTS idx_reports_fingerprint ON crash_reports(fingerprint);
	`
	if _, err := s.db.Exec(indices); err != nil {
		return err
	}

	return nil
}

func (s *Store) SaveReport(r *apex.CrashReport, projectID string, fingerprint string) error {
	query := `
	INSERT INTO crash_reports (id, project_id, message, stack_trace, os, arch, total_memory, free_memory, battery_level, ai_insight, resolved, fingerprint, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, to_timestamp($13))
	`
	ctx := r.Context
	if ctx == nil {
		ctx = &apex.DeviceContext{}
	}
	_, err := s.db.Exec(query,
		r.ErrorId,
		projectID,
		r.Message,
		r.StackTrace,
		ctx.Os,
		ctx.Arch,
		ctx.TotalMemory,
		ctx.FreeMemory,
		ctx.BatteryLevel,
		r.AiInsight,
		r.Resolved,
		fingerprint,
		r.Timestamp,
	)
	return err
}

// GetIssues returns deduplicated crash groups for a project, ordered by most
// recently seen, with occurrence counts and first/last-seen timestamps.
func (s *Store) GetIssues(projectID string, limit int, offset int) ([]*Issue, error) {
	query := `
	WITH ranked AS (
		SELECT
			COALESCE(NULLIF(fingerprint, ''), id::text) AS fp,
			id, message, stack_trace, COALESCE(ai_insight, '') AS ai_insight, resolved, project_id,
			EXTRACT(EPOCH FROM created_at)::BIGINT AS ts,
			ROW_NUMBER() OVER (
				PARTITION BY COALESCE(NULLIF(fingerprint, ''), id::text)
				ORDER BY created_at DESC
			) AS rn
		FROM crash_reports
		WHERE ($1 = '' OR project_id::text = $1)
	),
	agg AS (
		SELECT fp,
			COUNT(*) AS cnt,
			MIN(ts) AS first_seen,
			MAX(ts) AS last_seen,
			BOOL_AND(resolved) AS all_resolved
		FROM ranked GROUP BY fp
	)
	SELECT r.fp, r.project_id::text, r.message, r.stack_trace, r.ai_insight, r.id::text,
	       a.cnt, a.first_seen, a.last_seen, a.all_resolved
	FROM ranked r
	JOIN agg a ON a.fp = r.fp
	WHERE r.rn = 1
	ORDER BY a.last_seen DESC
	LIMIT $2 OFFSET $3
	`
	rows, err := s.db.Query(query, projectID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := make([]*Issue, 0)
	for rows.Next() {
		iss := &Issue{}
		if err := rows.Scan(
			&iss.Fingerprint, &iss.ProjectID, &iss.Message, &iss.StackTrace, &iss.AiInsight,
			&iss.LatestID, &iss.Count, &iss.FirstSeen, &iss.LastSeen, &iss.Resolved,
		); err != nil {
			return nil, err
		}
		issues = append(issues, iss)
	}
	return issues, nil
}

// PruneReports deletes reports created before the cutoff, returning the count removed.
func (s *Store) PruneReports(olderThan time.Time) (int64, error) {
	res, err := s.db.Exec("DELETE FROM crash_reports WHERE created_at < $1", olderThan)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// UpdateInsight writes the AI forensic analysis back onto an already-saved
// report. Used by the async enrichment pass so persistence isn't blocked on AI.
func (s *Store) UpdateInsight(reportID, insight string) error {
	_, err := s.db.Exec("UPDATE crash_reports SET ai_insight = $1 WHERE id = $2", insight, reportID)
	return err
}

func (s *Store) SaveProject(p *Project) error {
	query := `
	INSERT INTO projects (id, user_id, name, ingest_key)
	VALUES ($1, $2, $3, $4)
	`
	_, err := s.db.Exec(query, p.ID, p.UserID, p.Name, p.IngestKey)
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) GetDB() *sql.DB {
	if s == nil {
		return nil
	}
	return s.db
}
func (s *Store) GetSimilarReports(message string, limit int, projectID string) ([]*apex.CrashReport, error) {
	// Simple text similarity using ILIKE and prefix matching for RAG context
	query := `
	SELECT id, message, stack_trace, os, arch, total_memory, free_memory, battery_level, COALESCE(ai_insight, ''), resolved, project_id::text, EXTRACT(EPOCH FROM created_at)::BIGINT
	FROM crash_reports 
	WHERE project_id::text = $2 AND (message ILIKE $3 OR stack_trace ILIKE $3)
	ORDER BY created_at DESC 
	LIMIT $1
	`
	// Match first 15 chars for similarity
	pattern := "%"
	if len(message) > 15 {
		pattern = "%" + message[:15] + "%"
	} else if len(message) > 0 {
		pattern = "%" + message + "%"
	}

	rows, err := s.db.Query(query, limit, projectID, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reports := make([]*apex.CrashReport, 0)
	for rows.Next() {
		r := &apex.CrashReport{Context: &apex.DeviceContext{}}
		var timestamp int64
		err := rows.Scan(
			&r.ErrorId,
			&r.Message,
			&r.StackTrace,
			&r.Context.Os,
			&r.Context.Arch,
			&r.Context.TotalMemory,
			&r.Context.FreeMemory,
			&r.Context.BatteryLevel,
			&r.AiInsight,
			&r.Resolved,
			&r.ProjectId,
			&timestamp,
		)
		if err != nil {
			return nil, err
		}
		r.Timestamp = timestamp
		reports = append(reports, r)
	}

	return reports, nil
}
