# Development Guide

This guide covers setting up a development environment, understanding the codebase, running tests, and contributing to Apex.

---

## Development Environment

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.25+ | Backend receiver and agents |
| Node.js | 20+ | Dashboard frontend |
| npm | 10+ | Package management |
| Docker | 20.10+ | Infrastructure services |
| Docker Compose | v2.0+ | Multi-container orchestration |
| protoc | 7.34+ | Protocol Buffer compilation (optional) |

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/Segniko/Apex.git
cd Apex

# Install Go dependencies
go mod download

# Install dashboard dependencies
cd dashboard && npm install && cd ..

# Start infrastructure
docker-compose up -d cockroach redis

# Run the receiver
export DATABASE_URL="postgresql://root@127.0.0.1:5433/defaultdb?sslmode=disable"
export REDIS_URL="127.0.0.1:6379"
go run cmd/server/main.go
```

In a separate terminal:

```bash
# Run the dashboard
cd dashboard
export NEXT_PUBLIC_API_URL="http://localhost:8081"
export AUTH_SECRET="dev-secret-at-least-32-characters-long"
npm run dev
```

---

## Project Layout

### Go Backend

```
main.go              → Agent demo (not the server)
cmd/
├── server/main.go   → Production HTTP server entry point
├── inspector/main.go → Vault database inspector
└── simulate/main.go  → Crash report simulator

pkg/
├── agent/           → Edge Agent (capture + sync)
├── ai/              → Gemini AI integration
├── limiter/         → Redis rate limiter
├── receiver/        → Batch unpacking + heuristic analysis
├── storage/         → Database abstraction layer
├── syphon/          → Network sync engine
└── vault/           → Encrypted local storage

proto/
├── apex.proto       → Protobuf schema definition
└── apex.pb.go       → Generated Go code (do not edit)
```

### Dashboard (Next.js)

```
dashboard/src/
├── auth.ts          → NextAuth.js configuration
├── middleware.ts    → Route protection middleware
├── app/
│   ├── layout.tsx   → Root HTML layout
│   ├── page.tsx     → Landing page
│   ├── globals.css  → Theme and animations
│   ├── auth/        → Authentication pages
│   ├── docs/        → Documentation page
│   └── dashboard/   → Protected dashboard pages
├── components/      → Reusable UI components
└── lib/
    └── api.ts       → API client functions
```

---

## Running the Components

### Receiver Server

```bash
# Minimal (in-memory mode, no external dependencies)
go run cmd/server/main.go

# With database
export DATABASE_URL="postgresql://root@127.0.0.1:5433/defaultdb?sslmode=disable"
go run cmd/server/main.go

# With all features
export DATABASE_URL="postgresql://root@127.0.0.1:5433/defaultdb?sslmode=disable"
export REDIS_URL="127.0.0.1:6379"
export GEMINI_API_KEY="your-gemini-key"
export APEX_API_KEY="dev-test-key"
go run cmd/server/main.go
```

### Dashboard

```bash
cd dashboard

# Development mode (hot reload)
npm run dev

# Production build
npm run build
npm start

# Lint check
npm run lint
```

### Utilities

```bash
# Simulate crash reports (sends to receiver)
go run cmd/simulate/main.go

# Inspect local vault database
go run cmd/inspector/main.go

# End-to-end verification
go run scripts/verify/tactical_verify.go
```

---

## Code Style and Conventions

### Go

- **Structured logging** with `log/slog` -- All log statements use structured fields
- **Error handling** -- Errors are returned, not panicked (except in the agent's CapturePanic which deliberately catches panics)
- **Interface-based design** -- The `storage.Provider` interface allows swappable backends
- **Package naming** -- Short, single-word package names (`agent`, `vault`, `syphon`)
- **No external frameworks** -- The HTTP server uses `net/http` directly

### TypeScript / React

- **Client components** -- Dashboard pages use `'use client'` for interactivity
- **Functional components** -- All components are functions, no class components
- **Tailwind CSS** -- Styling uses Tailwind utility classes exclusively
- **Color palette** -- Strict industrial theme:
  - Brand: `#FFB800` (amber)
  - Background: `#080808` (near-black)
  - Card: `#121212`
  - Success: `#00FF41` (green)
  - Danger: `#FF4D00` (orange-red)
  - Neutral: `#888888`
- **Font** -- `Inter` for body, monospace for code/data
- **No blue or pink** -- The design system explicitly avoids these colors

---

## Modifying the Protobuf Schema

If you need to change the crash report structure:

1. Edit `proto/apex.proto`
2. Regenerate the Go code:

```bash
protoc --go_out=. --go_opt=paths=source_relative proto/apex.proto
```

3. Update the JSON fallback structures in `pkg/receiver/receiver.go` to match
4. Update agent implementations to include new fields

**Note:** Changes to the Protobuf schema should be backward-compatible. Add new fields with new field numbers; never reuse or remove existing field numbers.

---

## Key Architectural Patterns

### Storage Provider Pattern

All database operations go through the `storage.Provider` interface:

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

To add a new storage backend:
1. Create a new file in `pkg/storage/` (e.g., `mongodb.go`)
2. Implement all methods of the `Provider` interface
3. Add initialization logic in `cmd/server/main.go`

### Redis Stream Pattern

The receiver uses Redis streams for async processing:

```go
// Producer (HTTP handler)
redisClient.XAdd(ctx, &redis.XAddArgs{
    Stream: "apex:ingest",
    Values: map[string]interface{}{
        "report": serializedReport,
    },
})

// Consumer (background worker)
results := redisClient.XRead(ctx, &redis.XReadArgs{
    Streams: []string{"apex:ingest", "$"},
    Block:   0,
})
```

### SSE Streaming Pattern

The AI chat endpoint streams responses:

```go
// Server side
w.Header().Set("Content-Type", "text/event-stream")
w.Header().Set("Cache-Control", "no-cache")
w.Header().Set("Connection", "keep-alive")

// Stream chunks
fmt.Fprintf(w, "data: %s\n\n", chunk)
flusher.Flush()

// End stream
fmt.Fprintf(w, "data: [DONE]\n\n")
```

```typescript
// Client side (TacticalChat.tsx)
const response = await fetch('/api/chat', {
    method: 'POST',
    body: JSON.stringify({ message, report_id }),
});

const reader = response.body.getReader();
const decoder = new TextDecoder();

while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    const text = decoder.decode(value);
    // Parse SSE data lines
}
```

---

## Dashboard Development

### Adding a New Page

1. Create a new directory under `dashboard/src/app/`:

```
dashboard/src/app/your-page/
└── page.tsx
```

2. For protected pages (requires login), ensure the route is covered by `middleware.ts`.

3. Follow the existing tactical theme:

```tsx
'use client';

export default function YourPage() {
    return (
        <div className="min-h-screen bg-[#080808] text-white">
            <div className="w-full h-1 hazard-pattern" />
            <main className="max-w-6xl mx-auto px-6 py-20">
                <h1 className="text-6xl font-black italic tracking-tighter uppercase">
                    Your <span className="text-[#FFB800]">Page</span>
                </h1>
            </main>
        </div>
    );
}
```

### Adding a New Component

1. Create the component in `dashboard/src/components/`
2. Follow the existing patterns:
   - Use `'use client'` for interactive components
   - Use Tailwind CSS classes only
   - Use the established color variables

### API Client

All receiver API calls go through `dashboard/src/lib/api.ts`:

```typescript
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081';

export async function fetchReports(projectId?: string): Promise<CrashReport[]> {
    const url = projectId
        ? `${API_URL}/api/reports?project_id=${projectId}`
        : `${API_URL}/api/reports`;
    const res = await fetch(url);
    if (!res.ok) return [];
    return res.json();
}
```

To add new API endpoints, add functions to this file following the same pattern.

---

## Adding a New Agent

To create an agent for a new language:

1. Create a directory under `agents/your-language/`
2. Implement the core workflow:
   - Generate a UUID for each crash
   - Capture error message and stack trace
   - Collect system information (OS, architecture, memory)
   - Build a report object matching the JSON schema
   - Wrap in `{"reports": [...]}` batch
   - Compress with Zstandard
   - POST to the receiver with `X-Apex-API-Key` header
3. Add documentation to `docs/AGENTS.md`
4. Update the README with integration examples

The JSON schema all agents must produce:

```json
{
    "reports": [
        {
            "error_id": "uuid-string",
            "message": "error message",
            "stack_trace": "full stack trace",
            "timestamp": 1773785000,
            "context": {
                "os": "operating-system",
                "arch": "cpu-architecture",
                "total_memory": 8589934592,
                "free_memory": 4294967296,
                "battery_level": 1.0,
                "is_charging": true,
                "network_type": "wifi"
            }
        }
    ]
}
```

---

## Common Development Commands

| Command | Description |
|---------|-------------|
| `go run cmd/server/main.go` | Start the receiver |
| `go run cmd/simulate/main.go` | Send test crash reports |
| `go run cmd/inspector/main.go` | Inspect local vault |
| `go mod tidy` | Clean up Go dependencies |
| `cd dashboard && npm run dev` | Start dashboard with hot reload |
| `cd dashboard && npm run build` | Build dashboard for production |
| `cd dashboard && npm run lint` | Lint TypeScript code |
| `docker-compose up -d cockroach redis` | Start infrastructure |
| `docker-compose logs -f receiver` | Tail receiver logs |
| `docker-compose down` | Stop all services |
| `docker-compose down -v` | Stop and wipe all data |
