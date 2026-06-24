import { NextRequest, NextResponse } from "next/server";
import { requireSession, proxyGet } from "../_proxy";

// Returns deduplicated issues for a project. Requires an authenticated session.
export async function GET(req: NextRequest) {
  const authed = await requireSession();
  if (authed instanceof NextResponse) return authed;
  const sp = req.nextUrl.searchParams;
  const projectId = sp.get("project_id") || "";
  const limit = sp.get("limit") || "50";
  const offset = sp.get("offset") || "0";
  const qs = `?project_id=${encodeURIComponent(projectId)}&limit=${limit}&offset=${offset}`;
  return proxyGet(`/api/issues${qs}`);
}
