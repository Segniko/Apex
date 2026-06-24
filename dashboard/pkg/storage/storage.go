package storage

import apex "github.com/Segniko/Apex/proto"

// Provider is the interface for all storage implementations.
type Provider interface {
	SaveReport(r *apex.CrashReport) error
	GetReports(limit int) ([]*apex.CrashReport, error)
}
