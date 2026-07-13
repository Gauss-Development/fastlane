import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { Project } from "@/lib/design/types";

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const authorization = request.headers.get("authorization");
  if (!authorization) {
    return toFailureResponse(401, "UNAUTHORIZED", "Authentication required.");
  }

  const { id } = await params;
  const { upstream, payload } = await proxyGateway<Project>(request, `/api/v1/projects/${encodeURIComponent(id)}`, {
    method: "GET",
    headers: { authorization },
  });
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "GET_PROJECT_FAILED", "Failed to retrieve project.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
