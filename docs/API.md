# API Reference

The Apex Receiver exposes a REST API for crash report ingestion, data retrieval, project management, and AI-powered analysis. All endpoints are served from the receiver's HTTP server (default port `8081`).

---

## Base URL

- **Self-hosted:** `http://localhost:8081`
- **Cloud-hosted:** `https://your-deployed-receiver.com`

---

## Authentication

Most endpoints require an API key for authentication. The key is passed via the `X-Apex-API-Key` HTTP header.

API keys are generated per-project when a project is created. There is also a fallback `APEX_API_KEY` environment variable that acts as a universal key for demo and development purposes.

---

## Endpoints

### POST /ingest

Accepts a compressed batch of crash reports from Edge Agents.

**Headers:**

| Header | Required | Description |
|--------|----------|-------------|
| `X-Apex-API-Key` | Yes | Project ingest key |
| `Content-Type` | No | `application/x-protobuf` or `application/octet-stream` |
| `Content-Encoding` | No | `zstd` (informational; server always attempts Zstd decompression) |

**Request Body:**

Zstd-compressed binary data containing either:
- A Protocol Buffers encoded `BatchReport` message
- A JSON encoded object: `{"reports": [...]}`

The receiver tries Protobuf deserialization first, then falls back to JSON.

**Crash Report Fields (per report in batch):**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `error_id` | string | Yes | Unique UUID for the crash |
| `message` | string | Yes | Error message or panic value |
| `stack_trace` | string | Yes | Full stack trace |
| `timestamp` | int64 | Yes | Unix timestamp (seconds) |
| `context.os` | string | No | Operating system |
| `context.arch` | string | No | CPU architecture |
| `context.total_memory` | uint64 | No | Total system memory (bytes) |
| `context.free_memory` | uint64 | No | Free system memory (bytes) |
| `context.battery_level` | float | No | Battery level (0.0 - 1.0) |
| `context.is_charging` | bool | No | Whether device is charging |
| `context.network_type` | string | No | Network type (wifi/cellular/none) |

**Response:**

- **202 Accepted** -- Reports received and queued for processing

```json
{
    "status": "received",
    "count": 3
}
```

- **401 Unauthorized** -- Invalid or missing API key
- **400 Bad Request** -- Decompression or deserialization failure

**Example (curl):**

```bash
# Prepare a test report
echo '{"reports":[{"error_id":"test-001","message":"nil pointer","stack_trace":"main.go:42","timestamp":1773785000,"context":{"os":"linux","arch":"amd64"}}]}' | zstd -c | \
curl -X POST http://localhost:8081/ingest \
  -H "X-Apex-API-Key: your-api-key" \
  -H "Content-Type: application/octet-stream" \
  --data-binary @-
```

---

### GET /api/reports

Retrieves crash reports, optionally filtered by project.

**Query Parameters:**

| Parameter | Required | Description |
|-----------|----------|-------------|
| `project_id` | No | Filter reports by project ID |

**Response:**

```json
[
    {
        "error_id": "550e8400-e29b-41d4-a716-446655440000",
        "message": "runtime error: index out of range [5] with length 3",
        "stack_trace": "goroutine 1 [running]:\nmain.processItems()\n\t/app/main.go:42 +0x1a4\nmain.main()\n\t/app/main.go:15 +0x68",
        "timestamp": 1773785000,
        "context": {
            "os": "linux",
            "arch": "amd64",
            "total_memory": 8589934592,
            "free_memory": 4294967296,
            "battery_level": 1.0,
            "is_charging": true,
            "network_type": "wifi"
        },
        "ai_insight": "ROOT_CAUSE: Array index 5 exceeds bounds of slice with length 3...",
        "user_id": "",
        "project_id": "project-uuid"
    }
]
```

Returns up to 50 reports, ordered by most recent.

---

### GET /reports

Returns an HTML page rendering crash reports in a dashboard view. This is a legacy endpoint using Go's `html/template` rendering.

---

### GET /api/projects

Lists all projects belonging to a user.

**Query Parameters:**

| Parameter | Required | Description |
|-----------|----------|-------------|
| `user_id` | Yes | The user's identifier (GitHub ID) |

**Response:**

```json
[
    {
        "id": "project-uuid",
        "user_id": "github-user-id",
        "name": "Production API",
        "ingest_key": "apex_550e8400-e29b-41d4-a716-446655440000",
        "created_at": "2026-03-17T12:00:00Z"
    }
]
```

---

### POST /api/projects/create

Creates a new project and generates a unique ingest key.

**Request Body:**

```json
{
    "user_id": "github-user-id",
    "name": "My New Project"
}
```

**Response:**

- **200 OK**

```json
{
    "id": "generated-uuid",
    "user_id": "github-user-id",
    "name": "My New Project",
    "ingest_key": "apex_generated-uuid",
    "created_at": "2026-03-17T12:00:00Z"
}
```

- **400 Bad Request** -- Missing or invalid fields

The generated `ingest_key` follows the format `apex_{uuid}` and is unique across all projects.

---

### POST /api/chat

Initiates an AI-powered conversation about a specific crash report. Responses are streamed via Server-Sent Events (SSE).

**Request Body:**

```json
{
    "message": "What caused this crash and how do I fix it?",
    "report_id": "crash-report-error-id"
}
```

**Response:**

Content-Type: `text/event-stream`

```
data: The crash was caused by

data: a nil pointer dereference

data: in the `processItems` function

data: at line 42...

data: [DONE]
```

Each `data:` line contains a chunk of the AI response. The stream terminates with `data: [DONE]`.

**Rate Limits:**

- 10 chat messages per hour per report ID
- Returns a rate limit message if exceeded

**Error Responses:**

- **400 Bad Request** -- Missing message or report_id
- **429 Too Many Requests** -- Rate limit exceeded (returned as SSE message)
- **500 Internal Server Error** -- AI service failure (returned as SSE message)

---

### GET /api/status

Returns the current server status and infrastructure state.

**Response:**

```json
{
    "persistent": true,
    "status": "operational",
    "timestamp": 1773785000
}
```

| Field | Type | Description |
|-------|------|-------------|
| `persistent` | boolean | `true` if using database storage, `false` if in-memory |
| `status` | string | Always `"operational"` when server is running |
| `timestamp` | int64 | Current server Unix timestamp |

---

### GET /metrics

Prometheus-compatible metrics endpoint. Returns metrics in the Prometheus exposition format.

**Response:**

```
# HELP apex_reports_received_total Total number of crash reports received
# TYPE apex_reports_received_total counter
apex_reports_received_total 1547

# HELP apex_ingest_errors_total Total number of ingestion errors
# TYPE apex_ingest_errors_total counter
apex_ingest_errors_total 3

# HELP apex_ingest_duration_seconds Time spent processing ingest batches
# TYPE apex_ingest_duration_seconds histogram
apex_ingest_duration_seconds_bucket{le="0.005"} 1200
apex_ingest_duration_seconds_bucket{le="0.01"} 1400
apex_ingest_duration_seconds_bucket{le="0.025"} 1500
apex_ingest_duration_seconds_bucket{le="0.05"} 1540
apex_ingest_duration_seconds_bucket{le="0.1"} 1547
apex_ingest_duration_seconds_bucket{le="+Inf"} 1547
apex_ingest_duration_seconds_sum 5.234
apex_ingest_duration_seconds_count 1547
```

---

## CORS

The receiver sets the following CORS headers on all responses:

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, X-Apex-API-Key
```

This allows the Next.js dashboard (which may run on a different origin) to communicate with the receiver.

---

## Error Handling

All error responses follow a consistent format:

```json
{
    "error": "description of the error"
}
```

Common HTTP status codes:

| Code | Meaning |
|------|---------|
| 200 | Success |
| 202 | Accepted (ingest) |
| 400 | Bad request (malformed input) |
| 401 | Unauthorized (invalid API key) |
| 405 | Method not allowed |
| 429 | Rate limit exceeded |
| 500 | Internal server error |

---

## Wire Format

### Protocol Buffers Schema

The canonical schema is defined in `proto/apex.proto`:

```protobuf
syntax = "proto3";
package apex;
option go_package = "github.com/apex/monitor/proto";

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

### JSON Equivalent

Non-Go agents that don't have Protobuf tooling can send JSON instead. The receiver will automatically detect and parse JSON when Protobuf deserialization fails:

```json
{
    "reports": [
        {
            "error_id": "uuid",
            "message": "error message",
            "stack_trace": "stack trace text",
            "timestamp": 1773785000,
            "context": {
                "os": "linux",
                "arch": "amd64",
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

The JSON must be Zstd-compressed before transmission, identical to the Protobuf workflow.
