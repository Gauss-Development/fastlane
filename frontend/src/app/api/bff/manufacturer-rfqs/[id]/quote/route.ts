import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { Quote } from "@/lib/rfqs/types";

export async function POST(
  request: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const authorization = request.headers.get("authorization");
  if (!authorization) {
    return toFailureResponse(401, "UNAUTHORIZED", "Authentication required.");
  }

  const rawBody = await request.json().catch(() => null);
  if (!rawBody || typeof rawBody !== "object") {
    return toFailureResponse(400, "INVALID_REQUEST", "Invalid request body.");
  }

  const { id } = await params;
  const { upstream, payload } = await proxyGateway<Quote>(
    request,
    `/api/v1/manufacturer-rfqs/${encodeURIComponent(id)}/quote`,
    { method: "POST", headers: { authorization }, body: JSON.stringify(rawBody) },
  );
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "SUBMIT_QUOTE_FAILED", "Failed to submit quote.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
