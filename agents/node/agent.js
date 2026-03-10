const os = require('os');
const { v4: uuidv4 } = require('uuid');
const fzstd = require('fzstd');

class ApexAgent {
    constructor(ingestUrl, apiKey) {
        this.ingestUrl = ingestUrl;
        this.apiKey = apiKey;
        this.os = os.platform();
        this.arch = os.arch();
    }

    /**
     * Capture an error and sync it to the Apex Command Center.
     * @param {Error} error - The error object to capture.
     * @param {string} stackTrace - Optional custom stack trace.
     */
    async captureException(error, stackTrace) {
        const report = {
            error_id: uuidv4(),
            message: error.message,
            stack_trace: stackTrace || error.stack,
            timestamp: Math.floor(Date.now() / 1000),
            context: {
                os: this.os,
                arch: this.arch,
                total_memory: os.totalmem(),
                free_memory: os.freemem(),
                battery_level: 1.0, // Mocked for prototype
                is_charging: true,
                network_type: "wifi"
            }
        };

        const batch = { reports: [report] };
        
        try {
            // 1. JSON Encode
            const data = Buffer.from(jsonStringify(batch));
            
            // 2. Compress (Zstd)
            const compressed = fzstd.compress(data);
            
            console.log(`🚀 Apex_Node: Syncing forensic trace ${report.error_id}...`);

            // 3. Sync to Command Center
            const response = await fetch(this.ingestUrl, {
                method: 'POST',
                headers: {
                    'X-Apex-API-Key': this.apiKey,
                    'Content-Type': 'application/octet-stream'
                },
                body: compressed
            });

            console.log(`✅ Apex_Node: Server Response: ${response.status}`);
        } catch (err) {
            console.error(`❌ Apex_Node: Sync failed: ${err.message}`);
        }
    }
}

// Helper to handle BigInt or other non-JSON types if necessary
function jsonStringify(obj) {
    return JSON.stringify(obj, (key, value) =>
        typeof value === 'bigint' ? value.toString() : value
    );
}

module.exports = { ApexAgent };

// Simulation if run directly
if (require.main === module) {
    const agent = new ApexAgent("http://localhost:8081/ingest", "apex-prod-key-12345");

    console.log("--- Simulating Node.js Fatal Error ---");
    try {
        // Simulate a bug
        const obj = {};
        console.log(obj.missing.property); // Causes TypeError
    } catch (e) {
        agent.captureException(e);
    }
}
