package storage

import (
	"database/sql"

	apex "github.com/apex/monitor/proto"
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

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Initialize() error {
	schema := `
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

	CREATE INDEX IF NOT EXISTS idx_reports_created_at ON crash_reports(created_at);
	CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) SaveReport(r *apex.CrashReport) error {
	query := `
	INSERT INTO crash_reports (id, message, stack_trace, os, arch, total_memory, free_memory, battery_level, ai_insight, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, to_timestamp($10))
	`
	_, err := s.db.Exec(query,
		r.ErrorId,
		r.Message,
		r.StackTrace,
		r.Context.Os,
		r.Context.Arch,
		r.Context.TotalMemory,
		r.Context.FreeMemory,
		r.Context.FreeMemory,
		r.Context.BatteryLevel,
		r.AiInsight,
		r.Timestamp,
	)
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
