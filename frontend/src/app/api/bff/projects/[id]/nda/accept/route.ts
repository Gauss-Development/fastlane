import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { NDA } from "@/lib/design/types";

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
  const { upstream, payload } = await proxyGateway<NDA>(request, `/api/v1/projects/${encodeURIComponent(id)}/nda/accept`, {
    method: "POST",
    headers: { authorization },
    body: JSON.stringify(rawBody),
  });
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "ACCEPT_NDA_FAILED", "Failed to accept NDA.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
