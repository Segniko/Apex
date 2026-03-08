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

// AnalysisRule defines a heuristic for identifying root causes.
type AnalysisRule struct {
	Pattern     string
	Description string
	TacticalFix string
}

var forensicRules = []AnalysisRule{
	{
		Pattern:     "nil pointer dereference",
		Description: "CRITICAL: Nil pointer access detected.",
		TacticalFix: "Check if the object is initialized before use. Add a nil-check guard.",
	},
	{
		Pattern:     "index out of range",
		Description: "DATA_BREACH: Array boundary violation.",
		TacticalFix: "Verify slice length before indexing. Use 'len()' guard.",
	},
	{
		Pattern:     "context deadline exceeded",
		Description: "INFRA_PULSE: Operation timed out.",
		TacticalFix: "Check downstream service health or increase context timeout values.",
	},
	{
		Pattern:     "invalid memory address",
		Description: "MEMORY_FAULT: Unsafe memory operation.",
		TacticalFix: "Audit pointer arithmetic and slice capacity allocations.",
	},
	{
		Pattern:     "no such file or directory",
		Description: "IO_FAILURE: Resource missing.",
		TacticalFix: "Verify path existence and permissions in the target environment.",
	},
	{
		Pattern:     "json: cannot unmarshal",
		Description: "DNA_CORRUPTION: Data structure mismatch.",
		TacticalFix: "Sync Protobuf/JSON schema definitions between agent and receiver.",
	},
}

// Analyze performs a tactical root-cause analysis of the crash.
func (r *Receiver) Analyze(report *apex.CrashReport) string {
	content := fmt.Sprintf("%s %s", report.Message, report.StackTrace)
	contentBytes := []byte(content)

	for _, rule := range forensicRules {
		if bytes.Contains(contentBytes, []byte(rule.Pattern)) {
			return fmt.Sprintf("%s // TACTICAL_FIX: %s", rule.Description, rule.TacticalFix)
		}
	}

	// Dynamic fallback for database issues
	if bytes.Contains(contentBytes, []byte("connection")) || bytes.Contains(contentBytes, []byte("db")) {
		return "INFRA_PULSE: Network/Database connection failure. Check connection strings or pool health."
	}

	return "FORENSIC_SIG: Pattern unrecognized. Manual stack-trace audit required. Architecture seems stable."
}
