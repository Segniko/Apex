'use client';

import { UserButton } from '@/components/UserButton';
import { OnboardingGuide } from '@/components/OnboardingGuide';
import { CrashFeed } from '@/components/CrashFeed';
import { CrashReport, fetchReports, fetchProjects, fetchStatus, Project } from '@/lib/api';
import Link from 'next/link';
import { useEffect, useState } from 'react';
import { useSession } from 'next-auth/react';

export default function Dashboard() {
    const { data: session } = useSession();
    const [reports, setReports] = useState<CrashReport[]>([]);
    const [projects, setProjects] = useState<Project[]>([]);
    const [loading, setLoading] = useState(true);
    const [online, setOnline] = useState<boolean | null>(null);
    const [persistent, setPersistent] = useState(false);

    const loadData = async () => {
        const [data, status] = await Promise.all([fetchReports(), fetchStatus()]);
        setReports(data);
        setOnline(true);
        setPersistent(status.persistent);

        if (session?.user?.email && projects.length === 0) {
            const userProjects = await fetchProjects(session.user.email);
            setProjects(userProjects);
        }
        setLoading(false);
    };

    useEffect(() => { loadData(); const i = setInterval(loadData, 5000); return () => clearInterval(i); }, []);

    return (
        <div className="min-h-screen bg-[#080808] text-white overflow-x-hidden">
            <div className="w-full h-1 hazard-pattern mb-12 shadow-[0_0_15px_rgba(255,184,0,0.4)]" />

            <main className="max-w-6xl mx-auto px-6">
                <header className="flex flex-col md:flex-row justify-between items-start md:items-end gap-8 mb-12 border-b border-[#222] pb-10">
                    <div className="space-y-3">
                        <div className="flex items-center gap-3">
                            <div className="bg-[#FFB800] text-black text-[10px] font-black px-2 py-0.5 rounded uppercase tracking-widest">Live Signal</div>
                            <div className="h-[1px] w-8 bg-[#333]" />
                            <span className="text-[11px] font-mono text-gray-500 uppercase">apex.monitor.engine</span>
                        </div>
                        <h1 className="text-5xl md:text-7xl font-black italic tracking-tighter text-white uppercase">
                            Command <span className="text-[#FFB800]">Center</span>
                        </h1>
                        <div className="flex flex-wrap gap-6 pt-4">
                            <Link href="/" className="text-[11px] font-bold text-[#FFB800] hover:underline uppercase tracking-widest">Home</Link>
                            <Link href="/dashboard/projects" className="text-[11px] font-bold text-[#FFB800] hover:underline uppercase tracking-widest">Projects</Link>
                            <Link href="/docs" className="text-[11px] font-bold text-[#FFB800] hover:underline uppercase tracking-widest">Docs</Link>
                        </div>
                    </div>
                    <UserButton />
                </header>

                <div className="grid gap-10 lg:grid-cols-[1fr_280px]">
                    <div>
                        {!loading && reports.length === 0 && projects.length > 0 ? (
                            <OnboardingGuide project={projects[0]} />
                        ) : (
                            <CrashFeed reports={reports} loading={loading} label="Security Forensics Pipe" />
                        )}
                    </div>

                    {/* Sidebar — real infrastructure status */}
                    <aside className="space-y-8">
                        <div className="bg-[#111] p-6 rounded border border-[#222] relative overflow-hidden">
                            <div className="absolute top-0 right-0 w-16 h-16 hazard-pattern opacity-10 -mr-8 -mt-8 rotate-45" />
                            <h4 className="text-[11px] font-black text-[#FFB800] mb-6 tracking-widest uppercase border-b border-[#333] pb-2">Facility Status</h4>
                            <div className="space-y-4">
                                <StatusItem label="Receiver" status={online === null ? '…' : online ? 'ONLINE' : 'UNREACHABLE'} ok={!!online} />
                                <StatusItem label="Storage" status={persistent ? 'PERSISTENT' : 'MEMORY'} ok={persistent} />
                                <StatusItem label="Auth" status={session ? 'VERIFIED' : 'GUEST'} ok={!!session} />
                            </div>
                        </div>

                        <div className="p-6 border border-[#222] bg-[#FFB800]/5 text-sm text-[#FFB800]/70 leading-relaxed font-mono italic">
                            &quot;Architecture is not just about the build; it&apos;s about the recovery.&quot;
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
