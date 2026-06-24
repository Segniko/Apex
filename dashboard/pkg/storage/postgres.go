package storage

import (
	"database/sql"

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
	ALTER TABLE crash_reports ADD COLUMN IF NOT EXISTS ai_insight TEXT;
	CREATE INDEX IF NOT EXISTS idx_reports_created_at ON crash_reports(created_at);
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

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) GetDB() *sql.DB {
	if s == nil {
		return nil
	}
	return s.db
}
