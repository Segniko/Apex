import { NextRequest, NextResponse } from "next/server";
import { requireSession, proxyGet } from "../_proxy";

// Returns crash reports for a project. Requires an authenticated session.
export async function GET(req: NextRequest) {
  const authed = await requireSession();
  if (authed instanceof NextResponse) return authed;
  const projectId = req.nextUrl.searchParams.get("project_id") || "";
  const qs = projectId ? `?project_id=${encodeURIComponent(projectId)}` : "";
  return proxyGet(`/api/reports${qs}`);
}
