'use client';

import { CrashReport } from '@/lib/api';
import { groupReports, timeBuckets } from '@/lib/group';
import { CrashCard } from '@/components/CrashCard';
import { Sparkline } from '@/components/Sparkline';
import { useEffect, useMemo, useRef, useState } from 'react';

type StatusFilter = 'all' | 'unresolved' | 'resolved';
type View = 'issues' | 'events';

export function CrashFeed({
    reports,
    loading,
    emptySlot,
    label = 'Forensics Feed',
}: {
    reports: CrashReport[];
    loading: boolean;
    emptySlot?: React.ReactNode;
    label?: string;
}) {
    const [query, setQuery] = useState('');
    const [status, setStatus] = useState<StatusFilter>('all');
    const [view, setView] = useState<View>('issues');

    // Track which event ids we've already rendered so freshly-arrived crashes
    // can play the entrance animation (the "money moment" for a live HUD).
    const seenRef = useRef<Set<string> | null>(null);
    const [newIds, setNewIds] = useState<Set<string>>(new Set());

    useEffect(() => {
        const current = new Set(reports.map(r => r.error_id));
        if (seenRef.current === null) {
            seenRef.current = current; // first load: nothing is "new"
            return;
        }
        const fresh = new Set<string>();
        for (const id of current) if (!seenRef.current.has(id)) fresh.add(id);
        if (fresh.size) setNewIds(fresh);
        seenRef.current = current;
    }, [reports]);

    const filtered = useMemo(() => {
        const q = query.trim().toLowerCase();
        return reports.filter(r => {
            if (status === 'unresolved' && r.resolved) return false;
            if (status === 'resolved' && !r.resolved) return false;
            if (!q) return true;
            return (
                r.message?.toLowerCase().includes(q) ||
                r.stack_trace?.toLowerCase().includes(q) ||
                r.context?.os?.toLowerCase().includes(q) ||
                r.context?.arch?.toLowerCase().includes(q)
            );
        });
    }, [reports, query, status]);

    const issues = useMemo(() => groupReports(filtered), [filtered]);
    const buckets = useMemo(() => timeBuckets(reports), [reports]);

    const resolvedCount = reports.filter(r => r.resolved).length;
    const totalIssues = useMemo(() => groupReports(reports).length, [reports]);

    return (
        <div className="space-y-6">
            {/* Stat strip */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                <Stat label="Events" value={reports.length.toString()} />
                <Stat label="Issues" value={totalIssues.toString()} />
                <Stat label="Resolved" value={resolvedCount.toString()} color="#00FF41" />
                <div className="bg-[#111] border border-[#222] rounded p-3">
                    <span className="text-[10px] font-mono text-gray-500 uppercase tracking-widest block mb-1">Last 24h</span>
                    <Sparkline data={buckets} height={30} />
                </div>
            </div>

            {/* Controls */}
            <div className="flex flex-col sm:flex-row gap-3 sm:items-center">
                <div className="flex items-center gap-3 flex-1">
                    <div className="w-2 h-2 rounded-full bg-[#00FF41] shadow-[0_0_10px_#00FF41]" />
                    <span className="text-[11px] font-bold text-gray-400 tracking-[0.3em] uppercase whitespace-nowrap">{label}</span>
                </div>
                <input
                    value={query}
                    onChange={e => setQuery(e.target.value)}
                    placeholder="Search message, trace, OS…"
                    className="flex-1 sm:max-w-xs bg-[#111] border border-[#222] px-3 py-2 text-sm font-mono text-white focus:border-[#FFB800] outline-none transition-colors placeholder:text-gray-600 rounded"
                />
                <div className="flex border border-[#222] rounded overflow-hidden">
                    {(['all', 'unresolved', 'resolved'] as StatusFilter[]).map(s => (
                        <button
                            key={s}
                            onClick={() => setStatus(s)}
                            className={`px-3 py-2 text-[11px] font-bold uppercase tracking-widest transition-colors ${status === s ? 'bg-[#FFB800] text-black' : 'text-gray-400 hover:text-white'}`}
                        >
                            {s}
                        </button>
                    ))}
                </div>
                <div className="flex border border-[#222] rounded overflow-hidden">
                    {(['issues', 'events'] as View[]).map(v => (
                        <button
                            key={v}
                            onClick={() => setView(v)}
                            className={`px-3 py-2 text-[11px] font-bold uppercase tracking-widest transition-colors ${view === v ? 'bg-[#FFB800] text-black' : 'text-gray-400 hover:text-white'}`}
                        >
                            {v}
                        </button>
                    ))}
                </div>
            </div>

            {/* List */}
            {loading ? (
                <div className="space-y-6">
                    {[1, 2].map(i => <div key={i} className="h-56 bg-[#111] animate-pulse rounded border border-[#222]" />)}
                </div>
            ) : reports.length === 0 ? (
                emptySlot ?? <EmptyState />
            ) : filtered.length === 0 ? (
                <div className="py-20 text-center border-2 border-dashed border-[#222] rounded-xl">
                    <p className="text-sm font-mono text-gray-500">No events match your filters.</p>
                </div>
            ) : view === 'issues' ? (
                <div className="grid gap-8">
                    {issues.map(issue => (
                        <CrashCard
                            key={issue.fingerprint}
                            report={issue.latest}
                            count={issue.count}
                            isNew={newIds.has(issue.latest.error_id)}
                        />
                    ))}
                </div>
            ) : (
                <div className="grid gap-8">
                    {filtered.map(r => (
                        <CrashCard key={r.error_id} report={r} isNew={newIds.has(r.error_id)} />
                    ))}
                </div>
            )}
        </div>
    );
}

function Stat({ label, value, color = '#FFB800' }: { label: string; value: string; color?: string }) {
    return (
        <div className="bg-[#111] border border-[#222] rounded p-3">
            <span className="text-[10px] font-mono text-gray-500 uppercase tracking-widest block mb-1">{label}</span>
            <span className="text-3xl font-black italic tracking-tighter" style={{ color }}>{value}</span>
        </div>
    );
}

function EmptyState() {
    return (
        <div className="py-24 text-center border-2 border-dashed border-[#222] rounded-xl">
            <h2 className="text-lg font-black text-white tracking-widest uppercase">No signal detected</h2>
            <p className="text-sm font-mono text-gray-500 mt-2">Your project is connected and clear.</p>
        </div>
    );
}
