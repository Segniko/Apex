package receiver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	apex "github.com/apex/monitor/proto"
	"github.com/klauspost/compress/zstd"
	"google.golang.org/protobuf/proto"
)

// Receiver identifies a server capable of processing Apex Syphon batches.
type Receiver struct {
	decoder *zstd.Decoder
}

// New creates a new Receiver.
func New() (*Receiver, error) {
	dec, err := zstd.NewReader(nil)
	if err != nil {
		return nil, err
	}
	return &Receiver{decoder: dec}, nil
}

// Unpack decompresses and deserializes an Apex batch.
func (r *Receiver) Unpack(compressedData []byte) (*apex.BatchReport, error) {
	// 1. Decompress
	var decompressed bytes.Buffer
	zr, err := zstd.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	if _, err := io.Copy(&decompressed, zr); err != nil {
		return nil, err
	}

	// 2. Try Protobuf first
	batch := &apex.BatchReport{}
	if err := proto.Unmarshal(decompressed.Bytes(), batch); err == nil {
		return batch, nil
	}

	// 3. Fallback to JSON for non-Go agents
	if err := json.Unmarshal(decompressed.Bytes(), batch); err != nil {
		return nil, fmt.Errorf("failed to decode batch (tried Proto and JSON): %w", err)
	}

	return batch, nil
}

// Analyze performs a tactical root-cause analysis of the crash.
func (r *Receiver) Analyze(report *apex.CrashReport) string {
	msg := report.Message
	trace := report.StackTrace

	// Pattern 1: Nil Pointer
	if bytes.Contains([]byte(trace), []byte("nil pointer dereference")) {
		return "CRITICAL: Nil pointer access detected. Check if the object is initialized before use. Potential fix: Add a nil-check guard."
	}

	// Pattern 2: Index Out of Range
	if bytes.Contains([]byte(trace), []byte("index out of range")) {
		return "DATA_BREACH: Array boundary violation. Verify slice length before indexing. Potential fix: Use 'len()' guard."
	}

	// Pattern 3: Database/Connection Issue
	if bytes.Contains([]byte(msg), []byte("connection")) || bytes.Contains([]byte(msg), []byte("db")) {
		return "INFRA_PULSE: Network/Database timeout. Check connection string or pool limits. Potential fix: Increase timeout or check DB health."
	}

	// Fallback
	return "FORENSIC_SIG: Pattern unrecognized. Manual stack-trace audit required. Architecture seems stable otherwise."
}
