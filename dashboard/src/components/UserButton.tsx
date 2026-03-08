'use client';

import { signOut, useSession } from "next-auth/react";

export function UserButton() {
    const { data: session } = useSession();

    if (!session?.user) return null;

    const name = session.user.name || session.user.email || "User";
    const initial = name.charAt(0).toUpperCase();

    return (
        <div className="flex items-center gap-4 bg-[#111] border border-[#222] p-2 pl-4 rounded">
            <div className="flex flex-col items-end">
                <span className="text-[10px] font-black text-white italic uppercase tracking-tighter">{name}</span>
                <button
                    onClick={() => signOut({ callbackUrl: "/" })}
                    className="text-[8px] font-mono text-[#FFB800] hover:text-white uppercase tracking-widest transition-colors"
                >
                    [ Terminate_Session ]
                </button>
            </div>
            <div className="w-10 h-10 bg-[#FFB800] text-black flex items-center justify-center font-black italic text-xl shadow-[0_0_15px_rgba(255,184,0,0.3)]">
                {initial}
            </div>
        </div>
    );
}
