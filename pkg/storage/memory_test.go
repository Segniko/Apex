package storage

import (
	"testing"
	"time"

	apex "github.com/Segniko/Apex/proto"
)

func TestGetIssuesGrouping(t *testing.T) {
	m := NewMemoryStore()
	m.SaveReport(&apex.CrashReport{ErrorId: "a", Message: "boom", Timestamp: 100}, "p1", "fp1")
	m.SaveReport(&apex.CrashReport{ErrorId: "b", Message: "boom", Timestamp: 200}, "p1", "fp1")
	m.SaveReport(&apex.CrashReport{ErrorId: "c", Message: "other", Timestamp: 150}, "p1", "fp2")
	m.SaveReport(&apex.CrashReport{ErrorId: "d", Message: "boom", Timestamp: 300}, "p2", "fp1")

	issues, err := m.GetIssues("p1", 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 2 {
		t.Fatalf("want 2 issues for p1, got %d", len(issues))
	}
	// Ordered by most recently seen; fp1 (last seen 200) should be first.
	top := issues[0]
	if top.Fingerprint != "fp1" || top.Count != 2 {
		t.Fatalf("expected fp1 x2, got %s x%d", top.Fingerprint, top.Count)
	}
	if top.LatestID != "b" || top.FirstSeen != 100 || top.LastSeen != 200 {
		t.Fatalf("bad aggregation: %+v", top)
	}
}

func TestProjectIsolationAndPrune(t *testing.T) {
	m := NewMemoryStore()
	m.SaveReport(&apex.CrashReport{ErrorId: "old", Timestamp: 1000}, "p", "f1")
	m.SaveReport(&apex.CrashReport{ErrorId: "new", Timestamp: time.Now().Unix()}, "p", "f2")

	reports, _ := m.GetReports(50, "p")
	if len(reports) != 2 {
		t.Fatalf("want 2 reports, got %d", len(reports))
	}

	removed, err := m.PruneReports(time.Unix(5000, 0))
	if err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Fatalf("want 1 pruned, got %d", removed)
	}
	if reports, _ = m.GetReports(50, "p"); len(reports) != 1 {
		t.Fatalf("want 1 remaining, got %d", len(reports))
	}
}
