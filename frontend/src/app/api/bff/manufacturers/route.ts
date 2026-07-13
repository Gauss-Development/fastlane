import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { ListManufacturersResponse, Manufacturer } from "@/lib/manufacturers/types";

export async function GET(request: Request) {
  const authorization = request.headers.get("authorization");
  if (!authorization) {
    return toFailureResponse(401, "UNAUTHORIZED", "Authentication required.");
  }

  const { searchParams } = new URL(request.url);
  const query = new URLSearchParams();
  for (const key of [
    "cluster",
    "service_type",
    "assembly_type",
    "material",
    "verified_only",
    "min_layers_gte",
    "limit",
    "offset",
  ] as const) {
    const value = searchParams.get(key);
    if (value) query.set(key, value);
  }
  const suffix = query.size > 0 ? `?${query.toString()}` : "";

  const { upstream, payload } = await proxyGateway<ListManufacturersResponse>(request, `/api/v1/manufacturers${suffix}`, {
    method: "GET",
    headers: { authorization },
  });
  if (!payload?.success || payload.data === undefined) {
    return toFailureResponse(upstream.status, "LIST_MANUFACTURERS_FAILED", "Failed to retrieve manufacturers.", payload?.error);
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

  const { upstream, payload } = await proxyGateway<Manufacturer>(request, "/api/v1/manufacturers", {
    method: "POST",
    headers: { authorization },
    body: JSON.stringify(rawBody),
  });
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "CREATE_MANUFACTURER_FAILED", "Failed to create manufacturer.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
