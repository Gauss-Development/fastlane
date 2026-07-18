import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { Order } from "@/lib/orders/types";

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const authorization = request.headers.get("authorization");
  if (!authorization) {
    return toFailureResponse(401, "UNAUTHORIZED", "Authentication required.");
  }

  const { id } = await params;
  const { upstream, payload } = await proxyGateway<Order>(request, `/api/v1/orders/${encodeURIComponent(id)}`, {
    method: "GET",
    headers: { authorization },
  });
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "GET_ORDER_FAILED", "Failed to retrieve order.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
