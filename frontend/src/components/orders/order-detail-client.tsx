"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { CodeId } from "@/components/ui/code-id";
import { RouteIndicator } from "@/components/ui/route-indicator";
import { StatusPill, type StatusTone } from "@/components/ui/pill";
import { getOrder, listOrderEvents } from "@/lib/orders/client";
import type { OrderStatus } from "@/lib/orders/types";

const ORDER_TONE: Record<OrderStatus, StatusTone> = {
  pending: "neutral",
  confirmed: "info",
  in_production: "warning",
  qc: "warning",
  shipped: "info",
  delivered: "success",
  cancelled: "destructive",
};

function formatMoney(value: number) {
  if (!Number.isFinite(value) || value <= 0) return "—";
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" }).format(value);
}

function relativeTime(ts: string) {
  const ms = new Date(ts).getTime();
  if (!Number.isFinite(ms)) return "—";
  const minutes = Math.max(0, Math.floor((Date.now() - ms) / 60000));
  if (minutes < 1) return "just now";
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 48) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

export function OrderDetailClient({ orderId }: { orderId: string }) {
  const orderQuery = useQuery({
    queryKey: ["order", orderId],
    queryFn: () => getOrder(orderId),
  });

  const eventsQuery = useQuery({
    queryKey: ["order-events", orderId],
    queryFn: () => listOrderEvents(orderId),
    refetchInterval: 30_000,
  });

  const order = orderQuery.data;

  return (
    <main className="mx-auto flex w-full max-w-[1200px] flex-col gap-6 px-6 py-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.18em] text-muted-foreground">Order Detail</p>
          <div className="mt-2">
            <CodeId code={orderId} size="lg" copyable />
          </div>
        </div>
        <RouteIndicator size="sm" />
      </div>

      {orderQuery.isLoading ? (
        <Card>
          <CardContent className="py-10 text-center font-mono text-sm text-muted-foreground">
            Loading order…
          </CardContent>
        </Card>
      ) : null}

      {orderQuery.error ? (
        <Card className="border-destructive/50">
          <CardHeader>
            <CardTitle>Could not load order</CardTitle>
            <CardDescription>{(orderQuery.error as Error).message}</CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/orders" className="font-mono text-xs uppercase tracking-[0.08em] text-primary hover:underline">
              ← Back to Orders
            </Link>
          </CardContent>
        </Card>
      ) : null}

      {order ? (
        <Card>
          <CardHeader>
            <div className="flex flex-wrap items-center justify-between gap-3">
              <CardTitle className="font-mono text-base">Order Summary</CardTitle>
              <StatusPill tone={ORDER_TONE[order.status] ?? "neutral"}>{order.status}</StatusPill>
            </div>
          </CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {[
              ["Payment", order.payment_status],
              ["QC", order.qc_status],
              ["Total", formatMoney(order.total_usd)],
              ["Ship to", [order.shipping_city, order.shipping_country].filter(Boolean).join(", ") || order.shipping_address || "—"],
              ["RFQ", null],
              ["Quote", null],
              ["Created", order.created_at ? new Date(order.created_at).toLocaleString() : "—"],
              ["Warranty until", order.warranty_until || "—"],
            ].map(([label, value]) => {
              if (label === "RFQ") {
                return (
                  <div key="rfq">
                    <div className="font-mono text-[11px] uppercase tracking-[0.08em] text-muted-foreground">RFQ</div>
                    <div className="mt-1">
                      <Link href={`/rfqs/${encodeURIComponent(order.rfq_id)}`} className="hover:underline">
                        <CodeId code={order.rfq_id} size="sm" />
                      </Link>
                    </div>
                  </div>
                );
              }
              if (label === "Quote") {
                return (
                  <div key="quote">
                    <div className="font-mono text-[11px] uppercase tracking-[0.08em] text-muted-foreground">Quote</div>
                    <div className="mt-1">
                      <CodeId code={order.quote_id} size="sm" />
                    </div>
                  </div>
                );
              }
              return (
                <div key={label as string}>
                  <div className="font-mono text-[11px] uppercase tracking-[0.08em] text-muted-foreground">{label}</div>
                  <div className="mt-1 font-mono text-sm">{value as string}</div>
                </div>
              );
            })}
          </CardContent>
        </Card>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle>Timeline</CardTitle>
          <CardDescription>
            Order events ordered by time. Refreshes every 30 seconds.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {eventsQuery.isLoading ? (
            <p className="py-6 text-center font-mono text-sm text-muted-foreground">Loading timeline…</p>
          ) : null}

          {eventsQuery.data && eventsQuery.data.events.length === 0 ? (
            <p className="py-6 text-center text-sm text-muted-foreground">No events recorded yet.</p>
          ) : null}

          {eventsQuery.data && eventsQuery.data.events.length > 0 ? (
            <ol className="relative border-l border-border pl-6">
              {[...eventsQuery.data.events]
                .sort((a, b) => new Date(a.occurred_at).getTime() - new Date(b.occurred_at).getTime())
                .map((event) => (
                  <li key={event.id} className="mb-6 last:mb-0">
                    <span className="absolute -left-1.5 mt-1 size-3 rounded-full border border-border bg-card" aria-hidden />
                    <div className="flex flex-wrap items-baseline gap-2">
                      <span className="font-mono text-xs font-medium uppercase tracking-[0.08em]">
                        {event.to_status || event.event_type}
                      </span>
                      <span className="font-mono text-[10px] text-muted-foreground">
                        {relativeTime(event.occurred_at)}
                      </span>
                      {event.actor_type ? (
                        <span className="rounded-sm border border-border px-1.5 py-0.5 font-mono text-[10px] uppercase tracking-[0.06em] text-muted-foreground">
                          {event.actor_type}
                        </span>
                      ) : null}
                    </div>
                    {event.notes ? (
                      <p className="mt-1 text-sm text-muted-foreground">{event.notes}</p>
                    ) : null}
                    {event.location ? (
                      <p className="mt-0.5 font-mono text-[11px] text-muted-foreground">{event.location}</p>
                    ) : null}
                  </li>
                ))}
            </ol>
          ) : null}
        </CardContent>
      </Card>
    </main>
  );
}
