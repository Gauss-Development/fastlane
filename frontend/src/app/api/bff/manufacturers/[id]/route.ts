import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { Manufacturer } from "@/lib/manufacturers/types";

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const authorization = request.headers.get("authorization");
  if (!authorization) {
    return toFailureResponse(401, "UNAUTHORIZED", "Authentication required.");
  }

  const { id } = await params;
  const { upstream, payload } = await proxyGateway<Manufacturer>(request, `/api/v1/manufacturers/${encodeURIComponent(id)}`, {
    method: "GET",
    headers: { authorization },
  });
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "GET_MANUFACTURER_FAILED", "Failed to retrieve manufacturer.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}

export async function PUT(
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
  const { upstream, payload } = await proxyGateway<Manufacturer>(request, `/api/v1/manufacturers/${encodeURIComponent(id)}`, {
    method: "PUT",
    headers: { authorization },
    body: JSON.stringify(rawBody),
  });
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "UPDATE_MANUFACTURER_FAILED", "Failed to update manufacturer.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
