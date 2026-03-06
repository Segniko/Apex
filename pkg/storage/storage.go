package storage

import apex "github.com/apex/monitor/proto"

// Provider is the interface for all storage implementations.
type Provider interface {
	SaveReport(r *apex.CrashReport) error
	GetReports(limit int) ([]*apex.CrashReport, error)
}
