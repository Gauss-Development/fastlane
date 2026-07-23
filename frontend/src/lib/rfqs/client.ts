"use client";

import { BFF_BASE_URL } from "@/lib/auth/client-constants";
import { authenticatedFetch } from "@/lib/auth/client-api";
import type {
  CreateRFQParams,
  ListQuotesResponse,
  ListRFQsResponse,
  Quote,
  RFQ,
  SubmitManufacturerQuoteParams,
  SubmitSupplierQuoteParams,
  SupplierRFQView,
} from "@/lib/rfqs/types";

export async function createRFQ(params: CreateRFQParams): Promise<RFQ> {
  return authenticatedFetch<RFQ>(`${BFF_BASE_URL}/rfqs`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      query_text: params.queryText,
      parsed_specs: params.parsedSpecs ?? null,
      matched_product_ids: params.matchedProductIds ?? [],
      project_id: params.projectId ?? "",
      qty: params.qty,
      target_date: params.targetDate ?? "",
      shipping_address: params.shippingAddress ?? "",
      notes: params.notes ?? "",
    }),
  });
}

export async function listRFQs(params?: { status?: string; limit?: number; offset?: number }): Promise<ListRFQsResponse> {
  const query = new URLSearchParams();
  if (params?.status) query.set("status", params.status);
  if (params?.limit) query.set("limit", String(params.limit));
  if (params?.offset) query.set("offset", String(params.offset));
  const suffix = query.size > 0 ? `?${query.toString()}` : "";
  return authenticatedFetch<ListRFQsResponse>(`${BFF_BASE_URL}/rfqs${suffix}`);
}

// Manufacturer board: open RFQs across all buyers (buyer email + address blanked server-side).
export async function listOpenRFQs(params?: { limit?: number; offset?: number }): Promise<ListRFQsResponse> {
  const query = new URLSearchParams();
  if (params?.limit) query.set("limit", String(params.limit));
  if (params?.offset) query.set("offset", String(params.offset));
  const suffix = query.size > 0 ? `?${query.toString()}` : "";
  return authenticatedFetch<ListRFQsResponse>(`${BFF_BASE_URL}/manufacturer-rfqs${suffix}`);
}

export async function getRFQ(id: string): Promise<RFQ> {
  return authenticatedFetch<RFQ>(`${BFF_BASE_URL}/rfqs/${encodeURIComponent(id)}`);
}

export async function listQuotes(rfqId: string): Promise<ListQuotesResponse> {
  return authenticatedFetch<ListQuotesResponse>(`${BFF_BASE_URL}/rfqs/${encodeURIComponent(rfqId)}/quotes`);
}

export async function acceptQuote(rfqId: string, quoteId: string): Promise<Quote> {
  return authenticatedFetch<Quote>(
    `${BFF_BASE_URL}/rfqs/${encodeURIComponent(rfqId)}/quotes/${encodeURIComponent(quoteId)}/accept`,
    { method: "POST" },
  );
}

export async function submitManufacturerQuote(
  rfqId: string,
  body: SubmitManufacturerQuoteParams,
): Promise<Quote> {
  return authenticatedFetch<Quote>(
    `${BFF_BASE_URL}/manufacturer-rfqs/${encodeURIComponent(rfqId)}/quote`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    },
  );
}

// Supplier magic-link surface: public, token-gated — plain fetch, no session.

async function publicJSON<T>(input: string, init?: RequestInit): Promise<T> {
  const response = await fetch(input, init);
  const payload = (await response.json().catch(() => null)) as
    | { success: boolean; data?: T; error?: { code: string; message: string } }
    | null;
  if (!response.ok || !payload?.success || payload.data === undefined) {
    throw new Error(payload?.error?.message || "Request failed.");
  }
  return payload.data;
}

export async function getSupplierRFQ(token: string): Promise<SupplierRFQView> {
  return publicJSON<SupplierRFQView>(`${BFF_BASE_URL}/supplier-rfq/${encodeURIComponent(token)}`);
}

export async function submitSupplierQuote(token: string, params: SubmitSupplierQuoteParams): Promise<Quote> {
  return publicJSON<Quote>(`${BFF_BASE_URL}/supplier-rfq/${encodeURIComponent(token)}/quote`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      price_usd: params.priceUsd,
      lead_time_days: params.leadTimeDays,
      validity_date: params.validityDate ?? "",
      notes: params.notes ?? "",
    }),
  });
}
