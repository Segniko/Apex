import { CrashReport, resolveReport } from '@/lib/api';
import { useState } from 'react';

export function CrashCard({ report, count = 1, isNew = false }: { report: CrashReport; count?: number; isNew?: boolean }) {
    const [isResolved, setIsResolved] = useState(report.resolved);
    const date = new Date(report.timestamp * 1000).toLocaleString();

    const total = report.context?.total_memory || 0;
    const free = report.context?.free_memory || 0;
    // Show actual used memory when we know the total; otherwise report free honestly.
    const memLabel = total > 0 ? 'MEM_USED' : 'MEM_FREE';
    const memValue = total > 0
        ? `${((total - free) / 1024 / 1024).toFixed(0)} MB`
        : `${(free / 1024 / 1024).toFixed(0)} MB`;

    return (
        <div className={`industrial-card rounded-lg overflow-hidden border-l-4 transition-all ${isNew ? 'animate-crash-in' : ''} ${isResolved ? 'border-l-[#00FF41] opacity-70' : 'border-l-[#FF4D00]'}`}>
            {/* Hazard Header */}
            <div className="flex items-center justify-between px-5 py-2.5 bg-[#1a1a1a] border-b border-[#222]">
                <div className="flex items-center gap-2">
                    <div className="w-2 h-2 bg-[#FFB800]" />
                    <span className="text-[11px] font-mono font-bold text-[#FFB800] uppercase tracking-tight">
                        {report.error_id.substring(0, 12)}
                    </span>
                    {count > 1 && (
                        <span className="ml-1 px-2 py-0.5 bg-[#FF4D00]/15 border border-[#FF4D00]/40 rounded text-[11px] text-[#FF4D00] font-mono font-bold" title="Occurrences of this error">
                            ×{count.toLocaleString()}
                        </span>
                    )}
                </div>
                <span className="text-[11px] font-mono text-gray-400">{date}</span>
            </div>

            <div className="p-6">
                <h3 className="text-2xl md:text-3xl font-black text-white tracking-tight mb-6 leading-tight">
                    {report.message}
                </h3>

                {/* Telemetry Grid */}
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
                    <DataPoint label="OS_ENV" value={report.context?.os} color="#FFB800" />
                    <DataPoint label="ARCHITECTURE" value={report.context?.arch} color="#bbb" />
                    <DataPoint label={memLabel} value={memValue} color="#FFB800" />
                    <DataPoint label="PWR_LVL" value={`${(report.context?.battery_level || 0).toFixed(0)}%`} color={report.context?.battery_level < 20 ? '#FF4D00' : '#00FF41'} />
                </div>

                {/* Trace */}
                <div className="bg-black/60 p-5 rounded border border-[#222] font-mono text-xs leading-relaxed trace-surface overflow-x-auto max-h-[220px] mb-8">
                    {report.stack_trace.split('\n').map((line, i) => (
                        <div key={i} className="flex gap-4 hover:bg-[#FFB800]/5 py-0.5">
                            <span className="text-[#444] w-6 text-right select-none shrink-0">{i + 1}</span>
                            <span className={line.startsWith('\t') ? 'text-[#FF6A33]' : 'trace-surface'}>{line || ' '}</span>
                        </div>
                    ))}
                </div>

                {/* Tactical Actions */}
                <div className="flex flex-wrap gap-4 mb-8">
                    <button
                        onClick={() => {
                            const event = new CustomEvent('apex-chat-context', {
                                detail: { errorId: report.error_id, message: report.message }
                            });
                            window.dispatchEvent(event);
                        }}
                        className="flex-1 bg-[#FFB800]/10 border border-[#FFB800]/30 text-[#FFB800] px-4 py-3 text-xs font-bold uppercase tracking-widest hover:bg-[#FFB800] hover:text-black transition-all flex items-center justify-center gap-2"
                    >
                        Ask AI about this error
                    </button>
                    <button
                        onClick={async () => {
                            const newStatus = !isResolved;
                            setIsResolved(newStatus);
                            await resolveReport(report.error_id, newStatus);
                        }}
                        className={`px-6 py-3 text-xs font-bold uppercase tracking-widest transition-all border ${
                            isResolved
                                ? 'bg-[#00FF41]/10 border-[#00FF41] text-[#00FF41]'
                                : 'bg-[#111] border-[#FF4D00]/30 text-[#FF4D00] hover:border-[#FF4D00]'
                        }`}
                    >
                        {isResolved ? '✓ Resolved' : 'Mark resolved'}
                    </button>
                </div>

                {/* AI Insight Panel */}
                {report.ai_insight && (
                    <div className="bg-[#FFB800]/10 border border-[#FFB800]/30 p-5 rounded relative overflow-hidden">
                        <div className="absolute top-0 right-0 px-2 py-0.5 bg-[#FFB800] text-black text-[10px] font-black uppercase tracking-widest">
                            AI Forensics
                        </div>
                        <div className="flex gap-4 items-start">
                            <div className="mt-1 w-2 h-2 rounded-full bg-[#FFB800] shadow-[0_0_10px_#FFB800] shrink-0" />
                            <div className="space-y-2">
                                <h4 className="text-[11px] font-black text-[#FFB800] uppercase tracking-widest">Root-cause analysis</h4>
                                <div className="text-sm text-gray-200 leading-relaxed font-mono whitespace-pre-wrap">
                                    {report.ai_insight}
                                </div>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}

function DataPoint({ label, value, color }: { label: string, value: string, color: string }) {
    return (
        <div className="border-l border-[#333] pl-3">
            <span className="text-[10px] font-mono text-gray-500 block mb-1 tracking-widest">{label}</span>
            <span className="text-sm font-bold uppercase tracking-tight" style={{ color }}>{value || '---'}</span>
        </div>
    )
}
