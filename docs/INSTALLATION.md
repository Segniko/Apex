# Installation Guide

This guide covers all methods of setting up Apex, from the quickest cloud-hosted option to a fully self-hosted deployment.

---

## Option 1: Cloud-Hosted (Fastest)

Use the centralized Apex Command Center -- no installation required.

1. Navigate to [https://apex.vercel.app](https://apex.vercel.app)
2. Click **"Access Command Center"** and sign in with your GitHub account
3. Create a new project in the **Projects Hub**
4. Copy the generated **Ingest Key**
5. Integrate an Edge Agent into your application (see [Agent Integration](AGENTS.md))

This option uses Apex's hosted infrastructure. Your crash data is stored on managed CockroachDB and processed via managed Redis.

---

## Option 2: Docker Compose (Self-Hosted)

Run the complete Apex stack on your own infrastructure for full data sovereignty.

### Prerequisites

- Docker Engine 20.10+
- Docker Compose v2.0+
- At least 2 GB of RAM available for containers

### Steps

**1. Clone the repository:**

```bash
git clone https://github.com/Segniko/Apex.git
cd Apex
```

**2. Create an environment file:**

```bash
cat > .env << 'EOF'
# Required for AI-powered analysis (get from https://aistudio.google.com/)
GEMINI_API_KEY=your-gemini-api-key

# Default API key for demo/testing (any string)
APEX_API_KEY=my-demo-key

# Dashboard authentication (create at https://github.com/settings/developers)
AUTH_SECRET=any-random-secret-string-at-least-32-chars
AUTH_GITHUB_ID=your-github-oauth-client-id
AUTH_GITHUB_SECRET=your-github-oauth-client-secret

# Database (auto-configured by docker-compose)
DATABASE_URL=postgresql://root@cockroach:26257/defaultdb?sslmode=disable

# Redis (auto-configured by docker-compose)
REDIS_URL=redis:6379
EOF
```

**3. Launch all services:**

```bash
docker-compose up -d
```

**4. Verify the deployment:**

```bash
# Check all containers are running
docker-compose ps

# Test the receiver health
curl http://localhost:8081/api/status

# Access the dashboard
open http://localhost:3000
```

### Service Ports

| Service | URL | Purpose |
|---------|-----|---------|
| Dashboard | http://localhost:3000 | Web interface |
| Receiver | http://localhost:8081 | API and ingest endpoint |
| CockroachDB Admin | http://localhost:8082 | Database admin UI |
| Prometheus | http://localhost:9090 | Metrics query interface |
| Grafana | http://localhost:3001 | Metrics dashboards |

### Stopping the Stack

```bash
# Stop all services (preserves data)
docker-compose down

# Stop and remove all data volumes
docker-compose down -v
```

---

## Option 3: Local Development Setup

Run individual components natively for development and debugging.

### Prerequisites

- **Go 1.25+** -- [Download](https://go.dev/dl/)
- **Node.js 20+** and npm -- [Download](https://nodejs.org/)
- **Docker** (for infrastructure services only)

### Step 1: Start Infrastructure

Start only the database and cache services via Docker:

```bash
docker-compose up -d cockroach redis
```

Optionally start the monitoring stack:

```bash
docker-compose up -d prometheus grafana
```

### Step 2: Run the Receiver

```bash
# Set environment variables
export DATABASE_URL="postgresql://root@127.0.0.1:5433/defaultdb?sslmode=disable"
export REDIS_URL="127.0.0.1:6379"
export GEMINI_API_KEY="your-gemini-api-key"  # Optional
export APEX_API_KEY="dev-key"                # For testing

# Install Go dependencies and run
go mod download
go run cmd/server/main.go
```

The receiver starts on port `8081` by default. You should see:

```
INF Apex Receiver starting port=8081
INF Connected to PostgreSQL
INF Redis connected
INF Starting Redis ingest worker
```

### Step 3: Run the Dashboard

```bash
cd dashboard

# Install dependencies
npm install

# Set environment variables
export NEXT_PUBLIC_API_URL="http://localhost:8081"
export AUTH_SECRET="dev-secret-at-least-32-characters-long"
export AUTH_GITHUB_ID="your-github-oauth-id"
export AUTH_GITHUB_SECRET="your-github-oauth-secret"

# Start development server
npm run dev
```

The dashboard starts on port `3000`.

### Step 4: Verify End-to-End

Send a test crash report:

```bash
go run cmd/simulate/main.go
```

Or use the verification script:

```bash
go run scripts/verify/tactical_verify.go
```

Check the dashboard at http://localhost:3000 to see the crash report appear.

---

## Option 4: Receiver Only (Minimal)

Run just the receiver without any infrastructure. It will use in-memory storage and heuristic-based analysis (no AI):

```bash
go run cmd/server/main.go
```

No environment variables are required. The receiver will:
- Start on port 8081
- Use in-memory storage (data lost on restart)
- Accept any API key
- Provide basic heuristic analysis instead of AI

This is useful for quick testing and development.

---

## Setting Up GitHub OAuth

The dashboard requires a GitHub OAuth application for authentication:

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click **"New OAuth App"**
3. Fill in the form:
   - **Application name:** `Apex Dashboard` (or your preference)
   - **Homepage URL:** `http://localhost:3000` (or your deployed URL)
   - **Authorization callback URL:** `http://localhost:3000/api/auth/callback/github`
4. Click **"Register application"**
5. Copy the **Client ID** (set as `AUTH_GITHUB_ID`)
6. Generate a **Client Secret** (set as `AUTH_GITHUB_SECRET`)

For production, update the callback URL to your deployed dashboard URL.

---

## Setting Up Gemini AI

AI-powered crash analysis requires a Google Gemini API key:

1. Go to [Google AI Studio](https://aistudio.google.com/)
2. Sign in with your Google account
3. Navigate to **API Keys** and create a new key
4. Copy the key and set it as `GEMINI_API_KEY`

Without a Gemini API key, Apex will still function but will use built-in heuristic rules for crash analysis instead of AI.

---

## Windows Quick Start

A `run.bat` script is provided for Windows users:

```cmd
# From the repository root
run.bat
```

This script launches the receiver server. For the full stack, use Docker Compose as described above.

---

## Troubleshooting Installation

| Issue | Solution |
|-------|----------|
| `docker-compose up` fails | Ensure Docker is running and you have sufficient memory (2 GB+) |
| CockroachDB won't start | Check port 5433 is not in use. Try `docker-compose down -v` to reset |
| Receiver can't connect to database | Wait 10-15 seconds after `docker-compose up` for CockroachDB to initialize |
| Dashboard shows "Infrastructure Offline" | Ensure the receiver is running and `NEXT_PUBLIC_API_URL` is correct |
| GitHub OAuth redirect error | Verify the callback URL matches exactly: `{your-url}/api/auth/callback/github` |
| `go: module not found` | Run `go mod download` in the repository root |
| `npm install` fails in dashboard | Ensure Node.js 20+ is installed. Try deleting `node_modules` and retrying |

For more detailed troubleshooting, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md).
