'use client';

import Link from 'next/link';

export default function Docs() {
    const milestones = [
        { date: "2026-03-06", title: "Project Inception", desc: "Started Apex Monitoring Engine for Go production environments." },
        { date: "2026-03-06", title: "Dockerization Phase", desc: "Implemented containerization for PostgreSQL and Receiver." },
        { date: "2026-03-06", title: "Reliability Layer", desc: "Added Retries, Backoff, and Health checks. Fixed DB Auth conflicts." },
        { date: "2026-03-06", title: "Dashboard V2 (Next.js)", desc: "Upgraded from Go Templates to high-fidelity React dashboard." }
    ];

    return (
        <div className="min-h-screen bg-[#080808] text-white">
            <div className="w-full h-1 hazard-pattern" />

            <main className="max-w-4xl mx-auto px-6 py-20">
                <header className="mb-20">
                    <Link href="/" className="text-[10px] font-black text-[#FFB800] hover:underline uppercase tracking-widest mb-8 block">
                        ← Return to Command Center
                    </Link>
                    <h1 className="text-6xl font-black italic tracking-tighter uppercase mb-4">
                        The <span className="text-[#FFB800]">Ledger</span>
                    </h1>
                    <p className="text-gray-500 font-mono text-sm max-w-2xl">
                        Technical documentation and development log for the APEX monitoring infrastructure.
                    </p>
                </header>

                <section className="space-y-16">
                    <div>
                        <h2 className="text-[11px] font-black text-[#FFB800] uppercase tracking-[0.3em] mb-8 border-b border-[#222] pb-2">Technical Overview</h2>
                        <div className="grid md:grid-cols-2 gap-12">
                            <div className="space-y-4">
                                <h3 className="text-xl font-bold italic">The Core Engine</h3>
                                <p className="text-gray-400 text-sm leading-relaxed">
                                    Built in Go 1.24, the receiver handles thousands of concurrent reports using a distributed worker pool and structured logging.
                                </p>
                            </div>
                            <div className="space-y-4">
                                <h3 className="text-xl font-bold italic">The Dashboard</h3>
                                <p className="text-gray-400 text-sm leading-relaxed">
                                    Next.js 15 powered frontend utilizing Tailwind CSS v4 and Framer-inspired glassmorphism for real-time failure forensics.
                                </p>
                            </div>
                        </div>
                    </div>

                    <div>
                        <h2 className="text-[11px] font-black text-[#FFB800] uppercase tracking-[0.3em] mb-8 border-b border-[#222] pb-2">Development Log</h2>
                        <div className="space-y-8">
                            {milestones.map((m, i) => (
                                <div key={i} className="flex gap-8 group">
                                    <span className="text-[10px] font-mono text-gray-700 w-24 shrink-0 pt-1 group-hover:text-[#FFB800] transition-colors">{m.date}</span>
                                    <div className="space-y-1">
                                        <h4 className="font-bold text-gray-200 group-hover:text-white">{m.title}</h4>
                                        <p className="text-xs text-gray-500">{m.desc}</p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                </section>
            </main>

            <footer className="py-20 text-center text-[9px] font-mono text-gray-800 tracking-[0.4em] uppercase">
                Apex Systems Engineering // Internal_Only_v1
            </footer>
        </div>
    );
}
