"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { CodeId } from "@/components/ui/code-id";
import { StatusPill, type StatusTone } from "@/components/ui/pill";
import { Table, TableBody, Td, Th, TableHead, Tr } from "@/components/ui/table";
import { listOrders } from "@/lib/orders/client";
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

function age(ts: string) {
  const ms = new Date(ts).getTime();
  if (!Number.isFinite(ms)) return "—";
  const minutes = Math.max(0, Math.floor((Date.now() - ms) / 60000));
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  if (hours < 48) return `${hours}h`;
  return `${Math.floor(hours / 24)}d`;
}

export function OrdersClient() {
  const ordersQuery = useQuery({
    queryKey: ["orders"],
    queryFn: () => listOrders({ limit: 50 }),
  });

  return (
    <main className="mx-auto w-full max-w-[1200px] px-6 py-6">
      <Card>
        <CardHeader>
          <CardTitle>Orders</CardTitle>
        </CardHeader>
        <CardContent>
          {ordersQuery.isLoading ? (
            <p className="py-8 text-center font-mono text-sm text-muted-foreground">Loading orders…</p>
          ) : null}

          {ordersQuery.error ? (
            <p className="py-8 text-center text-sm text-destructive">
              {(ordersQuery.error as Error).message}
            </p>
          ) : null}

          {ordersQuery.data && ordersQuery.data.orders.length === 0 ? (
            <p className="py-10 text-center text-sm text-muted-foreground">
              No orders yet. Accept a quote on an RFQ to create one.
            </p>
          ) : null}

          {ordersQuery.data && ordersQuery.data.orders.length > 0 ? (
            <Table>
              <TableHead>
                <Tr>
                  <Th>Order</Th>
                  <Th>RFQ</Th>
                  <Th>Status</Th>
                  <Th numeric>Total</Th>
                  <Th numeric>Age</Th>
                </Tr>
              </TableHead>
              <TableBody>
                {ordersQuery.data.orders.map((order) => (
                  <Tr key={order.id}>
                    <Td>
                      <Link href={`/orders/${encodeURIComponent(order.id)}`} className="hover:underline">
                        <CodeId code={order.id} size="sm" />
                      </Link>
                    </Td>
                    <Td>
                      <Link href={`/rfqs/${encodeURIComponent(order.rfq_id)}`} className="hover:underline">
                        <CodeId code={order.rfq_id} size="sm" />
                      </Link>
                    </Td>
                    <Td>
                      <StatusPill tone={ORDER_TONE[order.status] ?? "neutral"}>{order.status}</StatusPill>
                    </Td>
                    <Td numeric className="font-mono">{formatMoney(order.total_usd)}</Td>
                    <Td numeric>{age(order.created_at)}</Td>
                  </Tr>
                ))}
              </TableBody>
            </Table>
          ) : null}
        </CardContent>
      </Card>
    </main>
  );
}
