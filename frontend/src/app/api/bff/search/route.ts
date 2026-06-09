import { proxyGateway, toFailureResponse, toSuccessResponse } from "@/lib/server/gateway";
import type { ParsedSpecs, SearchResponse } from "@/lib/search/types";

interface SearchRequestBody {
  query?: unknown;
  limit?: unknown;
  spec_overrides?: ParsedSpecs | null;
}

export async function POST(request: Request) {
  const authorization = request.headers.get("authorization");
  if (!authorization) {
    return toFailureResponse(401, "UNAUTHORIZED", "Authentication required.");
  }

  let body: SearchRequestBody;
  try {
    body = (await request.json()) as SearchRequestBody;
  } catch {
    return toFailureResponse(400, "INVALID_BODY", "Request body must be valid JSON.");
  }

  const query = typeof body.query === "string" ? body.query.trim() : "";
  if (!query) {
    return toFailureResponse(400, "MISSING_QUERY", "Search query is required.");
  }

  const rawLimit = typeof body.limit === "number" && Number.isFinite(body.limit)
    ? Math.trunc(body.limit)
    : 20;
  const limit = Math.min(Math.max(rawLimit, 1), 50);

  const { upstream, payload } = await proxyGateway<SearchResponse>(request, "/api/v1/search", {
    method: "POST",
    headers: { authorization },
    body: JSON.stringify({
      query,
      limit,
      spec_overrides: body.spec_overrides ?? null,
    }),
  });

  if (!payload?.success || payload.data === undefined) {
    return toFailureResponse(
      upstream.status,
      "SEARCH_FAILED",
      "Search failed.",
      payload?.error,
    );
  }

  return toSuccessResponse(payload.data, upstream.status);
}
