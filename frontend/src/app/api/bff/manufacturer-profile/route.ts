import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { Manufacturer } from "@/lib/manufacturers/types";

export async function GET(request: Request) {
  const authorization = request.headers.get("authorization");
  if (!authorization) {
    return toFailureResponse(401, "UNAUTHORIZED", "Authentication required.");
  }

  const { upstream, payload } = await proxyGateway<Manufacturer>(request, "/api/v1/manufacturer-profile", {
    method: "GET",
    headers: { authorization },
  });
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "GET_MANUFACTURER_PROFILE_FAILED", "Failed to retrieve manufacturer profile.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
