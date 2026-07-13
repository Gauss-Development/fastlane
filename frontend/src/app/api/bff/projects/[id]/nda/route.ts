import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { NDA } from "@/lib/design/types";

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const authorization = request.headers.get("authorization");
  if (!authorization) {
    return toFailureResponse(401, "UNAUTHORIZED", "Authentication required.");
  }

  const { id } = await params;
  const { upstream, payload } = await proxyGateway<NDA>(request, `/api/v1/projects/${encodeURIComponent(id)}/nda`, {
    method: "GET",
    headers: { authorization },
  });
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "GET_NDA_STATUS_FAILED", "Failed to retrieve NDA status.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
