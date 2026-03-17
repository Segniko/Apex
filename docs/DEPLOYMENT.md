# Deployment Guide

This guide covers deploying Apex to various environments, from local Docker to cloud platforms.

---

## Deployment Options

| Method | Best For | Components |
|--------|----------|------------|
| [Docker Compose](#docker-compose) | Self-hosted, full control | All services |
| [Render](#render) | Managed cloud, easy setup | Receiver + DB + Redis |
| [Vercel](#vercel) | Dashboard hosting | Dashboard only |
| [Manual](#manual-deployment) | Custom infrastructure | Individual services |

---

## Docker Compose

The recommended self-hosted deployment method. Runs all six services with a single command.

### Architecture

```
docker-compose.yml
├── cockroach     (CockroachDB v23.1.10)
├── redis         (Redis 7 Alpine)
├── receiver      (Go receiver, built from Dockerfile)
├── dashboard     (Next.js, built from dashboard/Dockerfile)
├── prometheus    (Prometheus, config from deploy/)
└── grafana       (Grafana, dashboards from deploy/)
```

### Production Deployment

**1. Clone and configure:**

```bash
git clone https://github.com/Segniko/Apex.git
cd Apex
```

**2. Create production environment file:**

```bash
cat > .env << 'EOF'
# AI Features
GEMINI_API_KEY=your-production-gemini-key

# Default API key (set to a strong random value)
APEX_API_KEY=$(openssl rand -hex 32)

# Dashboard Auth
AUTH_SECRET=$(openssl rand -hex 32)
AUTH_GITHUB_ID=your-github-oauth-id
AUTH_GITHUB_SECRET=your-github-oauth-secret

# Database (internal Docker network)
DATABASE_URL=postgresql://root@cockroach:26257/defaultdb?sslmode=disable

# Redis (internal Docker network)
REDIS_URL=redis:6379
EOF
```

**3. Deploy:**

```bash
docker-compose up -d
```

**4. Verify:**

```bash
# Check all services
docker-compose ps

# Check receiver health
curl http://localhost:8081/api/status

# Check dashboard
curl -s http://localhost:3000 | head -1
```

### Service Configuration

| Service | External Port | Internal Port | Volumes |
|---------|--------------|---------------|---------|
| `cockroach` | 5433, 8082 | 26257, 8080 | `cockroach-data:/cockroach/cockroach-data` |
| `redis` | 6379 | 6379 | None (ephemeral) |
| `receiver` | 8081 | 8081 | None |
| `dashboard` | 3000 | 3000 | None |
| `prometheus` | 9090 | 9090 | `deploy/prometheus:/etc/prometheus` |
| `grafana` | 3001 | 3000 | Grafana provisioning directories |

### Scaling

To run multiple receiver instances behind a load balancer:

```bash
docker-compose up -d --scale receiver=3
```

All receivers share the same CockroachDB and Redis instances, so horizontal scaling is straightforward.

### Updating

```bash
git pull
docker-compose build
docker-compose up -d
```

### Backup

CockroachDB data is stored in the `cockroach-data` Docker volume:

```bash
# Export database
docker-compose exec cockroach cockroach dump defaultdb --insecure > backup.sql

# Or back up the volume directly
docker run --rm -v apex_cockroach-data:/data -v $(pwd):/backup \
  alpine tar czf /backup/cockroach-backup.tar.gz /data
```

---

## Render

Render provides managed infrastructure with automatic deployments from GitHub.

### Blueprint

The `render.yaml` file defines the deployment blueprint:

```yaml
services:
  - type: web
    name: apex-backend
    env: go
    plan: free
    buildCommand: go build -o server cmd/server/main.go
    startCommand: ./server
    envVars:
      - key: DATABASE_URL
        fromDatabase:
          name: apex-db
          property: connectionString
      - key: REDIS_URL
        fromService:
          type: redis
          name: apex-cache
          property: connectionString
      - key: GEMINI_API_KEY
        sync: false
      - key: APEX_API_KEY
        sync: false

databases:
  - name: apex-db
    plan: free
    databaseName: apex
    user: apex

redis:
  - name: apex-cache
    plan: free
```

### Deployment Steps

1. Push the repository to GitHub.
2. Go to [Render Dashboard](https://dashboard.render.com/).
3. Click **"New"** > **"Blueprint"**.
4. Connect your GitHub repository.
5. Select the repository and confirm.
6. Set environment variables (`GEMINI_API_KEY`, `APEX_API_KEY`).
7. Click **"Apply"**.

Render will:
- Build the Go receiver
- Provision a PostgreSQL database
- Provision a Redis instance
- Wire all connection strings automatically

### Custom Domain

In the Render dashboard:
1. Go to your web service settings
2. Click **"Custom Domains"**
3. Add your domain and configure DNS

---

## Vercel

Vercel is used to host the Next.js dashboard only. The receiver must be deployed separately.

### Dashboard Deployment

**1. Deploy via Vercel CLI:**

```bash
cd dashboard
npx vercel --prod
```

**2. Or connect via Vercel Dashboard:**

1. Go to [vercel.com](https://vercel.com)
2. Import the repository
3. Set the root directory to `dashboard`
4. Configure environment variables

### Required Environment Variables

| Variable | Value |
|----------|-------|
| `NEXT_PUBLIC_API_URL` | URL of your deployed receiver (e.g., `https://apex-backend.onrender.com`) |
| `AUTH_SECRET` | Random secret string (32+ characters) |
| `AUTH_GITHUB_ID` | GitHub OAuth App Client ID |
| `AUTH_GITHUB_SECRET` | GitHub OAuth App Client Secret |

### Vercel Configuration

The `dashboard/vercel.json` configures the deployment:

```json
{
    "framework": "nextjs",
    "buildCommand": "npm run build",
    "outputDirectory": ".next"
}
```

### Important Notes

- The dashboard must be able to reach the receiver's API. Ensure the receiver has CORS configured (it does by default with `Access-Control-Allow-Origin: *`).
- Update your GitHub OAuth callback URL to match the Vercel deployment URL: `https://your-app.vercel.app/api/auth/callback/github`.

---

## Manual Deployment

For custom infrastructure setups.

### Receiver

**Build the binary:**

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o apex-receiver cmd/server/main.go
```

**Run with systemd:**

```ini
# /etc/systemd/system/apex-receiver.service
[Unit]
Description=Apex Crash Receiver
After=network.target

[Service]
Type=simple
User=apex
Environment=DATABASE_URL=postgresql://user:pass@localhost:5432/apex
Environment=REDIS_URL=localhost:6379
Environment=GEMINI_API_KEY=your-key
Environment=PORT=8081
ExecStart=/opt/apex/apex-receiver
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable apex-receiver
sudo systemctl start apex-receiver
```

### Dashboard

**Build the static assets:**

```bash
cd dashboard
npm install
npm run build
```

**Run with PM2:**

```bash
npm install -g pm2
pm2 start npm --name "apex-dashboard" -- start
pm2 save
pm2 startup
```

### Database

**PostgreSQL:**

```bash
# Install PostgreSQL
sudo apt install postgresql

# Create database
sudo -u postgres createdb apex
sudo -u postgres psql -c "CREATE USER apex WITH PASSWORD 'your-password';"
sudo -u postgres psql -c "GRANT ALL ON DATABASE apex TO apex;"
```

**CockroachDB:**

```bash
# Download and install
curl https://binaries.cockroachdb.com/cockroach-v23.1.10.linux-amd64.tgz | tar xz
sudo cp cockroach-v23.1.10.linux-amd64/cockroach /usr/local/bin/

# Start single-node (development)
cockroach start-single-node --insecure --listen-addr=:26257 --http-addr=:8082 --store=path=/var/lib/cockroach
```

### Redis

```bash
# Install Redis
sudo apt install redis-server

# Ensure it's running
sudo systemctl enable redis-server
sudo systemctl start redis-server
```

---

## Reverse Proxy

For production, place a reverse proxy (nginx, Caddy, Traefik) in front of the receiver and dashboard.

### Nginx Example

```nginx
server {
    listen 443 ssl;
    server_name api.apex.example.com;

    ssl_certificate /etc/ssl/certs/apex.pem;
    ssl_certificate_key /etc/ssl/private/apex.key;

    location / {
        proxy_pass http://localhost:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # SSE support for /api/chat
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 300s;
    }
}

server {
    listen 443 ssl;
    server_name apex.example.com;

    ssl_certificate /etc/ssl/certs/apex.pem;
    ssl_certificate_key /etc/ssl/private/apex.key;

    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

**Important:** Disable `proxy_buffering` for the `/api/chat` endpoint to support SSE streaming.

---

## Health Checks

Use the `/api/status` endpoint for health monitoring:

```bash
# Basic health check
curl -f http://localhost:8081/api/status || echo "Receiver is down"

# Check if database is connected
curl -s http://localhost:8081/api/status | jq '.persistent'
```

### Docker Health Check

Add to `docker-compose.yml`:

```yaml
receiver:
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:8081/api/status"]
    interval: 30s
    timeout: 10s
    retries: 3
    start_period: 10s
```

---

## Security Considerations

1. **Never expose CockroachDB or Redis ports** to the public internet. Use Docker internal networking or firewall rules.
2. **Use HTTPS** in production. The receiver and dashboard should be behind a TLS-terminating reverse proxy.
3. **Rotate API keys** periodically. Create new projects and migrate agents to new keys.
4. **Set strong `AUTH_SECRET`** values. Use `openssl rand -hex 32` to generate.
5. **Limit Gemini API usage** via the built-in rate limiter to prevent quota exhaustion.
6. **Back up CockroachDB** regularly if using Docker volumes for persistence.
