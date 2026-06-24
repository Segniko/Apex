import { CrashReport } from './api';

export interface Issue {
    fingerprint: string;
    latest: CrashReport;     // representative (most recent) report
    count: number;           // total occurrences in the current window
    firstSeen: number;       // unix seconds
    lastSeen: number;        // unix seconds
    resolved: boolean;       // resolved when every occurrence is resolved
    reports: CrashReport[];
}

// Normalize a stack trace so the same logical crash fingerprints identically:
// strip hex addresses, line/column numbers, and goroutine ids that vary per event.
function normalizeTrace(trace: string): string {
    return (trace || '')
        .replace(/0x[0-9a-fA-F]+/g, '<addr>')
        .replace(/:\d+(:\d+)?/g, ':<n>')
        .replace(/goroutine \d+/g, 'goroutine <n>')
        .replace(/\s+/g, ' ')
        .trim()
        .slice(0, 400);
}

function fingerprintOf(r: CrashReport): string {
    // Group by message + the first few normalized frames — matches how
    // Sentry-style tools collapse identical crashes into one issue.
    const head = normalizeTrace(r.stack_trace).split(' ').slice(0, 24).join(' ');
    return `${(r.message || '').trim().toLowerCase()}::${head}`;
}

// Collapse a flat list of crash events into deduplicated issues with counts.
export function groupReports(reports: CrashReport[]): Issue[] {
    const map = new Map<string, Issue>();

    for (const r of reports) {
        const fp = fingerprintOf(r);
        const existing = map.get(fp);
        if (!existing) {
            map.set(fp, {
                fingerprint: fp,
                latest: r,
                count: 1,
                firstSeen: r.timestamp,
                lastSeen: r.timestamp,
                resolved: r.resolved,
                reports: [r],
            });
        } else {
            existing.count += 1;
            existing.reports.push(r);
            existing.firstSeen = Math.min(existing.firstSeen, r.timestamp);
            existing.resolved = existing.resolved && r.resolved;
            if (r.timestamp > existing.lastSeen) {
                existing.lastSeen = r.timestamp;
                existing.latest = r;
            }
        }
    }

    return Array.from(map.values()).sort((a, b) => b.lastSeen - a.lastSeen);
}

// Bucket event timestamps into N equal slots over the last `windowMs`,
// returning counts for a sparkline.
export function timeBuckets(reports: CrashReport[], buckets = 24, windowMs = 24 * 60 * 60 * 1000): number[] {
    const now = Date.now();
    const start = now - windowMs;
    const slot = windowMs / buckets;
    const out = new Array(buckets).fill(0);
    for (const r of reports) {
        const t = r.timestamp * 1000;
        if (t < start || t > now) continue;
        const idx = Math.min(buckets - 1, Math.floor((t - start) / slot));
        out[idx] += 1;
    }
    return out;
}
