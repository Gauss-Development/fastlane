import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { Quote } from "@/lib/rfqs/types";

// Public: the magic-link token in the path is the only credential.
export async function POST(
  request: Request,
  { params }: { params: Promise<{ token: string }> },
) {
  const { token } = await params;

  const rawBody = await request.json().catch(() => null);
  if (!rawBody || typeof rawBody !== "object") {
    return toFailureResponse(400, "INVALID_REQUEST", "Invalid request body.");
  }

  const { upstream, payload } = await proxyGateway<Quote>(
    request,
    `/api/v1/supplier-rfq/${encodeURIComponent(token)}/quote`,
    { method: "POST", body: JSON.stringify(rawBody) },
  );
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "SUBMIT_QUOTE_FAILED", "Failed to submit quote.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
