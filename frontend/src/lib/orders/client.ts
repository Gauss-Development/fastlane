"use client";

import { BFF_BASE_URL } from "@/lib/auth/client-constants";
import { authenticatedFetch } from "@/lib/auth/client-api";
import type {
  AppendOrderEventBody,
  ListOrderEventsResponse,
  ListOrdersResponse,
  Order,
  OrderEvent,
} from "@/lib/orders/types";

export async function listOrders(params?: {
  status?: string;
  limit?: number;
  offset?: number;
}): Promise<ListOrdersResponse> {
  const query = new URLSearchParams();
  if (params?.status) query.set("status", params.status);
  if (params?.limit) query.set("limit", String(params.limit));
  if (params?.offset) query.set("offset", String(params.offset));
  const suffix = query.size > 0 ? `?${query.toString()}` : "";
  return authenticatedFetch<ListOrdersResponse>(`${BFF_BASE_URL}/orders${suffix}`);
}

export async function getOrder(id: string): Promise<Order> {
  return authenticatedFetch<Order>(`${BFF_BASE_URL}/orders/${encodeURIComponent(id)}`);
}

export async function listOrderEvents(id: string): Promise<ListOrderEventsResponse> {
  return authenticatedFetch<ListOrderEventsResponse>(
    `${BFF_BASE_URL}/orders/${encodeURIComponent(id)}/events`,
  );
}

export async function appendOrderEvent(id: string, body: AppendOrderEventBody): Promise<OrderEvent> {
  return authenticatedFetch<OrderEvent>(
    `${BFF_BASE_URL}/orders/${encodeURIComponent(id)}/events`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    },
  );
}
