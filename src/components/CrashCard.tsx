import { CrashReport } from '@/lib/api';

export function CrashCard({ report }: { report: CrashReport }) {
    const date = new Date(report.timestamp * 1000).toLocaleString();

    return (
        <div className="industrial-card rounded-lg overflow-hidden border-l-4 border-l-[#FF4D00]">
            {/* Hazard Header */}
            <div className="flex items-center justify-between px-5 py-2 bg-[#1a1a1a] border-b border-[#222]">
                <div className="flex items-center gap-2">
                    <div className="w-2 h-2 bg-[#FF4D00] animate-pulse" />
                    <span className="text-[10px] font-mono font-black text-[#FF4D00] uppercase tracking-tighter">
                        CRITICAL_FAILURE // {report.error_id.substring(0, 10)}
                    </span>
                </div>
                <span className="text-[9px] font-mono text-gray-500 uppercase">{date}</span>
            </div>

            <div className="p-6">
                <h3 className="text-3xl font-black text-white italic tracking-tighter mb-6 uppercase leading-none">
                    {report.message}
                </h3>

                {/* Telemetry Grid */}
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
                    <DataPoint label="OS_ENV" value={report.context?.os} color="#FFB800" />
                    <DataPoint label="ARCHITECTURE" value={report.context?.arch} color="#888" />
                    <DataPoint label="MEM_USED" value={`${((report.context?.free_memory || 0) / 1024 / 1024).toFixed(0)} MB`} color="#FFB800" />
                    <DataPoint label="PWR_LVL" value={`${(report.context?.battery_level || 0).toFixed(0)}%`} color={report.context?.battery_level < 20 ? '#FF4D00' : '#00FF41'} />
                </div>

                {/* Trace */}
                <div className="bg-black/60 p-5 rounded border border-[#222] font-mono text-[11px] leading-relaxed text-[#FFB800]/80 overflow-x-auto max-h-[300px]">
                    {report.stack_trace.split('\n').map((line, i) => (
                        <div key={i} className="flex gap-4 hover:bg-[#FFB800]/5 py-0.5">
                            <span className="text-[#333] w-6 text-right select-none">{i + 1}</span>
                            <span className={line.startsWith('\t') ? 'text-[#FF4D00]' : ''}>{line}</span>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
}

function DataPoint({ label, value, color }: { label: string, value: string, color: string }) {
    return (
        <div className="border-l border-[#333] pl-3">
            <span className="text-[8px] font-mono text-gray-600 block mb-1 tracking-widest">{label}</span>
            <span className="text-xs font-black uppercase tracking-tight" style={{ color }}>{value || '---'}</span>
        </div>
    )
}
