import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { DesignFile } from "@/lib/design/types";

export async function POST(
  request: Request,
  { params }: { params: Promise<{ fileId: string }> },
) {
  const authorization = request.headers.get("authorization");
  if (!authorization) {
    return toFailureResponse(401, "UNAUTHORIZED", "Authentication required.");
  }

  const rawBody = await request.json().catch(() => null);
  if (!rawBody || typeof rawBody !== "object") {
    return toFailureResponse(400, "INVALID_REQUEST", "Invalid request body.");
  }

  const { fileId } = await params;
  const { upstream, payload } = await proxyGateway<DesignFile>(request, `/api/v1/files/${encodeURIComponent(fileId)}/confirm`, {
    method: "POST",
    headers: { authorization },
    body: JSON.stringify(rawBody),
  });
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "CONFIRM_UPLOAD_FAILED", "Failed to confirm upload.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
