'use client';

import Link from 'next/link';
import Image from 'next/image';
import { useSession } from 'next-auth/react';
import { useEffect, useRef, useState } from 'react';

function Reveal({ children, className = '', delay = 0 }: { children: React.ReactNode; className?: string; delay?: number }) {
  const ref = useRef<HTMLDivElement>(null);
  const [shown, setShown] = useState(false);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const ob = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setShown(true);
          ob.disconnect();
        }
      },
      { threshold: 0.15 }
    );
    ob.observe(el);
    return () => ob.disconnect();
  }, []);

  return (
    <div ref={ref} className={className} style={shown ? { animationDelay: `${delay}ms` } : undefined}>
      <div className={shown ? 'reveal-up' : 'opacity-0'}>{children}</div>
    </div>
  );
}

export default function LandingPage() {
  const { data: session } = useSession();
  const dashboardLink = session ? "/dashboard/projects" : "/auth/login";

  return (
    <div className="min-h-screen bg-[#080808] text-white selection:bg-[#FFB800] selection:text-black">
      {/* Top Hazard Bar */}
      <div className="w-full h-1 hazard-pattern sticky top-0 z-[200] shadow-[0_0_20px_rgba(255,184,0,0.4)]" />

      {/* Global Navbar */}
      <nav className="absolute top-0 w-full z-50 flex justify-between items-center px-6 md:px-12 py-6 md:py-8">
        <div className="flex items-center gap-3">
          <div className="w-3 h-3 bg-[#FFB800] animate-pulse" />
          <Image
            src="/apex-logo.png"
            alt="Apex"
            width={584}
            height={276}
            priority
            className="h-8 w-auto object-contain drop-shadow-[0_0_15px_rgba(255,184,0,0.25)]"
          />
        </div>
        <div className="gap-8 items-center text-xs font-mono font-bold uppercase tracking-widest text-gray-400 hidden md:flex">
          <Link href="/docs" className="hover:text-[#FFB800] transition-colors">Mission Log</Link>
          <a href="https://github.com/Segniko/Apex" target="_blank" className="hover:text-[#FFB800] transition-colors">GitHub</a>
          <Link href={dashboardLink} className="border border-[#FFB800] text-[#FFB800] px-6 py-3 hover:bg-[#FFB800] hover:text-black transition-colors shadow-[0_0_20px_rgba(255,184,0,0.2)]">
            Command Center
          </Link>
        </div>
        {/* Mobile entry point */}
        <Link href={dashboardLink} className="md:hidden border border-[#FFB800] text-[#FFB800] px-4 py-2 text-[11px] font-mono font-bold uppercase tracking-widest">
          Launch
        </Link>
      </nav>

      {/* Hero Section */}
      <section className="relative min-h-screen flex flex-col justify-center items-center px-6 py-28 overflow-hidden">
        {/* Minimalist Tactical Background */}
        <div className="absolute inset-0 z-0 overflow-hidden">
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(255,184,0,0.05)_0%,transparent_70%)]" />
          <div className="absolute inset-0 opacity-[0.03]" style={{ backgroundImage: 'linear-gradient(#FFB800 1px, transparent 1px), linear-gradient(90deg, #FFB800 1px, transparent 1px)', backgroundSize: '40px 40px' }} />
          <div className="absolute inset-0 bg-gradient-to-b from-transparent via-[#080808]/50 to-[#080808]" />
          {/* Scanline Effect */}
          <div className="absolute inset-0 pointer-events-none opacity-[0.02]" style={{ backgroundImage: 'linear-gradient(rgba(255,184,0,0.5) 50%, transparent 50%)', backgroundSize: '100% 4px' }} />

          {/* Animated Diagonal Hazard Flow */}
          <div className="absolute -inset-[100%] z-0 pointer-events-none opacity-[0.15] [mask-image:radial-gradient(ellipse_at_center,black_10%,transparent_40%)] flex items-center justify-center transform -rotate-45">
            <div className="flex gap-20 animate-apex-slide">
              <div className="w-40 h-[400vh] bg-[#FFB800] shadow-[0_0_50px_rgba(255,184,0,0.5)]" />
              <div className="w-12 h-[400vh] bg-[#080808]" />
              <div className="w-40 h-[400vh] bg-[#FFB800] shadow-[0_0_50px_rgba(255,184,0,0.5)]" />
            </div>
          </div>
        </div>

        <div className="z-10 text-center space-y-8 max-w-5xl">
          <div className="inline-flex items-center gap-2 border border-[#FFB800]/30 px-3 py-1 rounded bg-[#FFB800]/5 backdrop-blur-sm">
            <div className="w-2 h-2 rounded-full bg-[#FFB800] animate-pulse" />
            <span className="text-[11px] font-black tracking-[0.4em] text-[#FFB800] uppercase">v2026 · Deployment Active</span>
          </div>

          <h1 className="text-6xl sm:text-7xl md:text-9xl font-black italic tracking-tighter uppercase leading-[0.85] drop-shadow-2xl">
            ARCHITECTURE OF <br />
            <span className="text-[#FFB800]">RECOVERY.</span>
          </h1>

          <p className="text-gray-300 font-mono text-sm md:text-lg max-w-2xl mx-auto leading-relaxed">
            Industrial-grade observability, 100% free and open source. Because high-performance recovery shouldn&apos;t have a price tag.
          </p>

          <div className="flex flex-col md:flex-row gap-4 md:gap-6 justify-center pt-8">
            <Link href={dashboardLink} className="bg-[#FFB800] text-black px-10 py-4 font-black uppercase tracking-tighter hover:bg-white transition-all transform hover:scale-105 active:scale-95 shadow-[0_0_30px_rgba(255,184,0,0.3)]">
              Initialize Deployment
            </Link>
            <a href="https://github.com/Segniko/Apex" target='_blank' className="border border-[#222] text-gray-300 px-10 py-4 font-black uppercase tracking-tighter hover:border-[#FFB800] hover:text-white transition-all backdrop-blur-md">
              Fork on GitHub
            </a>
          </div>
        </div>

        <div className="absolute bottom-10 left-1/2 -translate-x-1/2 animate-bounce opacity-20">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M7 13l5 5 5-5M7 6l5 5 5-5" /></svg>
        </div>
      </section>

      {/* Features Grid */}
      <section className="max-w-7xl mx-auto px-6 py-24 md:py-32 space-y-24 md:space-y-32">
        <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-6 md:gap-8">
          {[
            { title: 'Uncrashable Ingest', desc: 'Redis-buffered streams decouple signal from storage. High throughput, zero data loss.', tag: 'OSS_Core' },
            { title: 'Global Persistence', desc: 'Deploy CockroachDB clusters globally. High-availability persistence for the people.', tag: 'Scale_Ready' },
            { title: 'AI Forensics HUD', desc: 'Automated root-cause analysis is standard. Every developer gets a forensic concierge.', tag: 'Intelligence' },
            { title: 'Tactical Edge', desc: 'Open agents for Go, Python, and Node. Lightweight, compressed, and strictly zero-license.', tag: 'Multi_Stack' },
          ].map((f, i) => (
            <Reveal key={f.title} delay={i * 80}>
              <FeatureCard {...f} />
            </Reveal>
          ))}
        </div>

        {/* Live HUD Preview (replaces empty placeholder) */}
        <Reveal>
          <HudPreview />
        </Reveal>

        {/* Detailed Explanation */}
        <div className="grid md:grid-cols-2 gap-12 md:gap-20 items-center">
          <Reveal className="space-y-8">
            <h2 className="text-3xl md:text-4xl font-black italic uppercase tracking-tighter">Democratizing <span className="text-[#FFB800]">Observability.</span></h2>
            <div className="space-y-6 text-gray-300 font-mono text-sm leading-relaxed">
              <p>Observability shouldn&apos;t be a luxury. We believe every engineer—from students to enterprise architects—deserves access to high-fidelity failure forensics.</p>
              <p>Apex isn&apos;t just a tool; it&apos;s a statement. We&apos;ve taken the technology used by trillion-dollar giants and made it accessible to everyone, for free.</p>
              <ul className="space-y-4">
                <li className="flex gap-3"><span className="text-[#FFB800]">●</span> No hidden fees, no &quot;Pro&quot; features</li>
                <li className="flex gap-3"><span className="text-[#FFB800]">●</span> Fully self-hostable on any cloud</li>
                <li className="flex gap-3"><span className="text-[#FFB800]">●</span> Driven by the developer community</li>
              </ul>
            </div>
          </Reveal>
          <Reveal delay={120}>
            <div className="relative aspect-video bg-[#0c0c0c] rounded-lg border border-[#222] overflow-hidden group shadow-2xl">
              <div className="absolute inset-0 hazard-pattern opacity-5 group-hover:opacity-10 transition-opacity" />
              <div className="absolute inset-0 flex items-center justify-center">
                <span className="text-[11px] font-black text-[#FFB800] tracking-[0.4em] uppercase">Free Forever Engine</span>
              </div>
            </div>
          </Reveal>
        </div>

        {/* The Paradigm Shift: Apex vs Legacy */}
        <div className="space-y-12 md:space-y-16 pt-12 md:pt-20">
          <Reveal className="text-center space-y-4">
            <h2 className="text-4xl md:text-5xl font-black italic uppercase tracking-tighter">The <span className="text-[#FFB800]">Paradigm Shift</span></h2>
            <p className="text-gray-400 font-mono text-sm uppercase tracking-widest max-w-2xl mx-auto">Why the Architecture of Recovery rethinks legacy monitoring platforms.</p>
          </Reveal>
          <div className="grid md:grid-cols-3 gap-6">
            {[
              { title: 'Throughput Limit', legacy: 'Throttles and drops events during high-traffic incidents (when you need it most).', apex: 'Redis-buffered ingestion absorbs traffic spikes and offloads to storage with zero data loss.' },
              { title: 'Forensic Intelligence', legacy: 'Provides raw, noisy stack traces leaving you to guess the root cause.', apex: 'Built-in AI concierge instantly decodes the trace and provides a tactical step-by-step fix.' },
              { title: 'Cost Model', legacy: "Charges exorbitant tiered fees based on 'event volume' and 'seats'.", apex: '100% free and open source. Deploy the entire infrastructure yourself, forever.' },
            ].map((c, i) => (
              <Reveal key={c.title} delay={i * 80}>
                <CompareCard {...c} />
              </Reveal>
            ))}
          </div>
        </div>

        {/* How It Works (Zero to HUD) */}
        <div className="space-y-12 md:space-y-16 pt-12 md:pt-20">
          <Reveal className="text-center space-y-4">
            <h2 className="text-3xl md:text-4xl font-black italic uppercase tracking-tighter">Zero to HUD in <span className="text-[#FFB800]">60 Seconds</span></h2>
            <p className="text-gray-400 font-mono text-sm uppercase tracking-widest">Integrating the Edge Agent is dangerously simple.</p>
          </Reveal>
          <div className="grid md:grid-cols-3 gap-6">
            <Reveal delay={0}>
              <StepCard num="01" title="Get Keys" code={`// 1. Sign in with GitHub\n// 2. Create a workspace\n// 3. Copy ingest key "apex_..."`} />
            </Reveal>
            <Reveal delay={80}>
              <StepCard num="02" title="Initialize" code={`import "github.com/Segniko/Apex/pkg/agent"\n\na := agent.New(v, s, cfg)\ndefer a.CapturePanic()`} />
            </Reveal>
            <Reveal delay={160}>
              <StepCard num="03" title="Recover" code={`// 1. App panics.\n// 2. DNA extracted + synced.\n// 3. AI HUD decodes the fix.`} />
            </Reveal>
          </div>
        </div>

        {/* Call to Action */}
        <Reveal>
          <div className="text-center space-y-10 md:space-y-12 py-16 md:py-20 px-6 border border-[#222] bg-[#FFB800]/5 relative overflow-hidden">
            <div className="absolute top-0 right-0 w-32 h-32 hazard-pattern opacity-5 -mr-16 -mt-16 rotate-45" />
            <h2 className="text-4xl md:text-5xl font-black italic uppercase tracking-tighter">Join the <span className="text-[#FFB800]">Movement</span></h2>
            <p className="max-w-xl mx-auto text-gray-300 font-mono text-xs uppercase tracking-widest">Help us build the most powerful free monitoring engine on the planet.</p>
            <div className="flex flex-col sm:flex-row justify-center gap-4 sm:gap-6">
              <a href="https://github.com/Segniko/Apex" target="_blank" className="bg-[#FFB800] text-black px-12 py-4 font-black uppercase tracking-tighter hover:bg-white transition-all">
                GitHub Repository
              </a>
              <a href="https://github.com/Segniko/Apex/stargazers" target="_blank" className="border border-[#FFB800] text-[#FFB800] px-12 py-4 font-black uppercase tracking-tighter hover:bg-[#FFB800] hover:text-black transition-all">
                ★ Star on GitHub
              </a>
            </div>
          </div>
        </Reveal>
      </section>

      {/* Footer */}
      <footer className="border-t border-[#222] py-16 md:py-20 bg-[#0a0a0a]">
        <div className="max-w-7xl mx-auto px-6 flex flex-col md:flex-row justify-between items-center gap-8">
          <h2 className="text-2xl font-black italic uppercase tracking-tighter">APEX <span className="text-[#FFB800]">SYSTEMS</span></h2>
          <div className="flex flex-wrap justify-center gap-8 md:gap-12 text-[11px] font-mono text-gray-400 uppercase tracking-widest">
            <Link href={dashboardLink} className="hover:text-[#FFB800]">Dashboard</Link>
            <Link href="/docs" className="hover:text-[#FFB800]">Mission Log</Link>
            <span className="text-gray-700">© {new Date().getFullYear()} Apex · MIT</span>
          </div>
        </div>
      </footer>
    </div>
  );
}

function HudPreview() {
  return (
    <div className="border border-[#222] bg-[#0c0c0c] rounded-lg overflow-hidden shadow-2xl">
      <div className="flex items-center gap-2 px-4 py-3 bg-[#111] border-b border-[#222]">
        <div className="w-3 h-3 rounded-full bg-[#FF4D00]" />
        <div className="w-3 h-3 rounded-full bg-[#FFB800]" />
        <div className="w-3 h-3 rounded-full bg-[#00FF41]" />
        <span className="ml-3 text-[11px] font-mono text-gray-500 tracking-widest uppercase">apex · live forensics feed</span>
        <span className="ml-auto flex items-center gap-2 text-[11px] font-mono text-[#00FF41]">
          <span className="w-2 h-2 rounded-full bg-[#00FF41] animate-pulse" /> streaming
        </span>
      </div>
      <div className="grid md:grid-cols-[1fr_220px] gap-0">
        <div className="p-5 space-y-4 font-mono text-sm border-r border-[#222]">
          <div className="border-l-2 border-[#FF4D00] pl-4">
            <div className="text-white font-bold">runtime: nil pointer dereference</div>
            <div className="text-gray-500 text-xs mt-1">handler.go:88 · 1,204 events · last seen 2s ago</div>
          </div>
          <pre className="bg-black/60 border border-[#1a1a1a] rounded p-3 text-xs trace-surface overflow-x-auto">
{`goroutine 42 [running]:
  main.(*Handler).Serve(0x0)
      handler.go:88 +0x1f
  net/http.(*conn).serve()
      server.go:1995 +0x8d`}
          </pre>
          <div className="bg-[#FFB800]/5 border border-[#FFB800]/30 rounded p-3">
            <div className="text-[11px] font-black text-[#FFB800] uppercase tracking-widest mb-1">AI root cause</div>
            <p className="text-xs text-gray-300 leading-relaxed">
              <span className="text-[#FFB800]">h.cache</span> is nil because <span className="text-[#FFB800]">NewHandler</span> returned before initialization. Guard the call or initialize in the constructor.
            </p>
          </div>
        </div>
        <div className="p-5 space-y-4">
          <Spark />
          <div className="grid grid-cols-2 gap-3 text-center">
            <Metric label="Events" value="1.2k" />
            <Metric label="Issues" value="14" />
            <Metric label="Resolved" value="9" color="#00FF41" />
            <Metric label="P95 decode" value="1.4s" />
          </div>
        </div>
      </div>
    </div>
  );
}

function Spark() {
  const points = [4, 7, 5, 9, 6, 11, 8, 14, 10, 16, 12, 19];
  const max = Math.max(...points);
  const w = 180;
  const h = 48;
  const step = w / (points.length - 1);
  const path = points
    .map((p, i) => `${i === 0 ? 'M' : 'L'} ${(i * step).toFixed(1)} ${(h - (p / max) * h).toFixed(1)}`)
    .join(' ');
  return (
    <div>
      <div className="text-[11px] font-mono text-gray-500 uppercase tracking-widest mb-2">Last 12h</div>
      <svg viewBox={`0 0 ${w} ${h}`} className="w-full h-12" preserveAspectRatio="none">
        <path d={`${path} L ${w} ${h} L 0 ${h} Z`} fill="rgba(255,184,0,0.08)" />
        <path d={path} fill="none" stroke="#FFB800" strokeWidth="1.5" />
      </svg>
    </div>
  );
}

function Metric({ label, value, color = '#FFB800' }: { label: string; value: string; color?: string }) {
  return (
    <div className="bg-[#111] border border-[#222] rounded p-2">
      <div className="text-lg font-black italic tracking-tighter" style={{ color }}>{value}</div>
      <div className="text-[10px] font-mono text-gray-500 uppercase tracking-widest">{label}</div>
    </div>
  );
}

function FeatureCard({ title, desc, tag }: { title: string, desc: string, tag: string }) {
  return (
    <div className="group bg-[#111] p-8 border border-[#222] hover:border-[#FFB800]/50 transition-all space-y-4 shadow-xl h-full">
      <span className="text-[10px] font-black text-[#FFB800] tracking-widest uppercase opacity-50 group-hover:opacity-100 transition-opacity">[{tag}]</span>
      <h3 className="text-xl font-black italic uppercase tracking-tight">{title}</h3>
      <p className="text-gray-400 text-sm font-mono leading-relaxed">{desc}</p>
    </div>
  );
}

function CompareCard({ title, apex, legacy }: { title: string, apex: string, legacy: string }) {
  return (
    <div className="border border-[#222] bg-[#111] p-8 space-y-6 hover:border-[#FFB800]/30 transition-all h-full">
      <h4 className="text-xl font-black italic tracking-tighter text-white uppercase">{title}</h4>
      <div className="space-y-6">
        <div>
          <span className="text-[10px] text-gray-500 tracking-widest uppercase font-mono">Legacy Platform</span>
          <p className="text-gray-400 mt-2 text-sm leading-relaxed line-through decoration-red-900">{legacy}</p>
        </div>
        <div>
          <span className="text-[10px] text-[#FFB800] tracking-widest uppercase font-mono flex items-center gap-2">
            <span className="w-1.5 h-1.5 bg-[#FFB800]"></span>
            Apex Engine
          </span>
          <p className="text-white mt-2 text-sm leading-relaxed">{apex}</p>
        </div>
      </div>
    </div>
  );
}

function StepCard({ num, title, code }: { num: string, title: string, code: string }) {
  return (
    <div className="relative p-8 border border-[#222] bg-[#0a0a0a] overflow-hidden group hover:border-[#FFB800]/50 transition-all h-full">
      <div className="absolute top-0 right-0 p-4 text-[#FFB800]/5 font-black italic text-8xl transition-all group-hover:text-[#FFB800]/10 group-hover:scale-110 -mt-4 -mr-4">{num}</div>
      <h4 className="text-xl font-black uppercase tracking-tight text-[#FFB800] mb-6 relative z-10">{title}</h4>
      <pre className="text-xs trace-surface font-mono bg-black p-4 rounded border border-[#222] overflow-x-auto relative z-10 shadow-inner">
        <code>{code}</code>
      </pre>
    </div>
  );
}
