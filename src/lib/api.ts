export interface DeviceContext {
    os: string;
    arch: string;
    total_memory: number;
    free_memory: number;
    battery_level: number;
}

export interface CrashReport {
    error_id: string;
    message: string;
    stack_trace: string;
    timestamp: number;
    context: DeviceContext;
    ai_insight: string;
}

const API_BASE = "http://localhost:8081/api";

export async function fetchReports(): Promise<CrashReport[]> {
    try {
        const res = await fetch(`${API_BASE}/reports`, { cache: 'no-store' });
        if (!res.ok) throw new Error("Failed to fetch reports");
        return await res.json();
    } catch (err) {
        console.error("API Error:", err);
        return [];
    }
}
