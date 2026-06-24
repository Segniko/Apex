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
    resolved: boolean;
    project_id: string;
}

export interface Project {
    id: string;
    user_id: string;
    name: string;
    ingest_key: string;
    created_at: string;
}

export interface Issue {
    fingerprint: string;
    project_id: string;
    message: string;
    stack_trace: string;
    ai_insight: string;
    count: number;
    first_seen: number;
    last_seen: number;
    resolved: boolean;
    error_id: string;
}

const API_BASE = (process.env.NEXT_PUBLIC_API_URL || "http://localhost:8081") + "/api";

// When NEXT_PUBLIC_USE_BFF=1, reads go through same-origin Next route handlers
// that enforce the session and inject the server-only receiver secret. Otherwise
// they hit the receiver directly (the original/open behaviour).
const USE_BFF = process.env.NEXT_PUBLIC_USE_BFF === "1";
const READ_BASE = USE_BFF ? "/api/bff" : API_BASE;

// Optional shared key for mutating endpoints. Sent only when the deployment
// configures APEX_DASHBOARD_SECRET / NEXT_PUBLIC_DASHBOARD_KEY; otherwise the
// header is omitted and the open (demo) behaviour is preserved.
function mutationHeaders(): Record<string, string> {
    const headers: Record<string, string> = { "Content-Type": "application/json" };
    const key = process.env.NEXT_PUBLIC_DASHBOARD_KEY;
    if (key) headers["X-Apex-Dashboard-Key"] = key;
    return headers;
}

export async function fetchReports(projectId?: string): Promise<CrashReport[]> {
    try {
        const url = projectId
            ? `${READ_BASE}/reports?project_id=${projectId}`
            : `${READ_BASE}/reports`;
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
        // Through the BFF the user id comes from the session, not the client.
        const url = USE_BFF
            ? `${READ_BASE}/projects`
            : `${API_BASE}/projects?user_id=${userID}`;
        const res = await fetch(url, { cache: 'no-store' });
        if (!res.ok) throw new Error("Failed to fetch projects");
        const data = await res.json();
        return data || [];
    } catch (err) {
        console.error("API Error:", err);
        return [];
    }
}

export async function fetchIssues(projectId?: string, limit = 50, offset = 0): Promise<Issue[]> {
    try {
        const pid = projectId ? encodeURIComponent(projectId) : "";
        const res = await fetch(`${READ_BASE}/issues?project_id=${pid}&limit=${limit}&offset=${offset}`, { cache: 'no-store' });
        if (!res.ok) throw new Error("Failed to fetch issues");
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
        const res = await fetch(url, {
            method: 'POST',
            headers: mutationHeaders(),
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

export async function resolveReport(reportId: string, resolved: boolean): Promise<boolean> {
    try {
        const res = await fetch(`${API_BASE}/reports/resolve?id=${reportId}`, {
            method: 'PATCH',
            headers: mutationHeaders(),
            body: JSON.stringify({ resolved }),
        });
        return res.ok;
    } catch (err) {
        console.error("Failed to resolve report:", err);
        return false;
    }
}

// Streams an AI forensic answer for a specific report over SSE, invoking
// onChunk with the accumulated text as it arrives. Resolves when the stream
// ends ([DONE]) or the body closes.
export async function streamSolution(
    reportId: string,
    message: string,
    onChunk: (text: string) => void,
): Promise<void> {
    const res = await fetch(`${API_BASE}/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message, report_id: reportId }),
    });
    if (!res.ok || !res.body) throw new Error(`Stream failed: ${res.status}`);

    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let acc = '';

    while (true) {
        const { value, done } = await reader.read();
        if (done) break;
        const chunk = decoder.decode(value);
        for (const line of chunk.split('\n')) {
            if (!line.startsWith('data: ')) continue;
            const content = line.slice(6);
            if (content.trim() === '[DONE]') return;
            if (content === '') continue;
            acc += content;
            onChunk(acc);
        }
    }
}

export async function deleteProject(projectId: string): Promise<boolean> {
    try {
        const res = await fetch(`${API_BASE}/projects/delete?id=${projectId}`, {
            method: 'DELETE',
            headers: mutationHeaders(),
        });
        return res.ok;
    } catch (err) {
        console.error("Failed to delete project:", err);
        return false;
    }
}
