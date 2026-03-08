package agent

import "time"

// Config defines the operational parameters for the Apex Agent.
type Config struct {
	// IngestURL is the endpoint where Syphon sends batches.
	IngestURL string

	// APIKey is the authentication token for the receiver.
	APIKey string

	// SyncInterval determines how often the Agent checks the Vault for new crashes.
	SyncInterval time.Duration

	// BatchSize is the maximum number of crashes to send in one batch.
	BatchSize int
}

// DefaultConfig returns a sane set of defaults for production.
func DefaultConfig() Config {
	return Config{
		SyncInterval: 1 * time.Minute,
		BatchSize:    50,
	}
}
