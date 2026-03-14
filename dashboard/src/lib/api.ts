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

export interface Project {
    id: string;
    user_id: string;
    name: string;
    ingest_key: string;
    created_at: string;
}

const API_BASE = (process.env.NEXT_PUBLIC_API_URL || "http://localhost:8081") + "/api";

export async function fetchReports(projectId?: string): Promise<CrashReport[]> {
    try {
        const url = projectId 
            ? `${API_BASE}/reports?project_id=${projectId}` 
            : `${API_BASE}/reports`;
        const res = await fetch(url, { cache: 'no-store' });
        if (!res.ok) throw new Error("Failed to fetch reports");
        const data = await res.json();
        return data || [];
    } catch (err) {
        console.error("API Error:", err);
        return [];
    }
}

export async function fetchProjects(userID: string): Promise<Project[]> {
    try {
        const res = await fetch(`${API_BASE}/projects?user_id=${userID}`, { cache: 'no-store' });
        if (!res.ok) throw new Error("Failed to fetch projects");
        const data = await res.json();
        return data || [];
    } catch (err) {
        console.error("API Error:", err);
        return [];
    }
}

export async function createProject(userID: string, name: string): Promise<Project | null> {
    try {
        const url = `${API_BASE}/projects/create`;
        console.log(`[APEX] Attempting to create project: ${url}`, { userID, name });
        const res = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ user_id: userID, name }),
        });
        if (!res.ok) {
            const body = await res.text().catch(() => "");
            console.error(`[APEX] API Failure: ${res.status} ${res.statusText}`, body);
            throw new Error(`Failed to create project: ${res.status} ${res.statusText} - ${body}`);
        }
        return await res.json();
    } catch (err) {
        console.error("[APEX] API Exception:", err);
        return null;
    }
}

export async function fetchStatus(): Promise<{ persistent: boolean }> {
    try {
        const url = `${API_BASE}/status`;
        const res = await fetch(url, { cache: 'no-store' });
        if (!res.ok) {
            console.warn(`[APEX] Status Node Unreachable: ${res.status}`);
            throw new Error("Failed to fetch status");
        }
        return await res.json();
    } catch (err) {
        console.error("[APEX] Status Check Exception:", err);
        return { persistent: false };
    }
}
