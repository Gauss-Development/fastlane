import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { Quote } from "@/lib/rfqs/types";

export async function POST(
  request: Request,
  { params }: { params: Promise<{ id: string; quoteId: string }> },
) {
  const authorization = request.headers.get("authorization");
  if (!authorization) {
    return toFailureResponse(401, "UNAUTHORIZED", "Authentication required.");
  }

  const { id, quoteId } = await params;
  const { upstream, payload } = await proxyGateway<Quote>(
    request,
    `/api/v1/rfqs/${encodeURIComponent(id)}/quotes/${encodeURIComponent(quoteId)}/accept`,
    { method: "POST", headers: { authorization } },
  );
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "ACCEPT_QUOTE_FAILED", "Failed to accept quote.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
