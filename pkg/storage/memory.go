package storage

import (
	"sync"

	apex "github.com/apex/monitor/proto"
)

type MemoryStore struct {
	mu            sync.RWMutex
	reports       []*apex.CrashReport
	reportProject map[string]string // reportID -> projectID
	projects      []*Project
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		reports:       make([]*apex.CrashReport, 0),
		reportProject: make(map[string]string),
		projects:      make([]*Project, 0),
	}
}

func (m *MemoryStore) SaveReport(r *apex.CrashReport, projectID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Keep last 100 reports in memory
	if len(m.reports) >= 100 {
		old := m.reports[0]
		m.reports = m.reports[1:]
		delete(m.reportProject, old.ErrorId)
	}
	m.reports = append(m.reports, r)
	m.reportProject[r.ErrorId] = projectID
	return nil
}

func (m *MemoryStore) GetReports(limit int, projectID string) ([]*apex.CrashReport, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Filter in memory
	var filtered []*apex.CrashReport
	for i := len(m.reports) - 1; i >= 0; i-- {
		r := m.reports[i]
		if projectID == "" || m.reportProject[r.ErrorId] == projectID {
			filtered = append(filtered, r)
		}
		if len(filtered) >= limit {
			break
		}
	}

	return filtered, nil
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

func (m *MemoryStore) ValidateKey(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.projects {
		if p.IngestKey == key {
			return p.ID, nil
		}
	}
	// Fallback for demo
	return "00000000-0000-0000-0000-000000000000", nil
}
func (m *MemoryStore) GetSimilarReports(message string, limit int, projectID string) ([]*apex.CrashReport, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var similar []*apex.CrashReport
	// Very simple prefix/containment match for MemoryStore
	for i := len(m.reports) - 1; i >= 0; i-- {
		r := m.reports[i]
		if (projectID == "" || r.ProjectId == projectID) && r.Message == message {
			similar = append(similar, r)
		}
		if len(similar) >= limit {
			break
		}
	}
	return similar, nil
}

func (m *MemoryStore) ResolveReport(id string, resolved bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.reports {
		if r.ErrorId == id {
			r.Resolved = resolved
			return nil
		}
	}
	return nil
}

func (m *MemoryStore) DeleteProject(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. Delete the project
	for i, p := range m.projects {
		if p.ID == id {
			m.projects = append(m.projects[:i], m.projects[i+1:]...)
			break
		}
	}

	// 2. Delete associated reports
	var remaining []*apex.CrashReport
	for _, r := range m.reports {
		if r.ProjectId != id {
			remaining = append(remaining, r)
		}
	}
	m.reports = remaining

	return nil
}
