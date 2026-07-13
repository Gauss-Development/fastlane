import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { DownloadURLResponse } from "@/lib/design/types";

export async function GET(
  request: Request,
  { params }: { params: Promise<{ fileId: string }> },
) {
  const authorization = request.headers.get("authorization");
  if (!authorization) {
    return toFailureResponse(401, "UNAUTHORIZED", "Authentication required.");
  }

  const { fileId } = await params;
  const { upstream, payload } = await proxyGateway<DownloadURLResponse>(request, `/api/v1/files/${encodeURIComponent(fileId)}/download-url`, {
    method: "GET",
    headers: { authorization },
  });
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "REQUEST_DOWNLOAD_URL_FAILED", "Failed to request download URL.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
