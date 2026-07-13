import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { ListProjectsResponse, Project } from "@/lib/design/types";

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

  const { upstream, payload } = await proxyGateway<ListProjectsResponse>(request, `/api/v1/projects${suffix}`, {
    method: "GET",
    headers: { authorization },
  });
  if (!payload?.success || payload.data === undefined) {
    return toFailureResponse(upstream.status, "LIST_PROJECTS_FAILED", "Failed to retrieve projects.", payload?.error);
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

  const { upstream, payload } = await proxyGateway<Project>(request, "/api/v1/projects", {
    method: "POST",
    headers: { authorization },
    body: JSON.stringify(rawBody),
  });
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "CREATE_PROJECT_FAILED", "Failed to create project.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
