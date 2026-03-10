package storage

import (
	apex "github.com/apex/monitor/proto"
)

func (s *Store) GetReports(limit int) ([]*apex.CrashReport, error) {
	query := `
	SELECT id, message, stack_trace, os, arch, total_memory, free_memory, battery_level, COALESCE(ai_insight, ''), EXTRACT(EPOCH FROM created_at)::BIGINT
	FROM crash_reports 
	ORDER BY created_at DESC 
	LIMIT $1
	`
	rows, err := s.db.Query(query, limit)
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
func (s *Store) GetProjects(userID string) ([]*Project, error) {
	query := `
	SELECT id, user_id, name, ingest_key, created_at
	FROM projects
	WHERE user_id = $1
	ORDER BY created_at DESC
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]*Project, 0)
	for rows.Next() {
		p := &Project{}
		err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.IngestKey, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (s *Store) ValidateKey(key string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM projects WHERE ingest_key = $1)"
	err := s.db.QueryRow(query, key).Scan(&exists)
	return exists, err
}
