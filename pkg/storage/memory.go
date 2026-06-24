package storage

import (
	"sort"
	"sync"
	"time"

	apex "github.com/Segniko/Apex/proto"
)

type MemoryStore struct {
	mu                sync.RWMutex
	reports           []*apex.CrashReport
	reportProject     map[string]string // reportID -> projectID
	reportFingerprint map[string]string // reportID -> fingerprint
	projects          []*Project
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		reports:           make([]*apex.CrashReport, 0),
		reportProject:     make(map[string]string),
		reportFingerprint: make(map[string]string),
		projects:          make([]*Project, 0),
	}
}

func (m *MemoryStore) SaveReport(r *apex.CrashReport, projectID string, fingerprint string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Keep last 100 reports in memory
	if len(m.reports) >= 100 {
		old := m.reports[0]
		m.reports = m.reports[1:]
		delete(m.reportProject, old.ErrorId)
		delete(m.reportFingerprint, old.ErrorId)
	}
	r.ProjectId = projectID
	m.reports = append(m.reports, r)
	m.reportProject[r.ErrorId] = projectID
	m.reportFingerprint[r.ErrorId] = fingerprint
	return nil
}

// GetIssues groups in-memory reports by fingerprint with occurrence counts.
func (m *MemoryStore) GetIssues(projectID string, limit int, offset int) ([]*Issue, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	byFp := make(map[string]*Issue)
	for _, r := range m.reports {
		if projectID != "" && m.reportProject[r.ErrorId] != projectID {
			continue
		}
		fp := m.reportFingerprint[r.ErrorId]
		if fp == "" {
			fp = r.ErrorId
		}
		iss, ok := byFp[fp]
		if !ok {
			byFp[fp] = &Issue{
				Fingerprint: fp, ProjectID: m.reportProject[r.ErrorId], Message: r.Message,
				StackTrace: r.StackTrace, AiInsight: r.AiInsight, Count: 1,
				FirstSeen: r.Timestamp, LastSeen: r.Timestamp, Resolved: r.Resolved, LatestID: r.ErrorId,
			}
			continue
		}
		iss.Count++
		iss.Resolved = iss.Resolved && r.Resolved
		if r.Timestamp < iss.FirstSeen {
			iss.FirstSeen = r.Timestamp
		}
		if r.Timestamp >= iss.LastSeen {
			iss.LastSeen = r.Timestamp
			iss.Message = r.Message
			iss.StackTrace = r.StackTrace
			iss.AiInsight = r.AiInsight
			iss.LatestID = r.ErrorId
		}
	}

	issues := make([]*Issue, 0, len(byFp))
	for _, iss := range byFp {
		issues = append(issues, iss)
	}
	sort.Slice(issues, func(i, j int) bool { return issues[i].LastSeen > issues[j].LastSeen })

	if offset >= len(issues) {
		return []*Issue{}, nil
	}
	end := offset + limit
	if end > len(issues) {
		end = len(issues)
	}
	return issues[offset:end], nil
}

// PruneReports drops in-memory reports older than the cutoff.
func (m *MemoryStore) PruneReports(olderThan time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := olderThan.Unix()
	kept := make([]*apex.CrashReport, 0, len(m.reports))
	var removed int64
	for _, r := range m.reports {
		if r.Timestamp < cutoff {
			delete(m.reportProject, r.ErrorId)
			delete(m.reportFingerprint, r.ErrorId)
			removed++
			continue
		}
		kept = append(kept, r)
	}
	m.reports = kept
	return removed, nil
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

func (m *MemoryStore) UpdateInsight(reportID, insight string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.reports {
		if r.ErrorId == reportID {
			r.AiInsight = insight
			return nil
		}
	}
	return nil
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
