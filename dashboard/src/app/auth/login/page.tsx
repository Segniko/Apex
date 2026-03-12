import Link from 'next/link';
import { auth } from '../../../auth';
import { redirect } from 'next/navigation';

export default async function Login() {
    const session = await auth();
    if (session) {
        redirect("/dashboard/projects");
    }

    return (
        <div className="min-h-screen bg-[#080808] flex items-center justify-center p-6 selection:bg-[#FFB800] selection:text-black">
            <div className="absolute inset-0 z-0 opacity-10 hazard-pattern pointer-events-none" />

            <div className="w-full max-w-md bg-[#111] border border-[#222] p-8 md:p-12 relative z-10 shadow-2xl">
                <div className="flex flex-col items-center mb-10">
                    <div className="w-8 h-8 bg-[#FFB800] animate-pulse mb-6 shadow-[0_0_20px_rgba(255,184,0,0.4)]" />
                    <h1 className="text-3xl font-black italic tracking-tighter uppercase text-white">APEX <span className="text-[#FFB800]">ID</span></h1>
                    <p className="text-gray-500 font-mono text-xs uppercase tracking-widest mt-2">Authentication Required</p>
                </div>

                <div className="space-y-4">
                    <form
                        action={async () => {
                            "use server"
                            const { signIn } = await import("../../../auth")
                            await signIn("github", { redirectTo: "/dashboard/projects" })
                        }}
                        className="w-full"
                    >
                        <button type="submit" className="w-full border border-[#222] bg-black hover:border-white transition-colors p-4 flex items-center justify-center gap-3 group">
                            <svg viewBox="0 0 24 24" className="w-5 h-5 fill-white group-hover:scale-110 transition-transform"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" /></svg>
                            <span className="font-mono text-xs uppercase text-gray-400 group-hover:text-white transition-colors">Continue with GitHub</span>
                        </button>
                    </form>

                </div>

                <div className="mt-8 text-center border-t border-[#222] pt-6">
                    <Link href="/" className="text-gray-500 font-mono text-[10px] uppercase tracking-widest hover:text-[#FFB800] transition-colors">Abort & Return to Surface</Link>
                </div>
            </div>
        </div>
    );
}
