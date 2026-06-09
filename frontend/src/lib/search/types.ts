export interface ParsedSpecs {
  data_rate?: string;
  form_factor?: string;
  reach_km?: number;
  wavelength_nm?: number;
  compatibility?: string[];
  fiber_type?: string;
  qty_estimated?: number;
  free_text?: string;
}

export type ProductSpecs = Record<string, unknown>;

export interface ProductHit {
  id: string;
  supplier_id: string;
  sku: string;
  name: string;
  name_zh: string;
  category: string;
  specs: ProductSpecs | null;
  price_usd: number;
  moq: number;
  stock_qty: number;
  lead_time_days: number;
  datasheet_url: string;
  supplier_name: string;
  supplier_city: string;
  supplier_verified: boolean;
  match_score: number;
  vector_distance: number;
  match_explanation: string;
}

export interface SearchResponse {
  parsed_specs: ParsedSpecs | null;
  results: ProductHit[];
  query_id: string;
}

export interface SearchParams {
  query: string;
  limit?: number;
  specOverrides?: ParsedSpecs | null;
}
