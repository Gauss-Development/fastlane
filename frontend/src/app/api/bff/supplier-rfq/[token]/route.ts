import {
  proxyGateway,
  toFailureResponse,
  toSuccessResponse,
} from "@/lib/server/gateway";
import type { SupplierRFQView } from "@/lib/rfqs/types";

// Public: the magic-link token in the path is the only credential.
export async function GET(
  request: Request,
  { params }: { params: Promise<{ token: string }> },
) {
  const { token } = await params;
  const { upstream, payload } = await proxyGateway<SupplierRFQView>(
    request,
    `/api/v1/supplier-rfq/${encodeURIComponent(token)}`,
    { method: "GET" },
  );
  if (!payload?.success || !payload.data) {
    return toFailureResponse(upstream.status, "GET_SUPPLIER_RFQ_FAILED", "This link is invalid or has expired.", payload?.error);
  }
  return toSuccessResponse(payload.data, upstream.status);
}
