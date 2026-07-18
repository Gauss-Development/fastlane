import type { ParsedSpecs } from "@/lib/search/types";

export type RFQStatus = "open" | "quoted" | "accepted" | "closed";
export type QuoteStatus = "pending" | "submitted" | "accepted" | "rejected";

export interface RFQ {
  id: string;
  project_id?: string;
  buyer_id?: string;
  buyer_email?: string;
  buyer_company: string;
  query_text: string;
  parsed_specs: Record<string, unknown> | null;
  matched_product_ids: string[];
  status: RFQStatus;
  qty: number;
  target_date: string;
  shipping_address: string;
  notes: string;
  created_at: string;
}

export interface Quote {
  id: string;
  rfq_id: string;
  supplier_id: string;
  manufacturer_id?: string;
  product_id: string;
  price_usd: number;
  lead_time_days: number;
  validity_date: string;
  supplier_notes: string;
  match_score: number;
  status: QuoteStatus;
  submitted_at: string;
  created_at: string;
}

export interface SubmitManufacturerQuoteParams {
  price_usd: number;
  lead_time_days: number;
  validity_date: string;
  notes: string;
  product_id?: string;
}

export interface CreateRFQParams {
  queryText: string;
  parsedSpecs?: ParsedSpecs | null;
  matchedProductIds: string[];
  qty: number;
  targetDate?: string;
  shippingAddress?: string;
  notes?: string;
}

export interface ListRFQsResponse {
  rfqs: RFQ[];
  total: number;
  limit: number;
  offset: number;
}

export interface ListQuotesResponse {
  quotes: Quote[];
}

/** Supplier magic-link page payload (public, token-gated). */
export interface SupplierRFQView {
  rfq: RFQ;
  quote: Quote | null;
  supplier_name: string;
  supplier_id: string;
}

export interface SubmitSupplierQuoteParams {
  priceUsd: number;
  leadTimeDays: number;
  validityDate?: string;
  notes?: string;
}
