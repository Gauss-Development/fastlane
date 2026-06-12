import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { ListQuotesResponse } from "@/lib/rfqs/types";

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const authorization = request.headers.get("authorization");
  if (!authorization) {
    return toFailureResponse(401, "UNAUTHORIZED", "Authentication required.");
  }

  const { id } = await params;
  const { upstream, payload } = await proxyGateway<ListQuotesResponse>(
    request,
    `/api/v1/rfqs/${encodeURIComponent(id)}/quotes`,
    { method: "GET", headers: { authorization } },
  );
  if (!payload?.success || payload.data === undefined) {
    return toFailureResponse(upstream.status, "LIST_QUOTES_FAILED", "Failed to retrieve quotes.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
