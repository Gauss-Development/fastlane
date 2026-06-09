"use client";

import { BFF_BASE_URL } from "@/lib/auth/client-constants";
import { authenticatedFetch } from "@/lib/auth/client-api";
import type { SearchParams, SearchResponse } from "@/lib/search/types";

export async function search(params: SearchParams): Promise<SearchResponse> {
  const query = params.query.trim();
  if (!query) {
    throw new Error("Search query is required.");
  }
  return authenticatedFetch<SearchResponse>(`${BFF_BASE_URL}/search`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      query,
      limit: params.limit ?? 20,
      spec_overrides: params.specOverrides ?? null,
    }),
  });
}

export type {
  ParsedSpecs,
  ProductHit,
  ProductSpecs,
  SearchParams,
  SearchResponse,
} from "@/lib/search/types";
