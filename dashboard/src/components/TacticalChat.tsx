'use client';

import { useEffect, useRef, useState } from 'react';

export function TacticalChat() {
    const [isOpen, setIsOpen] = useState(false);
    const [messages, setMessages] = useState<{ role: 'ai' | 'user', text: string }[]>([
        { role: 'ai', text: 'APEX_AI_V1 online. Awaiting tactical inquiry.' }
    ]);
    const [input, setInput] = useState('');
    const [loading, setLoading] = useState(false);
    const [reportId, setReportId] = useState<string | null>(null);
    const scrollRef = useRef<HTMLDivElement>(null);

    const performChat = async (userMsg: string, currentReportId: string | null) => {
        if (!userMsg.trim() || loading) return;

        setLoading(true);
        // Add a placeholder for the AI response
        setMessages(prev => [...prev, { role: 'ai', text: '' }]);

        try {
            const apiBase = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081';
            const res = await fetch(`${apiBase}/api/chat`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ 
                    message: userMsg,
                    report_id: currentReportId || "" 
                })
            });

            if (!res.body) throw new Error("No response body");

            const reader = res.body.getReader();
            const decoder = new TextDecoder();
            let aiText = "";

            while (true) {
                const { value, done } = await reader.read();
                if (done) break;

                const chunk = decoder.decode(value);
                const lines = chunk.split('\n');

                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        const content = line.slice(6);
                        if (content.trim() === '[DONE]') break;
                        if (content === '') continue;

                        aiText += content;
                        
                        // Update the last message (AI response)
                        setMessages(prev => {
                            const newMessages = [...prev];
                            newMessages[newMessages.length - 1] = { role: 'ai', text: aiText };
                            return newMessages;
                        });
                    }
                }
            }
        } catch (err) {
            setMessages(prev => {
                const newMessages = [...prev];
                newMessages[newMessages.length - 1] = { role: 'ai', text: 'CONNECTION_ERROR: Failed to reach tactical node.' };
                return newMessages;
            });
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        const handleContextChat = (e: any) => {
            const { errorId, message } = e.detail;
            setIsOpen(true);
            setReportId(errorId);
            
            const introMsg = `APEX_AI initialized with Error_ID: ${errorId.substring(0, 12)}. Analyzing specific telemetry...`;
            const userQuery = `Can you explain why this error happened? "${message}"`;
            
            setMessages([
                { role: 'ai', text: introMsg },
                { role: 'user', text: userQuery }
            ]);
            
            performChat(userQuery, errorId);
        };

        window.addEventListener('apex-chat-context', handleContextChat);
        return () => window.removeEventListener('apex-chat-context', handleContextChat);
    }, []);

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
        performChat(userMsg, reportId);
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
                                <div className={`max-w-[90%] p-3 rounded text-[11px] font-mono leading-relaxed ${m.role === 'ai'
                                        ? 'bg-[#FFB800]/5 border border-[#FFB800]/20 text-[#FFB800]/80'
                                        : 'bg-[#222] text-gray-300 border border-[#333]'
                                    }`}>
                                    <span className="opacity-40 block mb-1 uppercase font-black tracking-tighter text-[8px]">
                                        {m.role === 'ai' ? 'APEX_DECODE' : 'OPERATOR'}
                                    </span>
                                    <div className="whitespace-pre-wrap">
                                        {m.text.split('```').map((part, index) => {
                                            if (index % 2 === 1) {
                                                // Code block
                                                const lines = part.split('\n');
                                                const lang = lines[0].trim();
                                                const code = lines.slice(1).join('\n');
                                                const isDiff = lang === 'diff' || code.trim().startsWith('---') || code.trim().startsWith('@@');

                                                return (
                                                    <div key={index} className="my-2 p-2 bg-black border border-[#FFB800]/10 rounded overflow-x-auto">
                                                        {lang && <div className="text-[8px] opacity-30 mb-1 uppercase">{lang}</div>}
                                                        <div className="text-[10px] leading-tight">
                                                            {code.split('\n').map((line, li) => (
                                                                <div key={li} className={
                                                                    isDiff ? (
                                                                        line.startsWith('+') ? 'text-green-500/80 bg-green-500/5' :
                                                                        line.startsWith('-') ? 'text-red-500/80 bg-red-500/5' :
                                                                        line.startsWith('@@') ? 'text-blue-400/60' : ''
                                                                    ) : ''
                                                                }>
                                                                    {line}
                                                                </div>
                                                            ))}
                                                        </div>
                                                    </div>
                                                );
                                            }
                                            return part;
                                        })}
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
