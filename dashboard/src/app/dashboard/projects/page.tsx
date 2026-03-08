import Link from 'next/link';

export default function ProjectsHub() {
    return (
        <div className="min-h-screen bg-[#080808] text-white selection:bg-[#FFB800] selection:text-black font-sans">
            {/* Top Header */}
            <header className="border-b border-[#222] bg-[#111] px-8 py-4 flex justify-between items-center">
                <div className="flex items-center gap-4">
                    <div className="w-2 h-2 bg-[#FFB800] animate-pulse rounded-full" />
                    <span className="font-black italic tracking-tighter uppercase text-xl">APEX <span className="text-[#FFB800]">Command</span></span>
                </div>
                <div className="flex items-center gap-6">
                    <span className="text-xs font-mono text-gray-500 uppercase tracking-widest hidden md:inline">ID: u_8291x</span>
                    <div className="w-8 h-8 bg-[#222] rounded-full overflow-hidden flex items-center justify-center border border-[#333] hover:border-[#FFB800] transition-colors cursor-pointer">
                        <span className="text-[10px] font-mono text-gray-400">GH</span>
                    </div>
                </div>
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
                        {/* Existing Project Card */}
                        <div>
                            <h2 className="text-[10px] font-mono text-gray-500 uppercase tracking-widest mb-3 flex items-center gap-2">
                                <span className="w-1 h-1 bg-green-500 rounded-full animate-pulse" />
                                Active Deployments
                            </h2>
                            <div className="border border-[#222] bg-[#111] p-8 hover:border-[#FFB800]/50 transition-all group relative overflow-hidden shadow-2xl">
                                <div className="absolute top-0 right-0 w-32 h-32 hazard-pattern opacity-5 -mr-16 -mt-16 rotate-45 group-hover:opacity-10 transition-opacity" />

                                <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-8 relative z-10">
                                    <div>
                                        <h3 className="text-3xl font-black italic tracking-tighter uppercase text-white mb-2">Production API</h3>
                                        <span className="text-[10px] font-mono text-[#FFB800] bg-[#FFB800]/10 px-3 py-1 uppercase tracking-widest border border-[#FFB800]/20">Receiving Signal</span>
                                    </div>
                                    <Link href="/dashboard" className="border border-[#FFB800] text-[#FFB800] px-8 py-3 text-xs font-black uppercase tracking-widest hover:bg-[#FFB800] hover:text-black transition-all shadow-[0_0_15px_rgba(255,184,0,0.1)] w-full sm:w-auto text-center">
                                        Open HUD
                                    </Link>
                                </div>

                                <div className="bg-[#080808] border border-[#222] p-5 relative z-10 group-hover:border-[#333] transition-colors">
                                    <div className="flex justify-between items-center mb-3">
                                        <span className="text-[10px] uppercase font-mono text-gray-500 flex items-center gap-2">
                                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" /></svg>
                                            Ingest Key (Keep Secret)
                                        </span>
                                        <button className="text-[10px] uppercase font-mono text-[#FFB800] hover:text-white transition-colors">Copy to Clipboard</button>
                                    </div>
                                    <code className="text-xs md:text-sm font-mono text-green-500/80 break-all select-all block p-3 bg-black border border-[#111] rounded shadow-inner">
                                        apx_prod_99x82jf928jfh2910cnf82nf
                                    </code>
                                </div>
                            </div>
                        </div>

                        {/* Create New Project */}
                        <div>
                            <button className="w-full border border-dashed border-[#333] bg-[#111]/50 hover:border-[#FFB800] hover:bg-[#FFB800]/5 p-12 flex flex-col items-center justify-center gap-4 transition-all group">
                                <div className="w-12 h-12 rounded-full border border-[#FFB800] flex items-center justify-center text-[#FFB800] text-2xl font-light group-hover:bg-[#FFB800] group-hover:text-black transition-all shadow-[0_0_15px_rgba(255,184,0,0.1)] group-hover:shadow-[0_0_25px_rgba(255,184,0,0.4)]">
                                    +
                                </div>
                                <span className="font-black italic uppercase tracking-tighter text-gray-400 group-hover:text-[#FFB800] transition-colors text-lg">Initialize New Project</span>
                            </button>
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

                        <div className="border border-[#222] bg-black p-6 flex items-center gap-4 hover:border-[#FFB800]/30 transition-colors cursor-pointer">
                            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#FFB800" strokeWidth="2"><path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" /></svg>
                            <div>
                                <h4 className="text-xs font-black uppercase tracking-widest text-white">Billing & Usage</h4>
                                <p className="text-[10px] font-mono text-gray-500 uppercase mt-1">100% Free Forever</p>
                            </div>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    );
}
