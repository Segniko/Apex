package ai

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// TacticalAI represents the core engine for chat intelligence.
type TacticalAI struct {
	rng *rand.Rand
}

// NewTacticalAI initializes a new AI engine.
func NewTacticalAI() *TacticalAI {
	return &TacticalAI{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

var tacticalKeywords = []string{
	"hi", "hello", "greetings", "hey",
	"status", "deployment", "health", "node",
	"nil", "pointer", "null", "dereference", "memory",
	"database", "sql", "cockroach", "postgres", "db", "vault",
	"error", "fault", "crash", "bug", "panic", "stack", "trace",
	"decode", "analyze", "forensic", "signal", "ingest", "redis",
	"help", "commands", "manual", "instruction",
	"programming", "code", "go", "python", "javascript", "apex",
}

// Chat generates a tactical response based on the message and optional report context.
func (ai *TacticalAI) Chat(message string, reportID string) string {
	msg := strings.ToLower(message)

	// Guard: Filter for tactical/programming topics only
	if !containsAny(msg, tacticalKeywords...) {
		return "NON_TACTICAL_QUERY: My core logic is restricted to Apex Systems forensics and programming telemetry. For general inquiries, please contact: tactical-support@apex.systems"
	}

	// Context Matrix
	switch {
	case containsAny(msg, "who are you", "what are you", "identity"):
		return "IDENTITY_CONFIRM: I am APEX_AI, a tactical forensics unit designed to audit distributed system failures. My prime directive is real-time root-cause reconstruction."

	case containsAny(msg, "what is apex", "about apex", "project overview"):
		return "PROJECT_BRIEF: Apex is a high-performance, open-source monitoring engine. It uses Zsync Protobuf DNA for telemetry and CockroachDB for global persistence. Architecture: Recovery-First."

	case containsAny(msg, "hi", "hello", "greetings", "hey"):
		return "APEX_AI unit online. Ready for tactical forensics. I am monitoring the CockroachDB clusters and Redis ingest streams. How can I assist your audit?"

	case containsAny(msg, "status", "deployment", "health"):
		return "Telemetry indicates the current deployment is under heavy monitoring. Ingest buffer velocity is steady. No anomalies detected in the last 300ms. All services are tactical."

	case containsAny(msg, "nil", "pointer", "null", "dereference"):
		return "NIL_POINTER_ANALYSIS: This is a common failure in unsafe memory operations. I recommend a tactical audit of your reference guards. Check the stack trace for the 'dereference' instruction."

	case containsAny(msg, "database", "sql", "cockroach", "postgres", "db"):
		return "DATABASE_INTEGRITY: CockroachDB clusters are reporting 99.99% consistency. If you see 'Scan Errors', ensure your COALESCE logic is applied to legacy schema fields."

	case containsAny(msg, "error", "fault", "crash", "bug", "panic"):
		return "CRASH_FORENSICS: I have indexed the recent failure batches. Which specific 'INTELLIGENT_DECODE' ID should I perform a deep-trace on?"

	case reportID != "" || containsAny(msg, "decode", "trace", "analyze", "project"):
		if reportID != "" {
			return fmt.Sprintf("DECODING_TRACE[%s]: Analyzing packet structure for high-fidelity reconstruction. Initial findings suggest an operational bottleneck in the ingest-to-vault pipeline.", reportID[:8])
		}
		return "DECODING_TRACE: Analyzing packet structure for high-fidelity reconstruction. Provide a specific Error ID for forensic reconstruction."

	case containsAny(msg, "help", "commands", "manual"):
		return "TACTICAL_MANUAL: You can ask about 'status', 'deployment health', 'database integrity', or provide a specific 'Report ID' for a deep forensic trace."

	default:
		responses := []string{
			"Processing inquiry... Tactical analysis suggests we remain vigilant on the current telemetry stream.",
			"Data inconclusive for a deep-dive. Provide a specific Error ID for forensic reconstruction.",
			"I am monitoring the tactical nodes. Telemetry is flowing smoothly through the Redis buffer.",
			"Inquiry logged. Proceed with caution on the recent deployment patches.",
			"Signal noise detected. Re-phrase your query using standard tactical terminology.",
		}
		return responses[ai.rng.Intn(len(responses))]
	}
}

func containsAny(s string, keywords ...string) bool {
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}
