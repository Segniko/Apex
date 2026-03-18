# Apex: The Architecture of Recovery

**Industrial-grade failure forensics. 100% Free. 100% Open Source.**

Apex is a high-performance crash monitoring and observability engine built for modern infrastructure. It captures, compresses, and syncs crash report "DNA" from your applications in real time, then uses AI-powered forensics to decode root causes and suggest fixes. Designed with a "Community First" philosophy, it brings enterprise-grade capabilities to every developer's terminal.

**Live Demo:** [https://apex-addis.vercel.app](https://apex-addis.vercel.app)

---

## Documentation Process

- I started by writing the documentation manually myself outlining the architecture, key features.
- Once I had a solid foundation I used Google's Code Wiki to generate additional structure, diagrams, and comprehensive overviews. I then reviewed everything refined the AI suggestions and merged the best parts into the final docs. Documentation process doesn't stop here. I will keep on updating it as I add more features and improve the existing ones.

## Recommendation for others

If you're documenting a project:

- Write the core documentation yourself first, feed your manual version + codebase into Code Wiki and then refine the output. Fix any erros and keep only what you actually want or what is needed.

## Table of Contents

- [Features](#features)
- [Architecture Overview](#architecture-overview)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Cloud-Hosted Quickstart](#cloud-hosted-quickstart)
  - [Self-Hosting with Docker](#self-hosting-with-docker)
  - [Running Locally (Development)](#running-locally-development)
- [Project Structure](#project-structure)
- [Core Concepts](#core-concepts)
  - [Crash Report DNA](#crash-report-dna)
  - [The Vault (Local Encrypted Storage)](#the-vault-local-encrypted-storage)
  - [The Syphon (Network Sync Engine)](#the-syphon-network-sync-engine)
  - [The Receiver (Ingest Server)](#the-receiver-ingest-server)
  - [AI Forensics (Gemini Integration)](#ai-forensics-gemini-integration)
  - [Rate Limiting](#rate-limiting)
- [Agent Integration](#agent-integration)
  - [Go Agent](#go-agent)
  - [Python Agent](#python-agent)
  - [Node.js Agent](#nodejs-agent)
- [API Reference](#api-reference)
- [Dashboard](#dashboard)
- [Observability Stack](#observability-stack)
- [Deployment](#deployment)
  - [Docker Compose (Self-Hosted)](#docker-compose-self-hosted)
  - [Render](#render)
  - [Vercel (Dashboard Only)](#vercel-dashboard-only)
- [Configuration](#configuration)
- [Contributing](#contributing)
- [Tactical Roadmap](#tactical-roadmap-the-future-of-apex)
- [Building Apex: Lessons & Technical Deep Dive](#building-apex-lessons--technical-deep-dive)
  - [Project Initialization & Package Layout](#project-initialization--package-layout)
  - [Why I Chose Protobuf](#why-i-chose-protobuf)
  - [Compiler Chain & Dependencies](#compiler-chain--dependencies)
  - [Key Go Concepts I Mastered](#key-go-concepts-i-mastered)
  - [Advanced Features & Encounters](#advanced-features--encounters)
  - [Operational Excellence](#operational-excellence)
- [License](#license)

---

## Features

- **Redis-Buffered Ingest Pipeline** -- Crash reports are offloaded to a Redis stream, decoupling reception from storage for high throughput and zero data loss.
- **AI-Powered Root-Cause Analysis** -- Every crash report is automatically analyzed by Google Gemini, providing structured forensic breakdowns including root cause, impact assessment, and tactical fixes.
- **Real-Time AI Chat** -- An interactive chat interface (SSE-based streaming) lets developers ask follow-up questions about specific crash reports with full source-code context.
- **Multi-Language Edge Agents** -- Lightweight agents for Go, Python, and Node.js that capture exceptions, collect device/system telemetry, compress payloads with Zstandard, and sync them to the receiver.
- **Encrypted Local Vault** -- An AES-256-GCM encrypted SQLite database on the agent side ensures crash data is stored securely before sync.
- **Protocol Buffers Serialization** -- Crash reports use Protobuf for compact, schema-enforced serialization with JSON fallback for non-Go agents.
- **CockroachDB Persistence** -- Production storage on CockroachDB for globally distributed, strongly consistent data with automatic failover.
- **Prometheus + Grafana Monitoring** -- Built-in metrics exposure (`/metrics`) with pre-configured Prometheus scraping and Grafana dashboards.
- **GitHub OAuth Authentication** -- The dashboard uses NextAuth.js with GitHub provider for secure user authentication and project isolation.
- **Project-Based Data Isolation** -- Each project gets a unique ingest key, ensuring crash data is isolated per workspace.
- **RAG-Enhanced AI Context** -- AI analysis pulls historical similar reports for retrieval-augmented context, improving forensic accuracy over time.
- **Resilient Sync with Exponential Backoff** -- The Syphon module retries failed transmissions with exponential backoff and network-aware sync decisions.

---

## Architecture Overview

```
+-------------------+     +-------------------+     +-------------------+
|   Go/Python/Node  |     |   Go/Python/Node  |     |   Go/Python/Node  |
|   Edge Agent      |     |   Edge Agent      |     |   Edge Agent      |
+--------+----------+     +--------+----------+     +--------+----------+
         |                         |                         |
         |  Zstd + Protobuf/JSON   |                         |
         +-------------------------+-------------------------+
                                   |
                                   v
                      +------------+-------------+
                      |   Apex Receiver (Go)     |
                      |   HTTP Ingest Server     |
                      |   Port 8081              |
                      +---+------------------+---+
                          |                  |
                          v                  v
                   +------+------+    +------+------+
                   | Redis Stream|    | Prometheus  |
                   | (Buffer)    |    | /metrics    |
                   +------+------+    +------+------+
                          |                  |
                          v                  v
                   +------+------+    +------+------+
                   | Worker Pool |    | Grafana     |
                   | (DB Writer) |    | Dashboards  |
                   +------+------+    +-------------+
                          |
                     +----+-----+
                     |          |
                     v          v
              +------+---+ +---+----------+
              |CockroachDB| |Gemini AI     |
              |(Postgres) | |(Forensics +  |
              |           | | RAG Context) |
              +------+----+ +---+----------+
                     |          |
                     v          v
              +------+----------+------+
              |  Next.js Dashboard     |
              |  (Port 3000)           |
              |  - Project Management  |
              |  - Crash HUD           |
              |  - AI Chat (SSE)       |
              |  - GitHub OAuth        |
              +------------------------+
```

**Data Flow:**

1. An Edge Agent captures a crash (panic, exception, error) along with device context (OS, architecture, memory, battery).
2. The report is serialized via Protobuf (Go) or JSON (Python/Node.js), compressed with Zstandard, and transmitted to the Receiver.
3. The Receiver validates the API key, decompresses and deserializes the batch, then pushes each report to a Redis stream.
4. A background worker reads from the Redis stream, runs Gemini AI analysis (with caching and rate limiting), and persists the enriched report to CockroachDB.
5. The Next.js dashboard polls the Receiver's REST API to display crash reports with AI insights and provides an interactive chat for deeper analysis.

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| **Backend / Receiver** | Go 1.25, `net/http`, structured logging (`log/slog`) |
| **Serialization** | Protocol Buffers (protoc v7.34), `google.golang.org/protobuf` |
| **Compression** | Zstandard (`github.com/klauspost/compress/zstd`) |
| **Database** | CockroachDB (PostgreSQL-compatible), `github.com/lib/pq` |
| **Cache / Message Queue** | Redis 7 (`github.com/redis/go-redis/v9`) |
| **AI Engine** | Google Gemini (`github.com/google/generative-ai-go`) |
| **Local Vault** | SQLite via `modernc.org/sqlite`, AES-256-GCM encryption |
| **Frontend / Dashboard** | Next.js 16, React 19, Tailwind CSS v4, TypeScript |
| **Authentication** | NextAuth.js v5 (GitHub OAuth) |
| **Metrics** | Prometheus client (`github.com/prometheus/client_golang`) |
| **Visualization** | Grafana with pre-provisioned dashboards |
| **Containerization** | Docker, Docker Compose |
| **Deployment** | Render, Vercel (dashboard) |

---

## Getting Started

### Prerequisites

- **Go 1.25+** (for backend development)
- **Node.js 20+** and **npm** (for dashboard development)
- **Docker** and **Docker Compose** (for self-hosting)
- A **Gemini API Key** from [Google AI Studio](https://aistudio.google.com/) (optional, for AI features)
- A **GitHub OAuth App** (optional, for dashboard authentication)

### Cloud-Hosted Quickstart

The fastest way to get started is to use the centralized Command Center:

1. Visit [https://apex.vercel.app](https://apex.vercel.app)
2. Sign in with GitHub
3. Create a project and copy your unique **Ingest Key**
4. Drop an Edge Agent into your codebase (see [Agent Integration](#agent-integration))

### Self-Hosting with Docker

For full data sovereignty, spin up the entire Apex stack locally:

```bash
# Clone the repository
git clone https://github.com/Segniko/Apex.git
cd Apex

# (Optional) Configure environment variables
cp .env.example .env  # Then edit .env with your keys

# Launch all services
docker-compose up -d
```

This starts:
- **CockroachDB** on port `5433` (DB) and `8082` (admin UI)
- **Redis** on port `6379`
- **Apex Receiver** on port `8081`
- **Next.js Dashboard** on port `3000`
- **Prometheus** on port `9090`
- **Grafana** on port `3001` (default login: `admin` / `admin`)

Access your local Command Center at **http://localhost:3000**.

### Running Locally (Development)

To run individual components for development:

**1. Start infrastructure:**
```bash
docker-compose up -d cockroach redis prometheus grafana
```

**2. Run the Go Receiver:**
```bash
# Set environment variables
export DATABASE_URL="postgresql://root@127.0.0.1:5433/defaultdb?sslmode=disable"
export REDIS_URL="127.0.0.1:6379"
export GEMINI_API_KEY="your-gemini-api-key"  # Optional

# Start the receiver
go run cmd/server/main.go
```

**3. Run the Dashboard:**
```bash
cd dashboard
npm install
npm run dev
```

The dashboard will be available at **http://localhost:3000** and connects to the receiver at **http://localhost:8081** by default (configurable via `NEXT_PUBLIC_API_URL`).

**Windows users** can use the provided `run.bat` script to launch all components at once.

---

## Project Structure

```
Apex/
├── main.go                    # Agent demo entrypoint
├── go.mod / go.sum            # Go module dependencies
├── Dockerfile                 # Receiver container build
├── docker-compose.yml         # Full stack orchestration
├── render.yaml                # Render.com deployment config
├── run.bat                    # Windows launch script
│
├── cmd/                       # Command-line entrypoints
│   ├── server/main.go         # Production receiver server (HTTP ingest + API)
│   ├── inspector/main.go      # Vault inspection utility
│   └── simulate/main.go       # Crash report simulation tool
│
├── pkg/                       # Core Go library packages
│   ├── agent/                 # Edge Agent (crash capture, sync loop)
│   │   ├── agent.go           # Agent struct, CapturePanic, sync logic
│   │   └── config.go          # Agent configuration (intervals, batch size)
│   ├── ai/                    # Gemini AI integration
│   │   └── ai.go              # TacticalAI: chat, stream, analyze
│   ├── limiter/               # Redis-based rate limiting
│   │   └── limiter.go         # Token bucket rate limiter
│   ├── receiver/              # Batch unpacking and heuristic analysis
│   │   └── receiver.go        # Zstd decompression, Proto/JSON decode, forensic rules
│   ├── storage/               # Data persistence layer
│   │   ├── storage.go         # Provider interface and Project model
│   │   ├── postgres.go        # CockroachDB/PostgreSQL implementation
│   │   ├── query.go           # Report and project query methods
│   │   └── memory.go          # In-memory fallback store
│   ├── syphon/                # Network sync engine
│   │   └── syphon.go          # Batch compression, retry with backoff
│   └── vault/                 # Encrypted local storage
│       └── vault.go           # AES-256-GCM encrypted SQLite vault
│
├── proto/                     # Protocol Buffer definitions
│   ├── apex.proto             # Schema: DeviceContext, CrashReport, BatchReport
│   └── apex.pb.go             # Generated Go code
│
├── agents/                    # Non-Go edge agents
│   ├── python/
│   │   └── agent.py           # Python agent (requests + zstandard)
│   └── node/
│       ├── agent.js           # Node.js agent (fzstd + uuid)
│       └── package.json       # Node agent dependencies
│
├── dashboard/                 # Next.js frontend application
│   ├── package.json           # Frontend dependencies
│   ├── Dockerfile             # Dashboard container build
│   ├── vercel.json            # Vercel deployment config
│   ├── next.config.ts         # Next.js configuration
│   ├── tsconfig.json          # TypeScript configuration
│   └── src/
│       ├── auth.ts            # NextAuth.js config (GitHub OAuth)
│       ├── middleware.ts      # Auth middleware (route protection)
│       ├── lib/
│       │   └── api.ts         # API client (fetch reports, projects, status)
│       ├── components/
│       │   ├── CrashCard.tsx   # Crash report display card
│       │   ├── TacticalChat.tsx# AI chat widget (SSE streaming)
│       │   ├── UserButton.tsx  # Auth user display/logout
│       │   ├── Providers.tsx   # NextAuth session provider
│       │   └── MobileBlocker.tsx # Mobile device gate
│       └── app/
│           ├── layout.tsx     # Root layout
│           ├── page.tsx       # Landing page
│           ├── globals.css    # Global styles and animations
│           ├── auth/login/    # Login page (GitHub OAuth)
│           ├── docs/          # Documentation/changelog page
│           ├── dashboard/
│           │   ├── page.tsx   # Global crash feed
│           │   └── projects/
│           │       ├── page.tsx       # Project hub (create/list)
│           │       └── [id]/page.tsx  # Per-project crash HUD
│           └── api/auth/      # NextAuth API routes
│
├── deploy/                    # Deployment configurations
│   ├── prometheus/
│   │   └── prometheus.yml     # Prometheus scrape config
│   └── grafana/
│       ├── dashboards/
│       │   └── apex_dashboard.json  # Pre-built Grafana dashboard
│       └── provisioning/
│           ├── dashboards/apex.yml  # Dashboard provisioning
│           └── datasources/prometheus.yml # Datasource config
│
├── scripts/
│   └── verify/
│       └── tactical_verify.go # End-to-end ingest verification script
│
└── templates/
    └── dashboard.html         # Legacy HTML dashboard template
```

---

## Core Concepts

### Crash Report DNA

Every crash is captured as a `CrashReport` Protobuf message containing:

| Field | Type | Description |
|-------|------|-------------|
| `error_id` | `string` | Unique UUID for the crash event |
| `message` | `string` | Error message or panic value |
| `stack_trace` | `string` | Full stack trace at time of crash |
| `timestamp` | `int64` | Unix timestamp of the crash |
| `context` | `DeviceContext` | System telemetry snapshot |
| `ai_insight` | `string` | AI-generated forensic analysis |

The `DeviceContext` captures:
- Operating system and architecture
- Total and free memory
- Battery level and charging status
- Network type (wifi, cellular, none)

Reports are batched into `BatchReport` messages for efficient transmission.

### The Vault (Local Encrypted Storage)

**Package:** `pkg/vault`

The Vault is an AES-256-GCM encrypted SQLite database that stores crash reports locally on the agent side before they are synced to the receiver. This ensures:

- **Offline resilience** -- Crashes are never lost even without network connectivity.
- **Data security** -- All crash data is encrypted at rest with a 32-byte key.
- **Automatic cleanup** -- Successfully synced reports are pruned to prevent unbounded growth.

```go
v, err := vault.New("apex.db", []byte("32-byte-encryption-key-here!!!!!"))
defer v.Close()

v.Save(report)           // Encrypt and store
reports, _ := v.FetchAll() // Decrypt and retrieve
v.Cleanup(timestamp)     // Prune old entries
```

### The Syphon (Network Sync Engine)

**Package:** `pkg/syphon`

The Syphon handles network transmission of crash batches from the agent to the receiver:

- **Batch serialization** -- Marshals `BatchReport` via Protobuf and compresses with Zstandard.
- **Network-aware sync** -- Only syncs when connected via WiFi (configurable). Cellular and offline states defer sync.
- **Retry with exponential backoff** -- Failed transmissions are retried up to 3 times with exponentially increasing delays (1s, 2s, 4s).
- **API key authentication** -- Every request includes the `X-Apex-API-Key` header.

### The Receiver (Ingest Server)

**Package:** `pkg/receiver`, **Entrypoint:** `cmd/server/main.go`

The Receiver is the central HTTP server that accepts crash report batches:

1. **Validates** the API key against stored project keys (or a default environment key).
2. **Decompresses** the Zstd-compressed payload.
3. **Deserializes** the batch (tries Protobuf first, falls back to JSON).
4. **Offloads** each report to a Redis stream for async processing.
5. A **background worker** reads from the stream, runs AI analysis, and persists to the database.

The Receiver also includes heuristic-based forensic rules for common crash patterns (nil pointer, index out of range, context deadline, etc.) as a fallback when AI is unavailable.

### AI Forensics (Gemini Integration)

**Package:** `pkg/ai`

The `TacticalAI` module integrates with Google's Gemini API to provide:

- **Automatic report analysis** (`AnalyzeReport`) -- Generates structured forensic breakdowns with `ROOT_CAUSE`, `IMPACT_ASSESSMENT`, and `TACTICAL_FIX` sections. Includes source code context and historical similar reports (RAG).
- **Interactive chat** (`Chat`, `ChatStream`) -- Real-time SSE-streamed conversations about specific crash reports with source-code awareness.
- **Insight caching** -- AI responses are cached in Redis for 24 hours using a SHA-256 fingerprint of the error message and stack trace.
- **Rate limiting** -- 100 AI analyses per hour per project, 10 chat messages per hour per report.

### Rate Limiting

**Package:** `pkg/limiter`

A Redis-backed sliding window rate limiter that protects AI quota and prevents abuse:

```go
limiter := limiter.NewRateLimiter(redisClient)

// Allow 100 requests per hour for this key
allowed, err := limiter.Allow(ctx, "project123:analysis", 100, 1*time.Hour)

// Check remaining quota
remaining, err := limiter.GetRemaining(ctx, "project123:analysis", 100)
```

The limiter fails closed (denies) when Redis is unavailable to protect downstream AI resources.

---

## Agent Integration

### Go Agent

The Go agent provides native crash capture with `defer`-based panic recovery:

```go
package main

import (
    "time"
    "github.com/Segniko/Apex/pkg/agent"
    "github.com/Segniko/Apex/pkg/syphon"
    "github.com/Segniko/Apex/pkg/vault"
)

func main() {
    // 1. Initialize encrypted local storage
    v, _ := vault.New("apex.db", []byte("your-32-byte-encryption-key!!!!!"))
    defer v.Close()

    // 2. Initialize network sync engine
    s, _ := syphon.New(nil) // nil = always sync

    // 3. Configure the agent
    cfg := agent.DefaultConfig()
    cfg.IngestURL = "https://apex.vercel.app/api/ingest" // Or your self-hosted receiver
    cfg.APIKey = "YOUR_INGEST_KEY"
    cfg.SyncInterval = 30 * time.Second

    // 4. Start the agent
    a := agent.New(v, s, cfg)
    defer a.Stop()

    // 5. Protect your code
    defer a.CapturePanic()

    // Your application code here...
}
```

**Configuration options (`agent.Config`):**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `IngestURL` | `string` | `""` | Receiver endpoint URL |
| `APIKey` | `string` | `""` | Project ingest key |
| `SyncInterval` | `time.Duration` | `1m` | How often to check vault and sync |
| `BatchSize` | `int` | `50` | Max reports per sync batch |

### Python Agent

```python
from agents.python.agent import ApexAgent
import traceback

agent = ApexAgent(
    ingest_url="https://apex.vercel.app/api/ingest",
    api_key="YOUR_INGEST_KEY"
)

try:
    # Your application code...
    risky_operation()
except Exception as e:
    agent.capture_exception(e, traceback.format_exc())
```

**Dependencies:** `requests`, `zstandard`

```bash
pip install requests zstandard
```

### Node.js Agent

```javascript
const { ApexAgent } = require('./agents/node/agent');

const agent = new ApexAgent(
    "https://apex.vercel.app/api/ingest",
    "YOUR_INGEST_KEY"
);

try {
    // Your application code...
    riskyOperation();
} catch (error) {
    await agent.captureException(error);
}
```

**Dependencies:** `uuid`, `fzstd`

```bash
cd agents/node
npm install
```

---

## API Reference

The Receiver exposes the following HTTP endpoints (default port `8081`):

### Ingest

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/ingest` | Accept a compressed batch of crash reports |

**Headers:**
- `X-Apex-API-Key` (required): Project ingest key
- `Content-Type`: `application/x-protobuf` or `application/octet-stream`

**Body:** Zstd-compressed Protobuf or JSON `BatchReport`

**Response:** `202 Accepted` with count of reports received

### Reports

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/reports` | HTML dashboard view of crash reports |
| `GET` | `/api/reports?project_id={id}` | JSON array of crash reports (up to 50) |

### Projects

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/projects?user_id={id}` | List projects for a user |
| `POST` | `/api/projects/create` | Create a new project |

**Create Project Body:**
```json
{
    "user_id": "github-user-id",
    "name": "My Production API"
}
```

**Response:**
```json
{
    "id": "uuid",
    "user_id": "github-user-id",
    "name": "My Production API",
    "ingest_key": "apex_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "created_at": "2026-03-17T00:00:00Z"
}
```

### AI Chat

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/chat` | Stream an AI response about a crash report |

**Body:**
```json
{
    "message": "Why did this nil pointer dereference happen?",
    "report_id": "crash-report-uuid"
}
```

**Response:** Server-Sent Events (SSE) stream. Each event is `data: <text>\n\n`. Stream ends with `data: [DONE]\n\n`.

### Status & Metrics

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/status` | Server status and storage mode |
| `GET` | `/metrics` | Prometheus metrics endpoint |

**Status Response:**
```json
{
    "persistent": true,
    "status": "operational",
    "timestamp": 1773785000
}
```

**Prometheus Metrics:**
- `apex_reports_received_total` -- Total crash reports received
- `apex_ingest_errors_total` -- Total ingestion errors
- `apex_ingest_duration_seconds` -- Time spent unpacking and routing batches

---

## Dashboard

The dashboard is a Next.js 16 application with a tactical/industrial design theme built with Tailwind CSS v4. It provides:

### Pages

| Route | Description |
|-------|-------------|
| `/` | Landing page with feature overview and integration quickstart |
| `/auth/login` | GitHub OAuth login page |
| `/dashboard` | Global crash report feed with real-time polling (5s interval) |
| `/dashboard/projects` | Project management hub -- create projects and copy ingest keys |
| `/dashboard/projects/[id]` | Per-project crash HUD with isolated crash feed |
| `/docs` | Technical documentation and development changelog |

### Key Components

- **CrashCard** -- Displays crash reports with telemetry grid (OS, architecture, memory, battery), syntax-highlighted stack traces, "Chat about this Error" action, and AI insight panel.
- **TacticalChat** -- Floating AI chat widget with SSE streaming, code block rendering with diff syntax highlighting, and crash-report-aware context.
- **UserButton** -- Displays authenticated user info with session termination.
- **MobileBlocker** -- Restricts access on mobile devices (desktop-only application).

### Authentication

The dashboard uses NextAuth.js v5 with GitHub as the OAuth provider. Protected routes are enforced via middleware. Configuration requires:

- `AUTH_SECRET` -- NextAuth secret key
- `AUTH_GITHUB_ID` -- GitHub OAuth App Client ID
- `AUTH_GITHUB_SECRET` -- GitHub OAuth App Client Secret

---

## Observability Stack

Apex includes a pre-configured observability stack:

### Prometheus

Scrapes the Receiver's `/metrics` endpoint every 15 seconds. Configuration at `deploy/prometheus/prometheus.yml`.

### Grafana

Pre-provisioned with:
- **Prometheus datasource** auto-configured at `deploy/grafana/provisioning/datasources/prometheus.yml`
- **Apex dashboard** at `deploy/grafana/dashboards/apex_dashboard.json`

Access Grafana at **http://localhost:3001** with credentials `admin` / `admin`.

---

## Deployment

### Docker Compose (Self-Hosted)

The full stack is defined in `docker-compose.yml` with six services:

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
| `cockroach` | `cockroachdb/cockroach:v23.1.10` | 5433, 8082 | Primary database |
| `redis` | `redis:7-alpine` | 6379 | Ingest buffer and cache |
| `receiver` | Built from `./Dockerfile` | 8081 | Crash report receiver |
| `dashboard` | Built from `./dashboard/Dockerfile` | 3000 | Web dashboard |
| `prometheus` | `prom/prometheus:latest` | 9090 | Metrics collection |
| `grafana` | `grafana/grafana:latest` | 3001 | Metrics visualization |

**Required environment variables** (set in `.env` or pass to `docker-compose`):

```bash
GEMINI_API_KEY=your-gemini-api-key         # For AI features
APEX_API_KEY=your-default-api-key          # Default ingest key for demo mode
AUTH_SECRET=your-nextauth-secret           # NextAuth.js secret
AUTH_GITHUB_ID=your-github-oauth-id        # GitHub OAuth App ID
AUTH_GITHUB_SECRET=your-github-oauth-secret # GitHub OAuth App Secret
```

### Render

The `render.yaml` blueprint defines:
- A **web service** (`apex-backend`) running the Go receiver
- A **Redis** instance (`apex-cache`)
- A **PostgreSQL database** (`apex-db`)

Deploy to Render by connecting the repository and using the blueprint.

### Vercel (Dashboard Only)

The dashboard is configured for Vercel deployment via `dashboard/vercel.json`. Set the `NEXT_PUBLIC_API_URL` environment variable to point to your hosted receiver.

---

## Configuration

### Environment Variables

| Variable | Component | Required | Description |
|----------|-----------|----------|-------------|
| `DATABASE_URL` | Receiver | No | PostgreSQL/CockroachDB connection string. Falls back to in-memory store. |
| `REDIS_URL` | Receiver | No | Redis connection string. Defaults to `127.0.0.1:6379`. |
| `GEMINI_API_KEY` | Receiver | No | Google Gemini API key for AI features. AI runs in degraded mode without it. |
| `APEX_API_KEY` | Receiver | No | Default ingest key for demo/fallback authentication. |
| `PORT` | Receiver | No | HTTP server port. Defaults to `8081`. |
| `NEXT_PUBLIC_API_URL` | Dashboard | No | Receiver URL for API calls. Defaults to `http://localhost:8081`. |
| `AUTH_SECRET` | Dashboard | Yes | NextAuth.js secret for session encryption. |
| `AUTH_GITHUB_ID` | Dashboard | Yes | GitHub OAuth App Client ID. |
| `AUTH_GITHUB_SECRET` | Dashboard | Yes | GitHub OAuth App Client Secret. |

### Storage Modes

The Receiver supports two storage modes:

1. **Persistent Mode** (CockroachDB/PostgreSQL) -- Full production storage with SQL queries, project isolation, and similar report retrieval. Activated when `DATABASE_URL` is set and connection succeeds.
2. **Memory Mode** -- In-memory store with a 100-report ring buffer. Used as a fallback when no database is available. Data is lost on restart.

### Graceful Degradation

- **No `DATABASE_URL`** -- Falls back to Memory Mode with a warning.
- **No `REDIS_URL`** -- Connects to `127.0.0.1:6379` by default.
- **No `GEMINI_API_KEY`** -- AI features return fallback forensic analysis using built-in heuristic rules.
- **Database connection failure** -- Retries up to 5 times with 2-second intervals before falling back to Memory Mode.

---

## Contributing

Apex is 100% open source. Contributions are welcome!

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Commit your changes: `git commit -m "Add your feature"`
4. Push to your branch: `git push origin feature/your-feature`
5. Open a Pull Request

### Development Tools

| Command | Description |
|---------|-------------|
| `go run cmd/server/main.go` | Run the receiver server |
| `go run cmd/inspector/main.go` | Inspect the local vault database |
| `go run cmd/simulate/main.go` | Send simulated crash reports |
| `go run scripts/verify/tactical_verify.go` | End-to-end ingest verification |
| `cd dashboard && npm run dev` | Run the dashboard in development mode |
| `cd dashboard && npm run lint` | Lint the dashboard code |
| `docker-compose up -d` | Start the full stack |

---

## Tactical Roadmap: The Future of Apex

Building Apex is just the beginning. I have several high-impact features planned to push the boundaries of observability:

- **eBPF Zero-Touch Monitoring** -- I want to implement eBPF-based agents that can capture context at the kernel level without modifying a single line of application code. This would allow Apex to monitor any running process, providing deep system-level telemetry (syscall failures, memory faults, network drops) with zero instrumentation overhead.

- **Self-Healing Hooks** -- I plan to add "Action Hooks" that allow the AI to trigger specific recovery scripts (like restarting a service or draining a node) when it detects a known critical failure pattern. Instead of just reporting the crash, Apex would actively participate in recovery.

- **Predictive Analytics** -- By analyzing historical crash data with Gemini, I want to build a system that alerts me when it detects a "regression pattern" before a full system outage occurs. The goal is to move from reactive forensics to proactive failure prevention.

- **Edge-AI Denoising** -- Moving smaller AI models directly into the Syphon to filter out noisy, low-signal errors before they even hit the receiver. This would reduce bandwidth, cut storage costs, and let the central AI focus on the crashes that actually matter.

---

## Building Apex: Lessons & Technical Deep Dive

This section documents the technical journey of building Apex from the ground up -- the architectural decisions, the Go concepts I mastered, the problems I hit, and how I solved them.

### Project Initialization & Package Layout

I initialized the project with `go mod init github.com/apex/monitor` and adopted the standard Go package layout to keep concerns cleanly separated:

| Package | Responsibility |
|---------|----------------|
| `pkg/agent` | The panic sensor -- captures crashes, collects context, and coordinates sync |
| `pkg/vault` | My encrypted local storage -- AES-256-GCM over SQLite for crash data at rest |
| `pkg/syphon` | My compression and sync logic -- Zstd batching with exponential backoff |
| `pkg/receiver` | The ingest engine -- decompression, deserialization, and routing |
| `pkg/ai` | The intelligence layer -- Gemini integration, RAG context, and streaming chat |
| `pkg/storage` | The persistence abstraction -- swappable backends via a `Provider` interface |
| `pkg/limiter` | Quota enforcement -- Redis-backed sliding window rate limiting |
| `proto/` | Where I keep my Protobuf definitions and generated Go code |

This layout made it straightforward to develop, test, and reason about each component independently.

### Why I Chose Protobuf

I found that Protocol Buffers produce significantly smaller payloads than JSON, which proved vital when I tested it in high-cost data regions. A typical crash report serialized as JSON weighs ~800 bytes; the same report in Protobuf is ~350 bytes. When you're ingesting thousands of reports per minute across paid network links, that 56% reduction adds up fast. Protobuf also gives me strict schema enforcement -- if a field type changes, the compiler catches it before runtime does.

### Compiler Chain & Dependencies

I use the following tools to manage my Protobuf environment:

```bash
# Install the Go Protobuf code generator
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Compile the schema to Go code
protoc --go_out=. --go_opt=paths=source_relative proto/apex.proto

# Import the Protobuf runtime
go get google.golang.org/protobuf/proto
```

The generated `proto/apex.pb.go` file is checked into the repository so that contributors don't need `protoc` installed to build the project.

### Key Go Concepts I Mastered

Building Apex forced me to go deep on several core Go concepts:

- **`defer` + `recover` for robust panic handling** -- The entire agent crash-capture system is built on deferred recovery. I learned to use `recover()` inside a deferred function to intercept panics, extract the panic value, capture a stack trace with `runtime.Stack()`, and then re-raise the panic so the application's normal crash behavior is preserved.

- **Pointers (`*` and `&`) for memory efficiency** -- Crash reports and device contexts are passed as pointers throughout the pipeline to avoid unnecessary copying. The Protobuf-generated code uses pointer receivers extensively, and I learned to work with this idiom naturally.

- **The global module cache and interfaces for testability** -- I designed the `storage.Provider` interface specifically so I could swap between CockroachDB in production and an in-memory store for development and testing, without changing a single line of business logic.

- **Goroutines and channels for the background sync engine** -- The agent's sync loop runs as a goroutine with a `time.Ticker`. I used channels and `select` statements to coordinate the sync timer with graceful shutdown signals.

- **`log/slog` for structured, tactical logging** -- Every log statement in the receiver uses structured fields (`slog.String`, `slog.Int`, `slog.Error`) so that logs are machine-parseable and can be piped into any observability stack.

### Advanced Features & Encounters

Some of the more challenging technical work I did:

- **AES-GCM encryption in the Vault** -- I implemented AES-256-GCM authenticated encryption to ensure crash data at rest is secure. Each report is encrypted with a unique random nonce to prevent nonce-reuse attacks. The encryption key must be exactly 32 bytes, which I enforce at initialization time.

- **Docker networking wall** -- I hit a connectivity wall early on with Docker. Containers couldn't reach each other using `localhost`. I fixed this by switching from `localhost` to `127.0.0.1` for my database and Redis connection strings in development, and using Docker service names (`cockroach`, `redis`) in the Compose network.

- **Deep memory diagnostics** -- I added `runtime.ReadMemStats` to provide deep memory diagnostics in every crash report. The agent captures total allocated memory, free memory, and GC statistics at the exact moment of the crash, giving me a precise snapshot of the system state.

- **Cyber-black dashboard** -- I built the Next.js dashboard with a strict industrial "cyber-black" theme using Tailwind CSS v4. The design system is built around amber (`#FFB800`), near-black backgrounds (`#080808`), and green (`#00FF41`) accents -- explicitly zero blue or pink anywhere in the interface.

### Operational Excellence

The final phase of the project was building a complete observability stack around the receiver:

1. **Prometheus instrumentation** -- I integrated the `prometheus/client_golang` library into the Go receiver, registering counters for `apex_reports_received_total` and `apex_ingest_errors_total`, plus a histogram for `apex_ingest_duration_seconds` to track processing latency.

2. **`/metrics` endpoint** -- I exposed a dedicated `/metrics` endpoint that Prometheus scrapes every 15 seconds. This gives me real-time visibility into ingest throughput, error rates, and latency distributions.

3. **Unified Docker Compose** -- I created a single `docker-compose.yml` that brings up my entire forensics suite with one command: CockroachDB, Redis, the receiver, the dashboard, Prometheus, and Grafana -- six services, zero manual configuration.

4. **Grafana dashboard** -- I configured a pre-provisioned Grafana dashboard at `deploy/grafana/dashboards/apex_dashboard.json` that visualizes all my Prometheus metrics with auto-refreshing panels. It provisions automatically on first boot via Grafana's YAML provisioning system.

---

## License

This project is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for details.

Copyright (c) 2026 Segniko     
