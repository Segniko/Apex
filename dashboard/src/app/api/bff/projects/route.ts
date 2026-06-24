import { NextResponse } from "next/server";
import { requireSession, proxyGet } from "../_proxy";

// Lists projects for the *authenticated* user. The user id is taken from the
// session, never the client, so a caller can't enumerate another user's data.
export async function GET() {
  const authed = await requireSession();
  if (authed instanceof NextResponse) return authed;
  return proxyGet(`/api/projects?user_id=${encodeURIComponent(authed.userId)}`);
}
