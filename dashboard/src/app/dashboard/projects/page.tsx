'use client';

import { UserButton } from '@/components/UserButton';
import Link from 'next/link';
import { useEffect, useState } from 'react';
import { useSession } from 'next-auth/react';
import { fetchProjects, createProject, Project, fetchStatus } from '@/lib/api';

export default function ProjectsHub() {
    const { data: session, status } = useSession();
    const [projects, setProjects] = useState<Project[]>([]);
    const [loading, setLoading] = useState(true);
    const [isCreating, setIsCreating] = useState(false);
    const [projectName, setProjectName] = useState('');
    const [isPersistent, setIsPersistent] = useState(false);

    useEffect(() => {
        // Check infrastructure status
        fetchStatus().then(status => setIsPersistent(status.persistent));

        if (status === 'loading') return;

        if (status === 'authenticated') {
            const userId = (session?.user as any).id || session?.user?.name || 'anonymous';
            fetchProjects(userId).then((data) => {
                setProjects(data);
                setLoading(false);
            });
        } else {
            // Unauthenticated or Error
            setLoading(false);
        }
    }, [session, status]);

    const handleCreate = async (e: React.FormEvent) => {
        e.preventDefault();
        const userId = (session?.user as any)?.id || session?.user?.name || 'anonymous';
        if (!projectName.trim()) return;

        setIsCreating(true);
        const newProject = await createProject(userId, projectName);
        if (newProject) {
            setProjects([newProject, ...projects]);
            setProjectName('');
            (window as any).showModal = false;
        }
        setIsCreating(false);
    };

    return (
        <div className="min-h-screen bg-[#080808] text-white selection:bg-[#FFB800] selection:text-black font-sans">
            {/* Top Header */}
            <header className="border-b border-[#222] bg-[#111] px-8 py-4 flex justify-between items-center">
                <div className="flex items-center gap-4">
                    <div className="w-2 h-2 bg-[#FFB800] animate-pulse rounded-full" />
                    <span className="font-black italic tracking-tighter uppercase text-xl">APEX <span className="text-[#FFB800]">Command</span></span>
                    
                    <div className={`px-3 py-1 text-[10px] font-mono border uppercase tracking-widest flex items-center gap-2 ${
                        isPersistent 
                        ? 'border-green-500/30 text-green-500 bg-green-500/5' 
                        : 'border-red-500/30 text-red-500 bg-red-500/5 animate-pulse'
                    }`}>
                        <div className={`w-1 h-1 rounded-full ${isPersistent ? 'bg-green-500' : 'bg-red-500'}`} />
                        {isPersistent ? 'Vault Online' : 'Infrastructure Offline (No Save)'}
                    </div>
                </div>
                <UserButton />
            </header>

            <main className="max-w-5xl mx-auto px-6 py-20">
                <div className="mb-16 border-l-2 border-[#FFB800] pl-6">
                    <h1 className="text-4xl md:text-5xl font-black italic tracking-tight uppercase">Your <span className="text-[#FFB800]">Workspaces</span></h1>
                    <p className="text-gray-400 font-mono text-sm max-w-2xl mt-4 leading-relaxed">
                        Projects isolate your crash data. Each project generates a unique Ingest Key. Drop this key into your Edge Agents to begin receiving telemetry.
                    </p>
                </div>

                <div className="grid lg:grid-cols-3 gap-8">
                    <div className="lg:col-span-2 space-y-8">
                        {loading ? (
                            <div className="font-mono text-xs text-gray-500 animate-pulse">Scanning tactical nodes...</div>
                        ) : projects && projects.length > 0 ? (
                            projects.map((p) => (
                                <div key={p.id}>
                                    <h2 className="text-[10px] font-mono text-gray-500 uppercase tracking-widest mb-3 flex items-center gap-2">
                                        <span className="w-1 h-1 bg-green-500 rounded-full animate-pulse" />
                                        Active Deployment
                                    </h2>
                                    <div className="border border-[#222] bg-[#111] p-8 hover:border-[#FFB800]/50 transition-all group relative overflow-hidden shadow-2xl">
                                        <div className="absolute top-0 right-0 w-32 h-32 hazard-pattern opacity-5 -mr-16 -mt-16 rotate-45 group-hover:opacity-10 transition-opacity" />

                                        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-8 relative z-10">
                                            <div>
                                                <h3 className="text-3xl font-black italic tracking-tighter uppercase text-white mb-2">{p.name}</h3>
                                                <span className="text-[10px] font-mono text-[#FFB800] bg-[#FFB800]/10 px-3 py-1 uppercase tracking-widest border border-[#FFB800]/20">Receiving Signal</span>
                                            </div>
                                            <Link href={`/dashboard/projects/${p.id}`} className="border border-[#FFB800] text-[#FFB800] px-8 py-3 text-xs font-black uppercase tracking-widest hover:bg-[#FFB800] hover:text-black transition-all shadow-[0_0_15px_rgba(255,184,0,0.1)] w-full sm:w-auto text-center">
                                                Open HUD
                                            </Link>
                                        </div>

                                        <div className="bg-[#080808] border border-[#222] p-4 relative z-10 group-hover:border-[#333] transition-colors">
                                            <div className="flex justify-between items-center mb-2">
                                                <span className="text-[9px] uppercase font-mono text-gray-500 flex items-center gap-2">
                                                    Ingest Key
                                                </span>
                                            </div>
                                            <code className="text-[10px] font-mono text-green-500/80 break-all select-all block p-2 bg-black border border-[#111]">
                                                {p.ingest_key}
                                            </code>
                                        </div>
                                    </div>
                                </div>
                            ))
                        ) : (
                            <div className="border border-dashed border-[#222] p-12 text-center text-gray-600 font-mono text-xs uppercase tracking-widest">
                                No active projects detected
                            </div>
                        )}

                        {/* Create New Project Section */}
                        <div className="mt-12">
                            <form onSubmit={handleCreate} className="space-y-4">
                                <div className="flex gap-2">
                                    <input
                                        type="text"
                                        placeholder="Project Name (e.g. My Production API)"
                                        value={projectName}
                                        onChange={(e) => setProjectName(e.target.value)}
                                        className="flex-1 bg-[#111] border border-[#222] px-4 py-3 text-sm font-mono focus:border-[#FFB800] outline-none transition-colors"
                                        required
                                    />
                                    <button
                                        type="submit"
                                        disabled={isCreating}
                                        className="bg-[#FFB800] text-black px-8 py-3 text-xs font-black uppercase tracking-widest hover:bg-[#FFD700] transition-all disabled:opacity-50"
                                    >
                                        {isCreating ? 'Initializing...' : 'Initialize'}
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>

                    {/* Right Sidebar */}
                    <div className="space-y-6">
                        <div className="border border-[#222] bg-[#111] p-6 relative overflow-hidden">
                            <div className="absolute top-0 right-0 w-1 h-full bg-[#FFB800]" />
                            <h3 className="text-sm font-black italic uppercase tracking-widest text-[#FFB800] mb-4">Tactical Brief</h3>
                            <p className="text-xs font-mono text-gray-400 leading-relaxed mb-4">
                                <strong>1. Create Project:</strong> Establish a designated drop zone for your telemetry.
                            </p>
                            <p className="text-xs font-mono text-gray-400 leading-relaxed mb-4">
                                <strong>2. Secure Key:</strong> The Ingest Key authorizes edge agents to transmit securely.
                            </p>
                            <p className="text-xs font-mono text-gray-400 leading-relaxed">
                                <strong>3. Deploy Agents:</strong> Integrate the Go or Python SDKs. Crashes instantly appear on your HUD.
                            </p>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    );
}
