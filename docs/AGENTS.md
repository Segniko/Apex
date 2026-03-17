# Agent Integration Guide

Apex Edge Agents are lightweight libraries that capture crashes, exceptions, and panics from your applications and transmit them to the Apex Receiver. This guide covers integration for all three supported languages.

---

## How Agents Work

Every agent follows the same core workflow:

1. **Capture** -- Intercept a crash (panic, exception, error) and collect device/system context.
2. **Package** -- Create a `CrashReport` with error details, stack trace, and telemetry.
3. **Compress** -- Serialize the report (Protobuf or JSON) and compress with Zstandard.
4. **Transmit** -- Send the compressed batch to the Apex Receiver via HTTP POST.

The Go agent adds two additional capabilities:
- **Local Vault** -- Encrypted local storage that persists reports across network outages.
- **Background Sync** -- A periodic sync loop that batches and transmits stored reports.

---

## Go Agent

The Go agent is the most feature-rich, providing automatic panic recovery, encrypted local storage, and background sync.

### Installation

The Go agent is part of the main Apex module:

```bash
go get github.com/Segniko/Apex
```

### Quick Start

```go
package main

import (
    "fmt"
    "time"

    "github.com/Segniko/Apex/pkg/agent"
    "github.com/Segniko/Apex/pkg/syphon"
    "github.com/Segniko/Apex/pkg/vault"
)

func main() {
    // Initialize the encrypted local vault
    v, err := vault.New("apex_crashes.db", []byte("your-32-byte-encryption-key!!!!!"))
    if err != nil {
        panic(err)
    }
    defer v.Close()

    // Initialize the network sync engine
    s, err := syphon.New(nil) // nil checker = always sync
    if err != nil {
        panic(err)
    }

    // Configure the agent
    cfg := agent.DefaultConfig()
    cfg.IngestURL = "http://localhost:8081/ingest"
    cfg.APIKey = "your-project-ingest-key"
    cfg.SyncInterval = 30 * time.Second
    cfg.BatchSize = 25

    // Create and start the agent
    a := agent.New(v, s, cfg)
    defer a.Stop()

    // Protect your entire main function from panics
    defer a.CapturePanic()

    // Your application logic here
    fmt.Println("Application running with Apex monitoring...")
    riskyOperation()
}

func riskyOperation() {
    // This panic will be captured by Apex
    var items []string
    fmt.Println(items[5]) // index out of range
}
```

### Configuration

The `agent.Config` struct controls agent behavior:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `IngestURL` | `string` | `""` | URL of the Apex Receiver's `/ingest` endpoint |
| `APIKey` | `string` | `""` | Project ingest key from the Apex dashboard |
| `SyncInterval` | `time.Duration` | `1m` | How often to check the vault and sync unsent reports |
| `BatchSize` | `int` | `50` | Maximum number of reports per sync transmission |

Use `agent.DefaultConfig()` to get sensible defaults, then override what you need.

### How It Works

**Panic Capture:**

```go
defer a.CapturePanic()
```

This deferred call uses `recover()` to catch panics. When a panic occurs:

1. The panic value is converted to an error message.
2. A full stack trace is captured via `runtime.Stack()`.
3. Device context is collected (OS, architecture, memory stats).
4. The report is encrypted and saved to the local Vault.
5. The panic is then re-raised so the application's normal crash behavior is preserved.

**Background Sync:**

The agent starts a goroutine that runs on a timer:

1. Fetches all reports from the Vault.
2. Splits them into batches of `BatchSize`.
3. For each batch, calls `Syphon.SendBatch()` to compress and transmit.
4. On success, cleans up the transmitted reports from the Vault.

**Graceful Shutdown:**

```go
defer a.Stop()
```

Stops the sync timer and performs one final sync attempt to flush any remaining reports.

### The Vault

The Vault provides AES-256-GCM encrypted local storage:

```go
// Create a vault with a 32-byte encryption key
v, err := vault.New("crashes.db", []byte("exactly-32-bytes-of-key-data!!!!"))

// Save a crash report (encrypted at rest)
v.Save(crashReport)

// Retrieve all stored reports (decrypted)
reports, err := v.FetchAll()

// Clean up reports older than a timestamp
v.Cleanup(cutoffTimestamp)

// Close the database
v.Close()
```

The encryption key **must** be exactly 32 bytes. Each report is encrypted with a unique random nonce.

### The Syphon

The Syphon handles compression and network transmission:

```go
// Create with optional network checker
s, err := syphon.New(networkChecker)

// Check if sync should proceed (network-aware)
if s.ShouldSync() {
    // Prepare and send a batch
    err := s.SendBatch(ingestURL, apiKey, reports)
}
```

**Network types:**
- `syphon.NetworkWifi` -- Always syncs
- `syphon.NetworkCellular` -- Defers sync
- `syphon.NetworkNone` -- Never syncs
- `syphon.NetworkUnknown` -- Always syncs (default)

### Advanced: Manual Capture

You can manually capture errors without panic recovery:

```go
import (
    "github.com/Segniko/Apex/proto"
    "github.com/google/uuid"
    "runtime"
    "time"
)

report := &proto.CrashReport{
    ErrorId:    uuid.New().String(),
    Message:    "custom error: database connection timeout",
    StackTrace: captureStackTrace(),
    Timestamp:  time.Now().Unix(),
    Context: &proto.DeviceContext{
        Os:          runtime.GOOS,
        Arch:        runtime.GOARCH,
        TotalMemory: totalMem,
        FreeMemory:  freeMem,
        NetworkType: "wifi",
    },
}

vault.Save(report)
```

---

## Python Agent

The Python agent provides exception capture with system telemetry collection.

### Installation

Copy `agents/python/agent.py` into your project, then install dependencies:

```bash
pip install requests zstandard
```

### Quick Start

```python
import traceback
from agent import ApexAgent

# Initialize the agent
agent = ApexAgent(
    ingest_url="http://localhost:8081/ingest",
    api_key="your-project-ingest-key"
)

# Capture exceptions
try:
    result = 1 / 0
except Exception as e:
    agent.capture_exception(e, traceback.format_exc())
    print("Error captured and sent to Apex")
```

### Global Exception Handler

Integrate with Python's exception hook for automatic capture:

```python
import sys
import traceback
from agent import ApexAgent

agent = ApexAgent(
    ingest_url="http://localhost:8081/ingest",
    api_key="your-project-ingest-key"
)

def apex_exception_handler(exc_type, exc_value, exc_tb):
    tb = ''.join(traceback.format_exception(exc_type, exc_value, exc_tb))
    agent.capture_exception(exc_value, tb)
    sys.__excepthook__(exc_type, exc_value, exc_tb)

sys.excepthook = apex_exception_handler

# Now any unhandled exception is automatically captured
raise ValueError("This will be captured by Apex")
```

### How It Works

The Python agent:

1. Collects system information:
   - `platform.system()` for OS
   - `platform.machine()` for architecture
   - `os.sysconf('SC_PAGE_SIZE') * os.sysconf('SC_PHYS_PAGES')` for total memory
   - Available memory via similar system calls

2. Creates a report dictionary with UUID, error message, stack trace, and timestamp.

3. Wraps the report in a `{"reports": [...]}` batch structure.

4. Serializes to JSON, compresses with `zstandard`, and POSTs to the receiver with:
   - `X-Apex-API-Key` header
   - `Content-Type: application/octet-stream`

### Framework Integration

**Flask:**

```python
from flask import Flask
from agent import ApexAgent
import traceback

app = Flask(__name__)
agent = ApexAgent("http://localhost:8081/ingest", "your-key")

@app.errorhandler(Exception)
def handle_exception(e):
    agent.capture_exception(e, traceback.format_exc())
    return "Internal Server Error", 500
```

**Django:**

```python
# In your middleware
from agent import ApexAgent
import traceback

agent = ApexAgent("http://localhost:8081/ingest", "your-key")

class ApexMiddleware:
    def __init__(self, get_response):
        self.get_response = get_response

    def __call__(self, request):
        return self.get_response(request)

    def process_exception(self, request, exception):
        agent.capture_exception(exception, traceback.format_exc())
        return None
```

---

## Node.js Agent

The Node.js agent captures JavaScript errors with system context.

### Installation

```bash
cd agents/node
npm install
```

Or install dependencies directly in your project:

```bash
npm install uuid fzstd
```

Then copy `agents/node/agent.js` into your project.

### Quick Start

```javascript
const { ApexAgent } = require('./agent');

const agent = new ApexAgent(
    "http://localhost:8081/ingest",
    "your-project-ingest-key"
);

// Capture errors manually
try {
    undefinedFunction();
} catch (error) {
    await agent.captureException(error);
    console.log("Error captured and sent to Apex");
}
```

### Global Error Handler

**Node.js Process:**

```javascript
const { ApexAgent } = require('./agent');

const agent = new ApexAgent(
    "http://localhost:8081/ingest",
    "your-key"
);

process.on('uncaughtException', async (error) => {
    await agent.captureException(error);
    process.exit(1);
});

process.on('unhandledRejection', async (reason) => {
    const error = reason instanceof Error ? reason : new Error(String(reason));
    await agent.captureException(error);
});
```

**Express.js:**

```javascript
const express = require('express');
const { ApexAgent } = require('./agent');

const app = express();
const agent = new ApexAgent("http://localhost:8081/ingest", "your-key");

// Error handling middleware (must be last)
app.use(async (err, req, res, next) => {
    await agent.captureException(err);
    res.status(500).json({ error: 'Internal Server Error' });
});
```

### How It Works

The Node.js agent:

1. Generates a UUID v4 for the error using the `uuid` package.

2. Collects system context via the `os` module:
   - `os.platform()` for OS
   - `os.arch()` for architecture
   - `os.totalmem()` for total memory
   - `os.freemem()` for free memory

3. Constructs a report object with error message, stack trace, timestamp, and context.

4. Wraps in a `{"reports": [...]}` batch, serializes to JSON.

5. Compresses with `fzstd` (Zstandard for JavaScript).

6. Sends via `fetch()` with:
   - `X-Apex-API-Key` header
   - `Content-Type: application/octet-stream`

---

## Comparison

| Feature | Go Agent | Python Agent | Node.js Agent |
|---------|----------|-------------|---------------|
| Panic/Exception Capture | Automatic (`defer`) | Manual (`try/except`) | Manual (`try/catch`) |
| Serialization | Protocol Buffers | JSON | JSON |
| Compression | Zstandard | Zstandard | Zstandard |
| Local Storage (Vault) | AES-256-GCM SQLite | Not included | Not included |
| Background Sync | Periodic timer | Not included | Not included |
| Network Awareness | WiFi/Cellular/None | Not included | Not included |
| Retry Logic | Exponential backoff | Not included | Not included |
| System Telemetry | OS, arch, memory, battery | OS, arch, memory | OS, arch, memory |
| Dependencies | Standard library + Apex | requests, zstandard | uuid, fzstd |

---

## Best Practices

1. **Use project-specific ingest keys** -- Create a separate project in the Apex dashboard for each application or service.

2. **Capture stack traces** -- Always include full stack traces. For Python, use `traceback.format_exc()`. For Node.js, use `error.stack`.

3. **Don't swallow panics** -- The Go agent's `CapturePanic()` re-raises the panic after capture, preserving normal crash behavior. Don't wrap it in additional recovery logic.

4. **Set reasonable sync intervals** -- For the Go agent, 30-60 seconds is a good balance between timeliness and network efficiency.

5. **Protect your encryption key** -- The Vault encryption key should be treated as a secret. Use environment variables or a secrets manager, never hardcode it.

6. **Handle agent errors gracefully** -- Agent failures should never crash your application. Wrap agent initialization in error handling.
