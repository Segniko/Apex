package storage

import (
	apex "github.com/Segniko/Apex/proto"
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

	var reports []*apex.CrashReport
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
