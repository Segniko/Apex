# Troubleshooting

Common issues and their solutions when working with Apex.

---

## Receiver Issues

### Receiver won't start

**Symptom:** `go run cmd/server/main.go` exits immediately or hangs.

**Solutions:**

1. **Port already in use:**
   ```bash
   # Check if port 8081 is occupied
   lsof -i :8081

   # Use a different port
   export PORT=9090
   go run cmd/server/main.go
   ```

2. **Missing Go version:**
   Apex requires Go 1.25+. Check with `go version`.

3. **Missing dependencies:**
   ```bash
   go mod download
   go mod tidy
   ```

### "Failed to connect to PostgreSQL" / Database connection errors

**Symptom:** Receiver starts but logs connection errors to the database.

**Solutions:**

1. **CockroachDB not running:**
   ```bash
   docker-compose up -d cockroach
   # Wait 10-15 seconds for initialization
   ```

2. **Wrong DATABASE_URL:**
   ```bash
   # For Docker Compose
   export DATABASE_URL="postgresql://root@127.0.0.1:5433/defaultdb?sslmode=disable"

   # For local PostgreSQL
   export DATABASE_URL="postgresql://user:password@localhost:5432/apex?sslmode=disable"
   ```

3. **CockroachDB not ready:**
   The receiver retries 5 times with 2-second intervals. If it still fails, it falls back to in-memory storage. Check CockroachDB logs:
   ```bash
   docker-compose logs cockroach
   ```

### Receiver falls back to Memory Mode

**Symptom:** Logs show "Using in-memory storage" and status API returns `{"persistent": false}`.

**Causes:**
- `DATABASE_URL` not set -- Set the environment variable
- Database unreachable -- Check that the database container is running
- Connection pool exhausted -- Restart the receiver

**Note:** Memory Mode keeps only the last 100 reports and loses all data on restart.

### Redis connection fails

**Symptom:** Logs show Redis connection errors. Reports go directly to database instead of through the stream.

**Solutions:**

1. **Redis not running:**
   ```bash
   docker-compose up -d redis
   ```

2. **Wrong REDIS_URL:**
   ```bash
   export REDIS_URL="127.0.0.1:6379"  # For local
   export REDIS_URL="redis:6379"        # For Docker internal
   ```

3. **Redis is full:**
   ```bash
   docker-compose exec redis redis-cli INFO memory
   ```

Without Redis, the receiver processes reports synchronously (no buffering) and AI caching is disabled.

---

## Dashboard Issues

### "Infrastructure Offline" banner

**Symptom:** The dashboard header shows "Infrastructure Offline (No Save)" with a red indicator.

**Cause:** The dashboard cannot reach the receiver, or the receiver is in Memory Mode.

**Solutions:**

1. Check the receiver is running and accessible:
   ```bash
   curl http://localhost:8081/api/status
   ```

2. Verify `NEXT_PUBLIC_API_URL` is set correctly:
   ```bash
   # Should match your receiver URL
   export NEXT_PUBLIC_API_URL="http://localhost:8081"
   ```

3. Check for CORS issues in browser console (F12 > Console).

### GitHub OAuth login fails

**Symptom:** Clicking "Continue with GitHub" redirects to an error page.

**Solutions:**

1. **Missing environment variables:**
   ```bash
   export AUTH_SECRET="your-secret-at-least-32-chars"
   export AUTH_GITHUB_ID="your-github-oauth-client-id"
   export AUTH_GITHUB_SECRET="your-github-oauth-client-secret"
   ```

2. **Wrong callback URL:**
   The callback URL in your GitHub OAuth App must exactly match:
   ```
   http://localhost:3000/api/auth/callback/github
   ```
   For production, replace `localhost:3000` with your deployed URL.

3. **OAuth app not authorized:**
   Go to [GitHub Developer Settings](https://github.com/settings/developers) and check your OAuth app configuration.

### Dashboard shows no crash reports

**Symptom:** The dashboard loads but shows "No Signal_Loss Detected".

**Solutions:**

1. **No reports have been sent:** Use the simulator to generate test data:
   ```bash
   go run cmd/simulate/main.go
   ```

2. **Wrong project ID:** Reports are filtered by project. Ensure the correct project is selected.

3. **API URL mismatch:** The dashboard must be able to reach the receiver. Check browser Network tab for failed requests.

4. **CORS blocked:** Check browser console for CORS errors. The receiver sets `Access-Control-Allow-Origin: *` by default.

### "SIGNAL LOSS" mobile blocker

**Symptom:** On mobile devices, the entire screen shows "SIGNAL LOSS" and the dashboard is inaccessible.

**Cause:** This is intentional. The Apex dashboard is a desktop-only application. The `MobileBlocker` component blocks access on screens smaller than the `md` breakpoint (768px).

**Solution:** Use a desktop browser or resize your browser window to be wider than 768px.

---

## Agent Issues

### Go agent: reports not syncing

**Symptom:** Crashes are captured but don't appear in the dashboard.

**Solutions:**

1. **Check IngestURL and APIKey:**
   ```go
   cfg := agent.DefaultConfig()
   cfg.IngestURL = "http://localhost:8081/ingest"  // Must include /ingest
   cfg.APIKey = "your-project-key"                 // Must match a project key
   ```

2. **Network connectivity:** The Syphon checks network type. If running locally, the default is `NetworkUnknown` which always syncs.

3. **Sync interval:** Default is 1 minute. Reports won't appear instantly:
   ```go
   cfg.SyncInterval = 10 * time.Second  // Faster for development
   ```

4. **Check vault:** Use the inspector to see stored reports:
   ```bash
   go run cmd/inspector/main.go
   ```

### Python/Node.js agent: compression errors

**Symptom:** Agent fails with Zstandard compression errors.

**Solutions:**

**Python:**
```bash
pip install zstandard
```

**Node.js:**
```bash
npm install fzstd
```

### Agent: 401 Unauthorized

**Symptom:** Agent receives 401 responses from the receiver.

**Cause:** The API key doesn't match any project key or the default `APEX_API_KEY`.

**Solutions:**

1. Create a project in the dashboard and use its ingest key.
2. Or set a default key on the receiver: `export APEX_API_KEY="your-key"`.
3. Ensure the key is passed in the `X-Apex-API-Key` header.

---

## AI / Gemini Issues

### AI insights show "HEURISTIC ANALYSIS" instead of AI

**Symptom:** Crash reports show basic pattern-matching analysis instead of detailed AI insights.

**Cause:** Gemini AI is not configured or unavailable.

**Solutions:**

1. **Set the API key:**
   ```bash
   export GEMINI_API_KEY="your-gemini-api-key"
   ```

2. **Check API quota:** Google AI Studio has rate limits. Check your usage at [Google AI Studio](https://aistudio.google.com/).

3. **Rate limited:** Apex limits AI analysis to 100 per hour per project. Check receiver logs for rate limit messages.

### AI chat not streaming

**Symptom:** The TacticalChat widget shows no response or loads indefinitely.

**Solutions:**

1. **Check receiver logs** for Gemini API errors.
2. **Check rate limit:** Chat is limited to 10 messages per hour per report.
3. **Proxy buffering:** If behind nginx, ensure `proxy_buffering off` is set for the `/api/chat` endpoint.
4. **CORS:** Check browser console for CORS errors on the streaming request.

---

## Docker Issues

### docker-compose up fails

**Symptom:** One or more services fail to start.

**Solutions:**

1. **Insufficient memory:**
   CockroachDB requires significant memory. Ensure at least 2 GB is available:
   ```bash
   docker stats
   ```

2. **Port conflicts:**
   ```bash
   # Check for port conflicts
   lsof -i :5433   # CockroachDB
   lsof -i :6379   # Redis
   lsof -i :8081   # Receiver
   lsof -i :3000   # Dashboard
   lsof -i :9090   # Prometheus
   lsof -i :3001   # Grafana
   ```

3. **Stale containers:**
   ```bash
   docker-compose down -v
   docker-compose up -d
   ```

4. **Build cache issues:**
   ```bash
   docker-compose build --no-cache
   docker-compose up -d
   ```

### CockroachDB "node is not ready"

**Symptom:** CockroachDB container starts but the receiver can't connect.

**Solution:** CockroachDB takes 10-15 seconds to initialize on first start. The receiver automatically retries 5 times. If it still fails, restart the receiver after CockroachDB is fully ready:

```bash
docker-compose restart receiver
```

---

## Metrics Issues

### Prometheus not scraping

**Symptom:** Grafana shows "No data" for Apex metrics.

**Solutions:**

1. **Check Prometheus targets:**
   Go to http://localhost:9090/targets and verify the `apex-receiver` target is UP.

2. **Receiver not exposing metrics:**
   ```bash
   curl http://localhost:8081/metrics
   ```

3. **Wrong scrape config:**
   Check `deploy/prometheus/prometheus.yml`:
   ```yaml
   scrape_configs:
     - job_name: 'apex-receiver'
       static_configs:
         - targets: ['receiver:8081']
   ```

### Grafana dashboard empty

**Symptom:** Grafana loads but the Apex dashboard shows no data.

**Solutions:**

1. **Check the datasource:**
   Go to Grafana > Configuration > Data Sources and verify Prometheus is configured and working.

2. **No traffic yet:**
   Send some test reports and wait 15-30 seconds for Prometheus to scrape:
   ```bash
   go run cmd/simulate/main.go
   ```

---

## Performance Issues

### Slow ingest response times

**Solutions:**

1. **Enable Redis buffering:** Without Redis, reports are processed synchronously. Add Redis to offload processing.

2. **Database connection pooling:** The PostgreSQL storage uses 25 max connections by default. For very high throughput, adjust in `pkg/storage/postgres.go`.

3. **Scale receivers:** Run multiple receiver instances behind a load balancer with `docker-compose up -d --scale receiver=3`.

### High memory usage

**Solutions:**

1. **Memory storage:** If using in-memory storage, it caps at 100 reports. Switch to PostgreSQL for unbounded storage.

2. **Redis memory:** Monitor Redis memory with:
   ```bash
   docker-compose exec redis redis-cli INFO memory
   ```

3. **CockroachDB cache:** CockroachDB caches aggressively. This is normal behavior but can be tuned via `--cache` flag.

---

## Getting Help

If you're still stuck:

1. Check the receiver logs: `docker-compose logs -f receiver`
2. Check the browser console (F12 > Console) for frontend errors
3. Open an issue on [GitHub](https://github.com/Segniko/Apex/issues)
