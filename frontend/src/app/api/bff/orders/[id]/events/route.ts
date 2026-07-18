import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { ListOrderEventsResponse, OrderEvent } from "@/lib/orders/types";

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const authorization = request.headers.get("authorization");
  if (!authorization) {
    return toFailureResponse(401, "UNAUTHORIZED", "Authentication required.");
  }

  const { id } = await params;
  const { upstream, payload } = await proxyGateway<ListOrderEventsResponse>(
    request,
    `/api/v1/orders/${encodeURIComponent(id)}/events`,
    { method: "GET", headers: { authorization } },
  );
  if (!payload?.success || payload.data === undefined) {
    return toFailureResponse(upstream.status, "LIST_EVENTS_FAILED", "Failed to retrieve order events.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}

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
  const { upstream, payload } = await proxyGateway<OrderEvent>(
    request,
    `/api/v1/orders/${encodeURIComponent(id)}/events`,
    { method: "POST", headers: { authorization }, body: JSON.stringify(rawBody) },
  );
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "APPEND_EVENT_FAILED", "Failed to append order event.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
