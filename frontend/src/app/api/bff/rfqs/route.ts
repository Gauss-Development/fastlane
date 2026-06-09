import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { ListRFQsResponse, RFQ } from "@/lib/rfqs/types";

export async function GET(request: Request) {
  const authorization = request.headers.get("authorization");
  if (!authorization) {
    return toFailureResponse(401, "UNAUTHORIZED", "Authentication required.");
  }

  const { searchParams } = new URL(request.url);
  const query = new URLSearchParams();
  for (const key of ["status", "limit", "offset"] as const) {
    const value = searchParams.get(key);
    if (value) query.set(key, value);
  }
  const suffix = query.size > 0 ? `?${query.toString()}` : "";

  const { upstream, payload } = await proxyGateway<ListRFQsResponse>(request, `/api/v1/rfqs${suffix}`, {
    method: "GET",
    headers: { authorization },
  });
  if (!payload?.success || payload.data === undefined) {
    return toFailureResponse(upstream.status, "LIST_RFQS_FAILED", "Failed to retrieve RFQs.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}

export async function POST(request: Request) {
  const authorization = request.headers.get("authorization");
  if (!authorization) {
    return toFailureResponse(401, "UNAUTHORIZED", "Authentication required.");
  }

  const rawBody = await request.json().catch(() => null);
  if (!rawBody || typeof rawBody !== "object") {
    return toFailureResponse(400, "INVALID_REQUEST", "Invalid request body.");
  }

  const { upstream, payload } = await proxyGateway<RFQ>(request, "/api/v1/rfqs", {
    method: "POST",
    headers: { authorization },
    body: JSON.stringify(rawBody),
  });
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "CREATE_RFQ_FAILED", "Failed to create RFQ.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
