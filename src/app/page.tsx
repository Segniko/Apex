'use client';

import Link from 'next/link';

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-[#080808] text-white selection:bg-[#FFB800] selection:text-black">
      {/* Top Hazard Bar */}
      <div className="w-full h-1 hazard-pattern sticky top-0 z-[200] shadow-[0_0_20px_rgba(255,184,0,0.4)]" />

      {/* Hero Section */}
      <section className="relative h-screen flex flex-col justify-center items-center px-6 overflow-hidden">
        {/* Minimalist Tactical Background */}
        <div className="absolute inset-0 z-0">
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(255,184,0,0.05)_0%,transparent_70%)]" />
          <div className="absolute inset-0 opacity-[0.03]" style={{ backgroundImage: 'linear-gradient(#FFB800 1px, transparent 1px), linear-gradient(90deg, #FFB800 1px, transparent 1px)', backgroundSize: '40px 40px' }} />
          <div className="absolute inset-0 bg-gradient-to-b from-transparent via-[#080808]/50 to-[#080808]" />
          {/* Scanline Effect */}
          <div className="absolute inset-0 pointer-events-none opacity-[0.02]" style={{ backgroundImage: 'linear-gradient(rgba(255,184,0,0.5) 50%, transparent 50%)', backgroundSize: '100% 4px' }} />
        </div>

        <div className="z-10 text-center space-y-8 max-w-5xl">
          <div className="inline-flex items-center gap-2 border border-[#FFB800]/30 px-3 py-1 rounded bg-[#FFB800]/5 backdrop-blur-sm">
            <div className="w-2 h-2 rounded-full bg-[#FFB800] animate-pulse" />
            <span className="text-[10px] font-black tracking-[0.4em] text-[#FFB800] uppercase">v2026.Deployment_Active</span>
          </div>

          <h1 className="text-7xl md:text-9xl font-black italic tracking-tighter uppercase leading-[0.8] drop-shadow-2xl">
            ARCHITECTURE OF <br />
            <span className="text-[#FFB800]">RECOVERY.</span>
          </h1>

          <p className="text-gray-400 font-mono text-sm md:text-lg max-w-2xl mx-auto leading-relaxed">
            Industrial grade observability, 100% Free and Open Source. Because high-performance recovery shouldn't have a price tag.
          </p>

          <div className="flex flex-col md:flex-row gap-6 justify-center pt-8">
            <Link href="/dashboard" className="bg-[#FFB800] text-black px-10 py-4 font-black uppercase tracking-tighter hover:bg-white transition-all transform hover:scale-105 active:scale-95 shadow-[0_0_30px_rgba(255,184,0,0.3)]">
              Launch Dashboard
            </Link>
            <Link href="https://github.com/Segniko/Apex" target='_blank' className="border border-[#222] text-gray-400 px-10 py-4 font-black uppercase tracking-tighter hover:border-[#FFB800] hover:text-white transition-all backdrop-blur-md">
              Fork on GitHub
            </Link>
          </div>
        </div>

        <div className="absolute bottom-10 left-1/2 -translate-x-1/2 animate-bounce opacity-20">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M7 13l5 5 5-5M7 6l5 5 5-5" /></svg>
        </div>
      </section>

      {/* Features Grid */}
      <section className="max-w-7xl mx-auto px-6 py-32 space-y-32">
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
          <FeatureCard
            title="Uncrashable Ingest"
            desc="Redis-buffered streams decouple signal from storage. Unlimited throughput, zero cost."
            tag="OSS_Core"
          />
          <FeatureCard
            title="Global Persistence"
            desc="Deploy CockroachDB clusters globally. High-availability persistence for the people."
            tag="Scale_Ready"
          />
          <FeatureCard
            title="AI Forensics HUD"
            desc="Automated root-cause analysis is standard. Every developer gets a forensic concierge."
            tag="Intelligence"
          />
          <FeatureCard
            title="Tactical Edge"
            desc="Open agents for Go and Python. Lightweight, compressed, and strictly zero-license."
            tag="Multi_Stack"
          />
        </div>

        {/* Detailed Explanation */}
        <div className="grid md:grid-cols-2 gap-20 items-center">
          <div className="space-y-8">
            <h2 className="text-4xl font-black italic uppercase tracking-tighter">Democratizing <span className="text-[#FFB800]">Observability.</span></h2>
            <div className="space-y-6 text-gray-400 font-mono text-sm leading-relaxed">
              <p>Observability shouldn't be a luxury. We believe every engineer—from students to enterprise architects—deserves access to high-fidelity failure forensics.</p>
              <p>Apex isn't just a tool; it's a statement. We've taken the technology used by trillion-dollar giants and made it accessible to everyone, for free.</p>
              <ul className="space-y-4">
                <li className="flex gap-3"><span className="text-[#FFB800]">●</span> No hidden fees, no "Pro" features</li>
                <li className="flex gap-3"><span className="text-[#FFB800]">●</span> Fully self-hostable on any cloud</li>
                <li className="flex gap-3"><span className="text-[#FFB800]">●</span> Driven by the developer community</li>
              </ul>
            </div>
          </div>
          <div className="relative aspect-video bg-[#111] rounded-lg border border-[#222] overflow-hidden group shadow-2xl">
            <div className="absolute inset-0 hazard-pattern opacity-5 group-hover:opacity-10 transition-opacity" />
            <div className="absolute inset-0 flex items-center justify-center">
              <span className="text-[10px] font-black text-[#FFB800] tracking-[0.5em] uppercase">Free_Forever_Engine</span>
            </div>
          </div>
        </div>

        {/* Call to Action */}
        <div className="text-center space-y-12 py-20 border border-[#222] bg-[#FFB800]/5 relative overflow-hidden">
          <div className="absolute top-0 right-0 w-32 h-32 hazard-pattern opacity-5 -mr-16 -mt-16 rotate-45" />
          <h2 className="text-5xl font-black italic uppercase tracking-tighter">Join the <span className="text-[#FFB800]">Movement</span></h2>
          <p className="max-w-xl mx-auto text-gray-400 font-mono text-xs uppercase tracking-widest">Help us build the most powerful free monitoring engine on the planet.</p>
          <div className="flex justify-center gap-6">
            <a href="https://github.com/Segniko/Apex" target='_blank'><button className="bg-[#FFB800] text-black px-12 py-4 font-black uppercase tracking-tighter hover:bg-white transition-all">
              GitHub Repository
            </button>
              <button className="border border-[#FFB800] text-[#FFB800] px-12 py-4 font-black uppercase tracking-tighter hover:bg-[#FFB800] hover:text-black transition-all">
                Star on GitHub
              </button></a>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-[#222] py-20 bg-[#0a0a0a]">
        <div className="max-w-7xl mx-auto px-6 flex flex-col md:flex-row justify-between items-center gap-8">
          <h2 className="text-2xl font-black italic uppercase tracking-tighter">APEX <span className="text-[#FFB800]">SYSTEMS</span></h2>
          <div className="flex gap-12 text-[10px] font-mono text-gray-500 uppercase tracking-widest">
            <Link href="/dashboard" className="hover:text-[#FFB800]">Dashboard</Link>
            <Link href="/docs" className="hover:text-[#FFB800]">Mission Log</Link>
            <span className="text-gray-800">Operational Log: 11-03-2026</span>
          </div>
        </div>
      </footer>
    </div>
  );
}

function FeatureCard({ title, desc, tag }: { title: string, desc: string, tag: string }) {
  return (
    <div className="group bg-[#111] p-8 border border-[#222] hover:border-[#FFB800]/50 transition-all space-y-4">
      <span className="text-[8px] font-black text-[#FFB800] tracking-widest uppercase opacity-40 group-hover:opacity-100 transition-opacity">[{tag}]</span>
      <h3 className="text-xl font-black italic uppercase tracking-tight">{title}</h3>
      <p className="text-gray-500 text-xs font-mono leading-relaxed">{desc}</p>
    </div>
  );
}

function PriceCard({ tier, price, features, featured }: { tier: string, price: string, features: string[], featured?: boolean }) {
  return (
    <div className={`relative p-10 border transition-all ${featured ? 'bg-[#FFB800]/5 border-[#FFB800] scale-105 z-10' : 'bg-[#111] border-[#222]'}`}>
      {featured && <div className="absolute top-0 left-1/2 -translate-x-1/2 bg-[#FFB800] text-black text-[9px] font-black uppercase px-4 py-1 tracking-widest -mt-3">Most Tactical</div>}
      <h4 className="text-3xl font-black italic uppercase tracking-tighter mb-2">{tier}</h4>
      <div className="flex items-baseline gap-1 mb-8">
        <span className="text-gray-500 text-sm">$</span>
        <span className="text-5xl font-black italic tracking-tighter">{price}</span>
        <span className="text-gray-500 text-xs font-mono">/MO</span>
      </div>
      <ul className="space-y-4 mb-10">
        {features.map((f, i) => (
          <li key={i} className="flex gap-3 text-[10px] font-mono text-gray-400 uppercase tracking-tight">
            <span className="text-[#FFB800]">✓</span> {f}
          </li>
        ))}
      </ul>
      <button className={`w-full py-4 font-black uppercase tracking-widest text-[10px] transition-all ${featured ? 'bg-[#FFB800] text-black hover:bg-white' : 'bg-white text-black hover:bg-[#FFB800]'}`}>
        Deploy Now
      </button>
    </div>
  );
}
