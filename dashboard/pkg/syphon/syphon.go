package syphon

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	apex "github.com/apex/monitor/proto"
	"github.com/klauspost/compress/zstd"
	"google.golang.org/protobuf/proto"
)

type NetworkType int

const (
	NetworkUnknown NetworkType = iota
	NetworkNone
	NetworkCellular
	NetworkWifi
)

type NetworkStatus interface {
	CurrentType() NetworkType
}

type Syphon struct {
	network NetworkStatus
	client  *http.Client
}

func New(ns NetworkStatus) (*Syphon, error) {
	return &Syphon{
		network: ns,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (s *Syphon) SendBatch(reports []*apex.CrashReport, url, apiKey string) error {
	if url == "" {
		return fmt.Errorf("ingest URL is required")
	}

	compressed, err := s.PrepareBatch(reports)
	if err != nil {
		return err
	}

	maxRetries := 3
	backoff := 1 * time.Second

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(compressed))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/x-protobuf")
		req.Header.Set("Content-Encoding", "zstd")
		req.Header.Set("X-Apex-API-Key", apiKey)

		resp, err := s.client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusOK {
				return nil
			}
			lastErr = fmt.Errorf("server returned status: %s", resp.Status)
		} else {
			lastErr = err
		}

		// Exponential backoff
		time.Sleep(backoff)
		backoff *= 2
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

func (s *Syphon) PrepareBatch(reports []*apex.CrashReport) ([]byte, error) {
	batch := &apex.BatchReport{
		Reports: reports,
	}

	data, err := proto.Marshal(batch)
	if err != nil {
		return nil, err
	}

	var compressed bytes.Buffer
	zw, _ := zstd.NewWriter(&compressed)
	zw.Write(data)
	zw.Close()

	return compressed.Bytes(), nil
}

func (s *Syphon) ShouldSync() bool {
	if s.network == nil {
		return true
	}

	current := s.network.CurrentType()
	switch current {
	case NetworkWifi:
		return true
	case NetworkCellular:
		return false
	default:
		return false
	}
}
