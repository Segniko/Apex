'use client';

import { useEffect, useRef, useState } from 'react';

export function TacticalChat() {
    const [isOpen, setIsOpen] = useState(false);
    const [messages, setMessages] = useState<{ role: 'ai' | 'user', text: string }[]>([
        { role: 'ai', text: 'APEX_AI_V1 online. Awaiting tactical inquiry.' }
    ]);
    const [input, setInput] = useState('');
    const [loading, setLoading] = useState(false);
    const scrollRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [messages]);

    const sendMessage = async () => {
        if (!input.trim() || loading) return;

        const userMsg = input;
        setInput('');
        setMessages(prev => [...prev, { role: 'user', text: userMsg }]);
        setLoading(true);

        try {
            const res = await fetch('http://localhost:8081/api/chat', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ message: userMsg })
            });
            const data = await res.json();
            setMessages(prev => [...prev, { role: 'ai', text: data.response }]);
        } catch (err) {
            setMessages(prev => [...prev, { role: 'ai', text: 'CONNECTION_ERROR: Failed to reach tactical node.' }]);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="fixed bottom-8 right-8 z-[100] flex flex-col items-end">
            {isOpen && (
                <div className="mb-4 w-80 md:w-96 h-[450px] bg-[#111] border border-[#FFB800]/30 rounded-lg shadow-[0_0_40px_rgba(0,0,0,0.8)] flex flex-col overflow-hidden animate-in slide-in-from-bottom-5 duration-300">
                    {/* Header */}
                    <div className="hazard-pattern h-1 w-full" />
                    <div className="p-4 bg-[#1a1a1a] border-b border-[#222] flex justify-between items-center">
                        <div className="flex items-center gap-2">
                            <div className="w-2 h-2 rounded-full bg-[#FFB800] animate-pulse" />
                            <span className="text-[10px] font-black text-[#FFB800] uppercase tracking-widest">Tactical_Forensics_AI</span>
                        </div>
                        <button onClick={() => setIsOpen(false)} className="text-gray-500 hover:text-white">✕</button>
                    </div>

                    {/* Messages */}
                    <div ref={scrollRef} className="flex-1 overflow-y-auto p-4 space-y-4 scrollbar-hide bg-[#080808]">
                        {messages.map((m, i) => (
                            <div key={i} className={`flex ${m.role === 'user' ? 'justify-end' : 'justify-start'}`}>
                                <div className={`max-w-[85%] p-3 rounded text-[11px] font-mono leading-relaxed ${m.role === 'ai'
                                        ? 'bg-[#FFB800]/5 border border-[#FFB800]/20 text-[#FFB800]/80'
                                        : 'bg-[#222] text-gray-300 border border-[#333]'
                                    }`}>
                                    <span className="opacity-40 block mb-1 uppercase font-black tracking-tighter text-[8px]">
                                        {m.role === 'ai' ? 'APEX_DECODE' : 'OPERATOR'}
                                    </span>
                                    <div className="whitespace-pre-wrap">
                                        {m.text}
                                    </div>
                                </div>
                            </div>
                        ))}
                        {loading && (
                            <div className="flex justify-start">
                                <div className="bg-[#FFB800]/5 border border-[#FFB800]/20 p-3 rounded text-[11px] font-mono text-[#FFB800]/40 animate-pulse">
                                    ANALYZING_TELEMETRY...
                                </div>
                            </div>
                        )}
                    </div>

                    {/* Input */}
                    <div className="p-4 bg-[#1a1a1a] border-t border-[#222] flex gap-2">
                        <input
                            type="text"
                            value={input}
                            onChange={(e) => setInput(e.target.value)}
                            onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
                            placeholder="INPUT_QUERY_HERE..."
                            className="flex-1 bg-black border border-[#333] px-3 py-2 text-[10px] font-mono text-[#FFB800] focus:border-[#FFB800] outline-none transition-colors placeholder:text-gray-700"
                        />
                        <button
                            onClick={sendMessage}
                            disabled={loading}
                            className="bg-[#FFB800] text-black px-3 py-1 text-[10px] font-black uppercase tracking-tighter hover:bg-[#FFD700] transition-colors disabled:opacity-50"
                        >
                            SEND
                        </button>
                    </div>
                </div>
            )}

            {/* FAB */}
            <button
                onClick={() => setIsOpen(!isOpen)}
                className="group relative w-14 h-14 bg-[#FFB800] rounded-full flex items-center justify-center shadow-[0_0_20px_rgba(255,184,0,0.3)] hover:scale-110 active:scale-95 transition-all duration-300"
            >
                <div className="absolute inset-0 rounded-full bg-[#FFB800] animate-ping opacity-20 group-hover:opacity-40" />
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="black" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
                    <path d="m21 15-3.5-2L21 11V7l-9-5L3 7v4l3.5 2L3 15v4l9 5 9-5v-4Z" />
                    <path d="M12 22v-5" />
                    <path d="m3.5 7 8.5 4.7 8.5-4.7" />
                    <path d="M12 11.7v5.3" />
                </svg>
            </button>
        </div>
    );
}
