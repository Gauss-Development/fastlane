import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { ListRFQsResponse } from "@/lib/rfqs/types";

export async function GET(request: Request) {
  const authorization = request.headers.get("authorization");
  if (!authorization) {
    return toFailureResponse(401, "UNAUTHORIZED", "Authentication required.");
  }

  const { searchParams } = new URL(request.url);
  const query = new URLSearchParams();
  for (const key of ["limit", "offset"] as const) {
    const value = searchParams.get(key);
    if (value) query.set(key, value);
  }
  const suffix = query.size > 0 ? `?${query.toString()}` : "";

  const { upstream, payload } = await proxyGateway<ListRFQsResponse>(request, `/api/v1/manufacturer-rfqs${suffix}`, {
    method: "GET",
    headers: { authorization },
  });
  if (!payload?.success || payload.data === undefined) {
    return toFailureResponse(upstream.status, "LIST_RFQS_FAILED", "Failed to retrieve open RFQs.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
