'use client';

import { useState } from 'react';
import { Project } from '@/lib/api';

interface OnboardingGuideProps {
    project: Project;
}

export function OnboardingGuide({ project }: OnboardingGuideProps) {
    const [selectedTab, setSelectedTab] = useState<'go' | 'python' | 'node'>('go');

    const snippets = {
        go: `package main

import (
    "github.com/apex-forensics/apex-go/agent"
)

func main() {
    // Initialize Apex Forensics Agent
    apex := agent.New("${project.ingest_key}")
    defer apex.Recover() // Captures panics automatically

    // Your application code...
}`,
        python: `from apex_py import ApexAgent

# Initialize the Apex Forensics Agent
agent = ApexAgent("${project.ingest_key}")

try:
    # Your application logic...
    risky_operation()
except Exception as e:
    # Manual capture for non-fatal errors
    agent.capture_exception(e)`,
        node: `const { ApexAgent } = require('apex-node');

// Initialize the Tactical Forwarder
const agent = new ApexAgent("${project.ingest_key}");

async function run() {
    try {
        await riskyTask();
    } catch (err) {
        await agent.captureException(err);
    }
}`
    };

    return (
        <div className="bg-[#111] border border-[#222] rounded-xl overflow-hidden shadow-2xl">
            {/* Header */}
            <div className="p-8 border-b border-[#222] bg-gradient-to-r from-[#111] to-[#1a1a1a]">
                <div className="flex items-center gap-4 mb-4">
                    <div className="bg-[#FFB800] text-black text-[10px] font-black px-2 py-0.5 rounded uppercase tracking-widest">Awaiting_Signal</div>
                    <div className="h-[1px] w-12 bg-[#333]" />
                    <span className="text-[10px] font-mono text-gray-500 uppercase">Onboarding_Initial_Forensics</span>
                </div>
                <h2 className="text-4xl font-black italic tracking-tighter text-white uppercase mb-4">
                    Project <span className="text-[#FFB800]">Activation</span>
                </h2>
                <p className="text-gray-400 text-sm max-w-2xl leading-relaxed italic border-l-2 border-[#FFB800] pl-4">
                    "Connect your infrastructure to the Apex network. Once the first signal is captured, this tactical HUD will synchronize with your live telemetry."
                </p>
            </div>

            {/* Ingest Key Box */}
            <div className="p-8 bg-[#0c0c0c] border-b border-[#222]">
                <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-6">
                    <div>
                        <span className="text-[10px] font-black text-[#FFB800] uppercase tracking-[0.3em] block mb-2">Project_Ingest_Key</span>
                        <code className="text-2xl font-mono text-white bg-black/50 p-3 rounded border border-[#333] select-all">
                            {project.ingest_key}
                        </code>
                    </div>
                </div>
            </div>

            {/* Language Selection */}
            <div className="flex border-b border-[#222] bg-[#111]">
                {(['go', 'python', 'node'] as const).map(lang => (
                    <button
                        key={lang}
                        onClick={() => setSelectedTab(lang)}
                        className={`px-8 py-4 text-[10px] font-black uppercase tracking-widest transition-all ${
                            selectedTab === lang 
                                ? 'bg-[#FFB800] text-black shadow-[inset_0_2px_10px_rgba(0,0,0,0.2)]' 
                                : 'text-gray-500 hover:text-white border-r border-[#222]'
                        }`}
                    >
                        {lang === 'node' ? 'Node.js' : lang.toUpperCase()} Agent
                    </button>
                ))}
            </div>

            {/* Code Snippet */}
            <div className="p-8 bg-black/40 font-mono text-[13px] leading-relaxed">
                <div className="flex items-center gap-3 mb-6">
                    <div className="w-1.5 h-1.5 rounded-full bg-[#00FF41] animate-pulse" />
                    <span className="text-[10px] text-gray-500 uppercase tracking-widest italic">Installation_Manifest</span>
                </div>
                <pre className="text-gray-300 overflow-x-auto p-6 bg-[#050505] border border-[#1a1a1a] rounded-lg">
                    <code>{snippets[selectedTab]}</code>
                </pre>
            </div>

            {/* Footer Tip */}
            <div className="p-6 bg-[#111] text-[10px] text-gray-600 font-mono italic text-center uppercase tracking-widest">
                System optimized for zero-latency crash forwarding // Verion_1.0_Tactical
            </div>
        </div>
    );
}
