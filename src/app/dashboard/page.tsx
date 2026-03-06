'use client';

import { CrashCard } from '@/components/CrashCard';
import { CrashReport, fetchReports } from '@/lib/api';
import Link from 'next/link';
import { useEffect, useState } from 'react';

export default function Dashboard() {
    const [reports, setReports] = useState<CrashReport[]>([]);
    const [loading, setLoading] = useState(true);

    const loadData = async () => {
        const data = await fetchReports();
        setReports(data);
        setLoading(false);
    };

    useEffect(() => { loadData(); const i = setInterval(loadData, 5000); return () => clearInterval(i); }, []);

    return (
        <div className="min-h-screen bg-[#080808] text-white overflow-x-hidden">
            {/* Industrial Header Bar */}
            <div className="w-full h-1 hazard-pattern mb-12 shadow-[0_0_15px_rgba(255,184,0,0.4)]" />

            <main className="max-w-6xl mx-auto px-6">
                <header className="flex flex-col md:flex-row justify-between items-start md:items-end gap-8 mb-16 border-b border-[#222] pb-12">
                    <div className="space-y-3">
                        <div className="flex items-center gap-3">
                            <div className="bg-[#FFB800] text-black text-[8px] font-black px-2 py-0.5 rounded uppercase tracking-widest">Live_Signal</div>
                            <div className="h-[1px] w-8 bg-[#333]" />
                            <span className="text-[10px] font-mono text-gray-600 uppercase">Apex.Monitor.Engine</span>
                        </div>
                        <h1 className="text-6xl font-black italic tracking-tighter text-white uppercase sm:text-7xl">
                            Command <span className="text-[#FFB800]">Center</span>
                        </h1>
                        <div className="flex gap-6 pt-4">
                            <Link href="/" className="text-[10px] font-black text-[#FFB800] hover:underline uppercase tracking-widest flex items-center gap-2">
                                [ Return to Front Door ]
                            </Link>
                            <Link href="/docs" className="text-[10px] font-black text-[#FFB800] hover:underline uppercase tracking-widest flex items-center gap-2">
                                [ View Documentation Log ]
                            </Link>
                        </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <StatBox label="Total Failures" value={reports.length.toString()} color="#FFB800" />
                        <StatBox label="Active Nodes" value="01 / 12" color="#888" />
                    </div>
                </header>

                {/* Live Feed Container */}
                <div className="grid gap-10 lg:grid-cols-[1fr_280px]">
                    <div className="space-y-8">
                        <div className="flex items-center gap-4 mb-4">
                            <div className="w-2 h-2 rounded-full bg-[#00FF41] shadow-[0_0_10px_#00FF41]" />
                            <span className="text-[10px] font-black text-gray-500 tracking-[0.4em] uppercase">Security Forensics Pipe</span>
                            <div className="h-[1px] flex-1 bg-[#222]" />
                        </div>

                        {loading ? (
                            <div className="space-y-6">
                                {[1, 2].map(i => <div key={i} className="h-64 bg-[#111] animate-pulse rounded border border-[#222]" />)}
                            </div>
                        ) : reports.length > 0 ? (
                            <div className="grid gap-8">
                                {reports.map(r => <CrashCard key={r.error_id} report={r} />)}
                            </div>
                        ) : (
                            <div className="py-32 text-center border-2 border-dashed border-[#222] rounded-xl opacity-40">
                                <h2 className="text-xl font-black text-white italic tracking-widest uppercase">No Breaches Detected</h2>
                            </div>
                        )}
                    </div>

                    {/* Sidebar */}
                    <aside className="space-y-8">
                        <div className="bg-[#111] p-6 rounded border border-[#222] relative overflow-hidden group">
                            <div className="absolute top-0 right-0 w-16 h-16 hazard-pattern opacity-10 -mr-8 -mt-8 rotate-45" />
                            <h4 className="text-[10px] font-black text-[#FFB800] mb-6 tracking-widest uppercase border-b border-[#333] pb-2">Facility Status</h4>
                            <div className="space-y-4">
                                <StatusItem label="Database" status="CONNECTED" ok />
                                <StatusItem label="Receiver" status="STABLE" ok />
                                <StatusItem label="Auth_Node" status="VERIFIED" ok />
                                <StatusItem label="Storage" status="CockroachDB" />
                            </div>
                        </div>

                        <div className="p-6 border border-[#222] bg-[#FFB800]/5 text-[11px] text-[#FFB800]/60 leading-relaxed font-mono italic">
                            "Architecture is not just about the build; it's about the recovery." // APEX_LOG_001
                        </div>
                    </aside>
                </div>
            </main>

            <footer className="py-20 text-center text-[9px] font-mono text-gray-700 tracking-[0.6em] uppercase">
                Apex Tactical Operations // 2026 // No_Signal_Loss
            </footer>
        </div>
    );
}

function StatBox({ label, value, color }: { label: string, value: string, color: string }) {
    return (
        <div className="bg-[#111] p-6 rounded border border-[#222] min-w-[140px]">
            <span className="text-[8px] font-black text-gray-600 uppercase tracking-widest block mb-1">{label}</span>
            <span className="text-4xl font-black italic tracking-tighter" style={{ color }}>{value}</span>
        </div>
    );
}

function StatusItem({ label, status, ok }: { label: string, status: string, ok?: boolean }) {
    return (
        <div className="flex justify-between items-center text-[10px] font-mono">
            <span className="text-gray-500 tracking-tight">{label}</span>
            <span className={ok ? 'text-[#00FF41]' : 'text-gray-400'}>{status}</span>
        </div>
    )
}
