'use client';

import { CrashFeed } from '@/components/CrashFeed';
import { UserButton } from '@/components/UserButton';
import { CrashReport, fetchReports, fetchStatus } from '@/lib/api';
import Link from 'next/link';
import { useEffect, useState, use } from 'react';

export default function ProjectDashboard({ params }: { params: Promise<{ id: string }> }) {
    const { id } = use(params);
    const [reports, setReports] = useState<CrashReport[]>([]);
    const [loading, setLoading] = useState(true);
    const [isPersistent, setIsPersistent] = useState(false);

    const loadData = async () => {
        const [reportData, statusData] = await Promise.all([
            fetchReports(id),
            fetchStatus()
        ]);
        setReports(reportData);
        setIsPersistent(statusData.persistent);
        setLoading(false);
    };

    useEffect(() => {
        loadData();
        const i = setInterval(loadData, 5000);
        return () => clearInterval(i);
    }, [id]);

    return (
        <div className="min-h-screen bg-[#080808] text-white overflow-x-hidden">
            <div className="w-full h-1 hazard-pattern mb-12 shadow-[0_0_15px_rgba(255,184,0,0.4)]" />

            <main className="max-w-6xl mx-auto px-6">
                <header className="flex flex-col md:flex-row justify-between items-start md:items-end gap-8 mb-12 border-b border-[#222] pb-10">
                    <div className="space-y-3">
                        <div className="flex items-center gap-3">
                            <div className="bg-[#FFB800] text-black text-[10px] font-black px-2 py-0.5 rounded uppercase tracking-widest">Workspace HUD</div>
                            <div className="h-[1px] w-8 bg-[#333]" />
                            <div className={`px-2 py-0.5 text-[10px] font-mono border uppercase tracking-widest flex items-center gap-1 ${
                                isPersistent
                                ? 'border-green-500/30 text-green-500 bg-green-500/5'
                                : 'border-red-500/30 text-red-500 bg-red-500/5'
                            }`}>
                                <div className={`w-1 h-1 rounded-full ${isPersistent ? 'bg-green-500' : 'bg-red-500'}`} />
                                {isPersistent ? 'Persistent' : 'Memory Mode'}
                            </div>
                        </div>
                        <h1 className="text-5xl md:text-7xl font-black italic tracking-tighter text-white uppercase">
                            Mission <span className="text-[#FFB800]">Control</span>
                        </h1>
                        <p className="text-[11px] font-mono text-gray-500 uppercase tracking-widest">Project: {id}</p>
                        <div className="flex gap-6 pt-4">
                            <Link href="/dashboard/projects" className="text-[11px] font-bold text-[#FFB800] hover:underline uppercase tracking-widest">← Back to Hub</Link>
                        </div>
                    </div>
                    <UserButton />
                </header>

                <div className="grid gap-10 lg:grid-cols-[1fr_280px]">
                    <div>
                        <CrashFeed reports={reports} loading={loading} label="Tactical Forensics Pipe" />
                    </div>

                    <aside className="space-y-8">
                        <div className="bg-[#111] p-6 rounded border border-[#222] relative overflow-hidden">
                            <div className="absolute top-0 right-0 w-16 h-16 hazard-pattern opacity-10 -mr-8 -mt-8 rotate-45" />
                            <h4 className="text-[11px] font-black text-[#FFB800] mb-6 tracking-widest uppercase border-b border-[#333] pb-2">Operational Area</h4>
                            <div className="space-y-4">
                                <StatusItem label="Isolation" status="ACTIVE" ok />
                                <StatusItem label="Storage" status={isPersistent ? 'PERSISTENT' : 'MEMORY'} ok={isPersistent} />
                                <StatusItem label="Project" status={id.substring(0, 8)} />
                            </div>
                        </div>

                        <div className="p-6 border border-[#222] bg-[#FFB800]/5 text-sm text-[#FFB800]/70 leading-relaxed font-mono italic">
                            &quot;Data isolation is the first step of tactical security.&quot;
                        </div>
                    </aside>
                </div>
            </main>

            <footer className="py-20 text-center text-[10px] font-mono text-gray-700 tracking-[0.5em] uppercase">
                Apex Tactical Operations · {new Date().getFullYear()}
            </footer>
        </div>
    );
}

function StatusItem({ label, status, ok }: { label: string, status: string, ok?: boolean }) {
    return (
        <div className="flex justify-between items-center text-[11px] font-mono">
            <span className="text-gray-400 tracking-tight">{label}</span>
            <span className={ok ? 'text-[#00FF41]' : 'text-gray-400'}>{status}</span>
        </div>
    );
}
