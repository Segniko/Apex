package storage

import (
	"sync"

	apex "github.com/apex/monitor/proto"
)

type MemoryStore struct {
	mu       sync.RWMutex
	reports  []*apex.CrashReport
	projects []*Project
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		reports:  make([]*apex.CrashReport, 0),
		projects: make([]*Project, 0),
	}
}

func (m *MemoryStore) SaveReport(r *apex.CrashReport) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Keep last 100 reports in memory
	if len(m.reports) >= 100 {
		m.reports = m.reports[1:]
	}
	m.reports = append(m.reports, r)
	return nil
}

func (m *MemoryStore) GetReports(limit int) ([]*apex.CrashReport, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := len(m.reports)
	if count > limit {
		count = limit
	}

	// Return in reverse order (newest first)
	results := make([]*apex.CrashReport, 0, count)
	for i := len(m.reports) - 1; i >= len(m.reports)-count; i-- {
		results = append(results, m.reports[i])
	}

	return results, nil
}

func (m *MemoryStore) SaveProject(p *Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects = append(m.projects, p)
	return nil
}

func (m *MemoryStore) GetProjects(userID string) ([]*Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var userProjects []*Project
	for _, p := range m.projects {
		if p.UserID == userID || userID == "anonymous" {
			userProjects = append(userProjects, p)
		}
	}
	return userProjects, nil
}

func (m *MemoryStore) ValidateKey(key string) (bool, error) {
	// In memory mode, accept everything for now to avoid blocking ingest
	return true, nil
}
