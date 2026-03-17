# Architecture

This document provides a deep dive into the internal architecture of Apex, covering system design, data flow, component responsibilities, and key design decisions.

---

## System Overview

Apex follows a distributed, event-driven architecture with three primary tiers:

1. **Edge Tier** -- Lightweight agents embedded in user applications that capture crashes and transmit them.
2. **Ingest Tier** -- A Go-based HTTP server that validates, decompresses, and routes crash data through a Redis stream to background workers.
3. **Presentation Tier** -- A Next.js dashboard that provides real-time visibility into crash data with AI-assisted analysis.

```
                         EDGE TIER
    ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
    │  Go Agent    │  │ Python Agent │  │ Node.js Agent│
    │  ┌────────┐  │  │  ┌────────┐  │  │  ┌────────┐  │
    │  │ Vault  │  │  │  │Capture │  │  │  │Capture │  │
    │  │(SQLite)│  │  │  └───┬────┘  │  │  └───┬────┘  │
    │  └───┬────┘  │  │      │       │  │      │       │
    │  ┌───┴────┐  │  │  ┌───┴────┐  │  │  ┌───┴────┐  │
    │  │ Syphon │  │  │  │Compress│  │  │  │Compress│  │
    │  └───┬────┘  │  │  └───┬────┘  │  │  └───┬────┘  │
    └──────┼───────┘  └──────┼───────┘  └──────┼───────┘
           │                 │                 │
           │  Zstd + Proto   │  Zstd + JSON    │  Zstd + JSON
           └────────────┬────┴────────────┬────┘
                        │                 │
                        v                 v
                  ┌─────────────────────────────┐
                  │      INGEST TIER            │
                  │  ┌───────────────────────┐  │
                  │  │   HTTP Server (:8081) │  │
                  │  │   - /ingest           │  │
                  │  │   - /api/reports      │  │
                  │  │   - /api/projects     │  │
                  │  │   - /api/chat         │  │
                  │  │   - /metrics          │  │
                  │  └──────────┬────────────┘  │
                  │             │               │
                  │      ┌──────┴──────┐        │
                  │      │ Redis Stream│        │
                  │      │ (apex:ingest)│       │
                  │      └──────┬──────┘        │
                  │             │               │
                  │      ┌──────┴──────┐        │
                  │      │   Worker    │        │
                  │      │  - AI Call  │        │
                  │      │  - DB Write │        │
                  │      └──────┬──────┘        │
                  │             │               │
                  │   ┌─────────┴─────────┐     │
                  │   │                   │     │
                  │   v                   v     │
                  │ ┌──────────┐  ┌────────┐   │
                  │ │CockroachDB│ │Gemini  │   │
                  │ │/Postgres │  │  AI    │   │
                  │ └──────────┘  └────────┘   │
                  └─────────────────────────────┘
                        │
                        v
                  ┌─────────────────────────────┐
                  │    PRESENTATION TIER         │
                  │  ┌───────────────────────┐   │
                  │  │ Next.js Dashboard     │   │
                  │  │ - GitHub OAuth        │   │
                  │  │ - Project Management  │   │
                  │  │ - Crash HUD           │   │
                  │  │ - AI Chat (SSE)       │   │
                  │  └───────────────────────┘   │
                  └─────────────────────────────┘
```

---

## Component Details

### 1. Edge Agents

#### Go Agent (`pkg/agent`)

The Go agent is the most feature-complete, with:

- **Panic recovery** via `defer agent.CapturePanic()` which catches panics, converts them to crash reports with full stack traces, and stores them in the Vault.
- **Background sync loop** that periodically checks the Vault for unsent reports and dispatches them via the Syphon.
- **Graceful shutdown** via `agent.Stop()` that flushes remaining reports before exiting.

The agent collects device context at capture time:
- OS and architecture (`runtime.GOOS`, `runtime.GOARCH`)
- Memory statistics (`runtime.MemStats`)
- Battery level and charging status (placeholder for mobile targets)
- Network type classification

#### Python Agent (`agents/python`)

A streamlined agent that:
- Accepts exception objects and formatted tracebacks
- Collects OS, architecture, and memory info via `platform` and `psutil`
- Serializes to JSON, compresses with `zstandard`, and POSTs to the receiver
- Single-shot capture (no background sync loop or local storage)

#### Node.js Agent (`agents/node`)

Similar to the Python agent:
- Captures errors with their stack traces
- Collects system info via Node.js `os` module
- Compresses with `fzstd` (Zstandard for JavaScript)
- Sends via `fetch` API

### 2. Vault (`pkg/vault`)

The Vault provides crash-data-at-rest encryption using AES-256-GCM:

```
┌─────────────────────────────────────┐
│            Vault                    │
│  ┌───────────┐   ┌──────────────┐  │
│  │ SQLite DB │   │ AES-256-GCM  │  │
│  │           │   │              │  │
│  │ crash_dna │◄──│ encrypt()    │  │
│  │  (table)  │──►│ decrypt()    │  │
│  └───────────┘   └──────────────┘  │
│                                     │
│  Operations:                        │
│  - Save(report) → encrypt → INSERT  │
│  - FetchAll() → SELECT → decrypt    │
│  - Cleanup(ts) → DELETE WHERE < ts  │
└─────────────────────────────────────┘
```

**Design decisions:**
- Uses `modernc.org/sqlite` (pure Go, no CGO) for maximum portability.
- Random nonce per encryption operation prevents nonce reuse attacks.
- Reports are serialized as JSON before encryption.
- Cleanup is timestamp-based, allowing the agent to prune after successful sync.

### 3. Syphon (`pkg/syphon`)

The Syphon manages the network boundary between agents and the receiver:

```
Reports → Marshal(Protobuf) → Compress(Zstd) → HTTP POST
                                                  │
                                          ┌───────┴───────┐
                                          │  Retry Logic  │
                                          │  Attempt 1: 0s│
                                          │  Attempt 2: 1s│
                                          │  Attempt 3: 2s│
                                          │  Attempt 4: 4s│
                                          └───────────────┘
```

**Network type awareness:**
- `NetworkWifi` -- Always sync
- `NetworkCellular` -- Defer sync (configurable)
- `NetworkNone` -- Never sync
- `NetworkUnknown` -- Always sync (default)

### 4. Receiver (`cmd/server`, `pkg/receiver`)

The receiver is the central HTTP server. Its request lifecycle for `/ingest`:

```
HTTP Request
    │
    ▼
┌──────────────────────────────────────────────┐
│ 1. API Key Validation                        │
│    - Check X-Apex-API-Key header             │
│    - Validate against project keys in DB     │
│    - Fall back to APEX_API_KEY env var        │
│    - Return 401 if invalid                   │
└──────────────────┬───────────────────────────┘
                   │
                   ▼
┌──────────────────────────────────────────────┐
│ 2. Body Read + Decompression                 │
│    - Read full request body                  │
│    - Decompress with Zstandard               │
└──────────────────┬───────────────────────────┘
                   │
                   ▼
┌──────────────────────────────────────────────┐
│ 3. Deserialization                           │
│    - Try Protocol Buffers unmarshal first    │
│    - Fall back to JSON unmarshal             │
│    - Extract individual reports from batch   │
└──────────────────┬───────────────────────────┘
                   │
                   ▼
┌──────────────────────────────────────────────┐
│ 4. Report Routing                            │
│    IF Redis available:                       │
│      → Push to Redis stream (apex:ingest)    │
│      → Background worker processes async     │
│    ELSE:                                     │
│      → Direct DB save (synchronous)          │
│      → Heuristic analysis (no AI)            │
└──────────────────┬───────────────────────────┘
                   │
                   ▼
             202 Accepted
```

**Background Worker (Redis consumer):**

```
Redis Stream (apex:ingest)
    │
    ▼
┌──────────────────────────────────────┐
│ 1. Read report from stream           │
│ 2. Generate cache fingerprint        │
│    (SHA-256 of message + stack)      │
│ 3. Check Redis cache for insight     │
│ 4. If not cached:                    │
│    a. Check rate limit               │
│    b. Fetch similar reports (RAG)    │
│    c. Call Gemini AI analysis        │
│    d. Cache result (24h TTL)         │
│ 5. Attach AI insight to report       │
│ 6. Save to database                  │
│ 7. Update Prometheus metrics         │
└──────────────────────────────────────┘
```

### 5. Storage Layer (`pkg/storage`)

The storage layer uses an interface-based design for swappable backends:

```go
type Provider interface {
    SaveReport(report)
    GetReports(projectID) []CrashReport
    SaveProject(project)
    GetProjects(userID) []Project
    ValidateKey(apiKey) bool
    GetSimilarReports(message, projectID) []CrashReport
    Close()
}
```

**PostgreSQL/CockroachDB implementation** (`postgres.go`, `query.go`):
- Connection pooling: 25 max open/idle connections, 5-minute lifetime
- Auto-creates `crash_reports` and `projects` tables on initialization
- Indices on `created_at`, `user_id`, and `project_id` for query performance
- Similar report retrieval uses `ILIKE` text matching

**Memory implementation** (`memory.go`):
- Ring buffer of 100 reports (oldest evicted)
- Thread-safe via `sync.RWMutex`
- Used as fallback when no database is available

### 6. AI Module (`pkg/ai`)

The `TacticalAI` struct wraps the Google Gemini client:

```
┌─────────────────────────────────────────────────┐
│                 TacticalAI                       │
│                                                  │
│  AnalyzeReport(report)                           │
│  ├── Build prompt with report data               │
│  ├── Include similar reports (RAG context)       │
│  ├── Call Gemini gemini-2.0-flash                │
│  └── Return structured forensic analysis         │
│                                                  │
│  Chat(message, report)                           │
│  ├── Build context from crash report             │
│  ├── Call Gemini with conversation context        │
│  └── Return response text                        │
│                                                  │
│  ChatStream(message, report, writer)             │
│  ├── Build context from crash report             │
│  ├── Stream Gemini response via SSE              │
│  └── Write chunks as data: events                │
│                                                  │
│  Dependencies:                                   │
│  - Redis (caching, rate limiting)                │
│  - Storage Provider (similar reports for RAG)    │
│  - Gemini API client                             │
└─────────────────────────────────────────────────┘
```

### 7. Rate Limiter (`pkg/limiter`)

Uses Redis sorted sets for a sliding window algorithm:

```
Key: "ratelimit:{identifier}"
Score: Unix timestamp (nanoseconds)
Member: Unique request ID

Allow(key, limit, window):
  1. Remove entries older than (now - window)
  2. Count remaining entries
  3. If count < limit: add new entry, return true
  4. Else: return false
```

**Fail-closed design:** If Redis is unavailable, all requests are denied. This protects the Gemini API quota from being exceeded during Redis outages.

### 8. Dashboard (`dashboard/`)

The Next.js dashboard follows the App Router pattern:

```
dashboard/src/
├── auth.ts              # NextAuth config (server-side)
├── middleware.ts         # Route protection
├── app/
│   ├── layout.tsx       # Root layout with Providers
│   ├── page.tsx         # Public landing page
│   ├── auth/login/      # Login (server component)
│   ├── docs/            # Documentation page
│   └── dashboard/       # Protected routes
│       ├── page.tsx     # Global crash feed
│       └── projects/
│           ├── page.tsx        # Project CRUD
│           └── [id]/page.tsx   # Per-project HUD
├── components/
│   ├── CrashCard.tsx    # Report display + actions
│   ├── TacticalChat.tsx # SSE AI chat widget
│   ├── UserButton.tsx   # Auth state display
│   ├── Providers.tsx    # SessionProvider wrapper
│   └── MobileBlocker.tsx# Mobile gate
└── lib/
    └── api.ts           # HTTP client for receiver API
```

**Key patterns:**
- All dashboard pages are client components (`'use client'`) for real-time interactivity
- Reports are polled every 5 seconds via `setInterval`
- The login page is a server component that checks session and redirects
- TacticalChat uses SSE (`EventSource` pattern via `fetch` + `ReadableStream`) for streaming AI responses
- Authentication state is managed via NextAuth's `useSession` hook

---

## Data Models

### CrashReport (Protobuf)

```protobuf
message DeviceContext {
    string os = 1;
    string arch = 2;
    uint64 total_memory = 3;
    uint64 free_memory = 4;
    float battery_level = 5;
    bool is_charging = 6;
    string network_type = 7;
}

message CrashReport {
    string error_id = 1;
    string message = 2;
    string stack_trace = 3;
    int64 timestamp = 4;
    DeviceContext context = 5;
    string ai_insight = 6;
}

message BatchReport {
    repeated CrashReport reports = 1;
}
```

### Project (SQL)

```sql
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    ingest_key TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_projects_user ON projects(user_id);
```

### CrashReport (SQL)

```sql
CREATE TABLE IF NOT EXISTS crash_reports (
    error_id TEXT PRIMARY KEY,
    message TEXT,
    stack_trace TEXT,
    timestamp BIGINT,
    os TEXT,
    arch TEXT,
    total_memory BIGINT,
    free_memory BIGINT,
    battery_level REAL,
    is_charging BOOLEAN,
    network_type TEXT,
    ai_insight TEXT,
    user_id TEXT DEFAULT '',
    project_id TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reports_created ON crash_reports(created_at);
CREATE INDEX IF NOT EXISTS idx_reports_user ON crash_reports(user_id);
CREATE INDEX IF NOT EXISTS idx_reports_project ON crash_reports(project_id);
```

---

## Infrastructure

### Docker Compose Services

```yaml
Services:
  cockroach    → CockroachDB v23.1.10 (port 5433/8082)
  redis        → Redis 7 Alpine (port 6379)
  receiver     → Go receiver (port 8081, depends on cockroach + redis)
  dashboard    → Next.js app (port 3000)
  prometheus   → Prometheus (port 9090, scrapes receiver:8081)
  grafana      → Grafana (port 3001, reads from prometheus:9090)
```

### Metrics Pipeline

```
Receiver ──(/metrics)──► Prometheus ──► Grafana
                           │
                           │ scrape_interval: 15s
                           │ job_name: apex-receiver
                           │ target: receiver:8081
```

**Exported metrics:**
- `apex_reports_received_total` (counter) -- Cumulative crash reports ingested
- `apex_ingest_errors_total` (counter) -- Cumulative ingestion failures
- `apex_ingest_duration_seconds` (histogram) -- Processing time per batch

---

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Redis as ingest buffer** | Decouples HTTP acceptance from database writes. Allows the receiver to return 202 immediately while AI analysis runs asynchronously. |
| **Protobuf + JSON fallback** | Protobuf for Go-to-Go efficiency; JSON fallback ensures Python/Node.js agents work without protoc tooling. |
| **Zstandard compression** | Best compression ratio at high speed. The `klauspost/compress` implementation is one of the fastest available. |
| **AES-256-GCM for Vault** | Authenticated encryption prevents both tampering and data exposure. GCM mode is hardware-accelerated on modern CPUs. |
| **Sliding window rate limiter** | More accurate than fixed-window counters; prevents burst abuse at window boundaries. |
| **In-memory fallback storage** | Ensures the receiver always starts, even without infrastructure. Useful for development and demos. |
| **SSE for AI chat** | Server-Sent Events provide simple, unidirectional streaming without the complexity of WebSockets. Ideal for progressive text rendering. |
| **NextAuth.js v5** | Battle-tested authentication with JWT sessions. GitHub provider leverages existing developer identities. |
| **CockroachDB over PostgreSQL** | While using the PostgreSQL wire protocol, CockroachDB provides automatic sharding and geo-distribution for production scale. The code is compatible with standard PostgreSQL. |
| **Pure Go SQLite** | `modernc.org/sqlite` avoids CGO dependency, making the agent binary fully self-contained and cross-compilable. |
