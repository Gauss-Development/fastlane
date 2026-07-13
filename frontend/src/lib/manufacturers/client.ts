"use client";

import { BFF_BASE_URL } from "@/lib/auth/client-constants";
import { authenticatedFetch } from "@/lib/auth/client-api";
import type { ListManufacturersResponse, Manufacturer } from "@/lib/manufacturers/types";

export async function listManufacturers(filters?: {
  cluster?: string;
  service_type?: string;
  assembly_type?: string;
  material?: string;
  verified_only?: boolean;
  min_layers_gte?: number;
  limit?: number;
  offset?: number;
}): Promise<ListManufacturersResponse> {
  const query = new URLSearchParams();
  if (filters?.cluster) query.set("cluster", filters.cluster);
  if (filters?.service_type) query.set("service_type", filters.service_type);
  if (filters?.assembly_type) query.set("assembly_type", filters.assembly_type);
  if (filters?.material) query.set("material", filters.material);
  if (filters?.verified_only) query.set("verified_only", "true");
  if (filters?.min_layers_gte) query.set("min_layers_gte", String(filters.min_layers_gte));
  if (filters?.limit) query.set("limit", String(filters.limit));
  if (filters?.offset) query.set("offset", String(filters.offset));
  const suffix = query.size > 0 ? `?${query.toString()}` : "";
  return authenticatedFetch<ListManufacturersResponse>(`${BFF_BASE_URL}/manufacturers${suffix}`);
}

export async function getManufacturer(id: string): Promise<Manufacturer> {
  return authenticatedFetch<Manufacturer>(`${BFF_BASE_URL}/manufacturers/${encodeURIComponent(id)}`);
}

export async function getMyManufacturer(): Promise<Manufacturer> {
  return authenticatedFetch<Manufacturer>(`${BFF_BASE_URL}/manufacturer-profile`);
}

export async function createManufacturer(body: Partial<Manufacturer>): Promise<Manufacturer> {
  return authenticatedFetch<Manufacturer>(`${BFF_BASE_URL}/manufacturers`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
}

export async function updateManufacturer(
  id: string,
  body: Partial<Manufacturer>,
): Promise<Manufacturer> {
  return authenticatedFetch<Manufacturer>(`${BFF_BASE_URL}/manufacturers/${encodeURIComponent(id)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
}
