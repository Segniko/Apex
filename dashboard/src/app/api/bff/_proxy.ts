import { auth } from "@/auth";
import { NextResponse } from "next/server";

// Server-only receiver address + shared secret. The browser never sees these;
// the BFF authenticates the session, then talks to the Go receiver on its behalf.
const RECEIVER =
  process.env.APEX_RECEIVER_URL ||
  process.env.NEXT_PUBLIC_API_URL ||
  "http://localhost:8081";
const INTERNAL_KEY = process.env.APEX_INTERNAL_SECRET || "";

export interface AuthedRequest {
  userId: string;
}

// Verifies there is a signed-in user and returns their stable id, or a 401.
export async function requireSession(): Promise<AuthedRequest | NextResponse> {
  const session = await auth();
  const userId =
    (session?.user as { id?: string } | undefined)?.id || session?.user?.email;
  if (!session || !userId) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }
  return { userId };
}

// Proxies a GET to the receiver with the internal key attached.
export async function proxyGet(path: string): Promise<NextResponse> {
  const res = await fetch(`${RECEIVER}${path}`, {
    headers: INTERNAL_KEY ? { "X-Apex-Internal-Key": INTERNAL_KEY } : {},
    cache: "no-store",
  });
  const body = await res.text();
  return new NextResponse(body, {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}
