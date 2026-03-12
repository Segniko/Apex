# APEX: THE ARCHITECTURE OF RECOVERY

**Industrial grade failure forensics. 100% Free. 100% Open Source.**

Apex is a high performance monitoring engine built to survive where others fail. Designed with a "Community First" philosophy, it brings trillion-dollar infrastructure capabilities to every developer's terminal. 

## Zero to HUD in 60 Seconds (Quickstart)

The fastest way to get started is to use our free centralized Command Center.

1. **Access the Terminal**: Head over to [https://apex.vercel.app](https://apex.vercel.app)
2. **Get your API Key**: Create a project and secure your unique API Key.
3. **Integrate**: Drop the Agent into your codebase (see below).

## Self-Hosting (Docker Quickstart)

If you require total data sovereignty, you can spin up the entire Apex engine (Receiver, Redis Buffer, CockroachDB, and HUD) locally or on your own servers:

1. **Clone & Launch**:
   ```bash
   git clone https://github.com/Segniko/Apex.git
   cd Apex
   docker-compose up -d
   ```

2. **Access Your Local Terminal**:
   - **Command Center**: `http://localhost:3000`
   - **Metrics HUD**: `http://localhost:3000/metrics`

## Integrating the Tactical Edge (Agents)

Apex works by dropping a lightweight "Agent" into your code. When a crash occurs, the agent packs, compresses, and syncs the DNA of the failure to your Apex Receiver.

### Go Integration (Local Package)
```go
import "github.com/Segniko/Apex/pkg/agent" // Update with your actual repo path

func main() {
    // Points to the central cloud or your self-hosted receiver
    apex := agent.New("https://apex.vercel.app/api/ingest", "YOUR_API_KEY")
    defer apex.Recover() // Uncrashable start.
    
    // Your mission-critical code here...
}
```

### Python Integration (Local Script)
```python
# Import the agent.py file directly into your project for now
from agent import ApexAgent

agent = ApexAgent(url="https://apex.vercel.app/api/ingest", key="YOUR_API_KEY")

try:
    # Operations...
except Exception as e:
    agent.sync(e) # Captured. Decoded. Synced.
```

### Node.js Integration
```javascript
const { ApexAgent } = require('./agents/node/agent');

const agent = new ApexAgent("https://apex.vercel.app/api/ingest", "YOUR_API_KEY");

try {
    // Your logic...
} catch (error) {
    agent.captureException(error); // Captured. Compressed. Synced.
}
```

## Managed AI & Security Architecture
Apex is designed as a **Managed AI Provider**. This means the host (you) provides the AI "brain" (Gemini) while ensuring that your API keys are never exposed and your usage quotas are protected.

- **Vaulted Secrets**: Your `GEMINI_API_KEY` stays on the server as an environment variable. It never travels to the browser or client agents.
- **Intelligent Caching**: Duplicate crashes are identified by their fingerprint (SHA-256). We only run AI analysis once per unique crash and serve the cached insight to all subsequent hits, preserving your quota.
- **Redis Rate Limiting**: Every project is capped on AI usage (Analysis & Chat) to prevent "Log Bombing" or accidental over-usage from draining your tokens.

## Setup & Environment
Ensure you have a `.env` file (copied from `.env.example`) with the following:
- `GEMINI_API_KEY`: Your Google AI Studio key.
- `DATABASE_URL`: Your CockroachDB/Postgres connection string.
- `REDIS_URL`: Your Redis instance address.

## Join the Movement
Apex is 100% Open Source. Every line of code, every design token, and every architectural decision belongs to the community.

- **Star on GitHub**: Help us reach more developers.

**Apex: Built for the build. Built for the recovery.** 
