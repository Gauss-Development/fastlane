"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { ArrowRight, Search } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
} from "@/components/ui/card";
import { CodeId } from "@/components/ui/code-id";
import { StatusPill, type StatusTone } from "@/components/ui/pill";
import { Table, TableBody, Td, Th, TableHead, Tr } from "@/components/ui/table";
import { listOpenRFQs, listRFQs } from "@/lib/rfqs/client";
import type { RFQStatus } from "@/lib/rfqs/types";
import { listOrders } from "@/lib/orders/client";
import { useAuthStore } from "@/lib/stores/auth-store";

const RFQ_TONE: Record<RFQStatus, StatusTone> = {
  open: "info",
  quoted: "warning",
  accepted: "success",
  closed: "neutral",
};

// Backend order statuses are richer than the frontend union — key by string
// with a neutral fallback so unmapped states still render.
const ORDER_TONE: Record<string, StatusTone> = {
  pending_payment: "warning",
  paid: "info",
  in_production: "info",
  ready_for_qc: "info",
  qc_in_progress: "info",
  qc_passed: "success",
  qc_failed: "destructive",
  shipped_from_cn: "info",
  in_transit: "info",
  out_for_delivery: "info",
  delivered: "success",
  completed: "success",
  cancelled: "destructive",
  refunded: "destructive",
  disputed: "warning",
};

function humanize(status: string) {
  return status.replace(/_/g, " ");
}

function age(createdAt: string) {
  const then = new Date(createdAt).getTime();
  if (Number.isNaN(then)) return "—";
  const mins = Math.max(0, Math.round((Date.now() - then) / 60000));
  if (mins < 60) return `${mins}m`;
  const hours = Math.round(mins / 60);
  if (hours < 24) return `${hours}h`;
  return `${Math.round(hours / 24)}d`;
}

function usd(value: number) {
  return `$${Math.round(value).toLocaleString("en-US")}`;
}

function StatGrid({ items }: { items: { label: string; value: string | number }[] }) {
  return (
    <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
      {items.map((item) => (
        <div key={item.label} className="border border-border bg-card p-4">
          <div className="font-mono text-3xl tabular-nums text-foreground">{item.value}</div>
          <div className="mt-1 font-mono text-[11px] uppercase tracking-[0.1em] text-muted-foreground">
            {item.label}
          </div>
        </div>
      ))}
    </div>
  );
}

function SectionHeader({ title, href, cta }: { title: string; href: string; cta: string }) {
  return (
    <div className="mb-3 flex items-end justify-between gap-3">
      <h2 className="text-lg">{title}</h2>
      <Link
        href={href}
        className="font-mono text-xs uppercase tracking-[0.08em] text-primary hover:underline"
      >
        {cta} →
      </Link>
    </div>
  );
}

function Shell({
  eyebrow,
  title,
  description,
  action,
  children,
}: {
  eyebrow: string;
  title: string;
  description: string;
  action?: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <main className="mx-auto w-full max-w-[1200px] space-y-6 px-6 py-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="mb-1 font-mono text-xs uppercase tracking-[0.16em] text-muted-foreground">
            {eyebrow}
          </p>
          <h1 className="text-2xl">{title}</h1>
          <p className="mt-1 max-w-2xl text-sm text-muted-foreground">{description}</p>
        </div>
        {action}
      </div>
      {children}
    </main>
  );
}

function BuyerDashboard({ email }: { email?: string }) {
  const rfqsQuery = useQuery({
    queryKey: ["rfqs", "dashboard"],
    queryFn: () => listRFQs({ limit: 50 }),
  });
  const ordersQuery = useQuery({
    queryKey: ["orders", "dashboard"],
    queryFn: () => listOrders({ limit: 50 }),
  });

  const rfqs = rfqsQuery.data?.rfqs ?? [];
  const orders = ordersQuery.data?.orders ?? [];
  const openCount = rfqs.filter((r) => r.status === "open").length;
  const quotedCount = rfqs.filter((r) => r.status === "quoted").length;

  return (
    <Shell
      eyebrow="Buyer workspace"
      title="Dashboard"
      description={`Sourcing overview${email ? ` for ${email}` : ""}. Start from search, track quotes, follow orders to delivery.`}
      action={
        <div className="flex gap-2">
          <Button asChild size="lg" variant="outline" className="font-mono uppercase tracking-[0.08em]">
            <Link href="/rfqs/new">New request</Link>
          </Button>
          <Button asChild size="lg" className="font-mono uppercase tracking-[0.08em]">
            <Link href="/search">
              <Search className="size-4" /> New search
            </Link>
          </Button>
        </div>
      }
    >
      <StatGrid
        items={[
          { label: "Open RFQs", value: openCount },
          { label: "Awaiting decision", value: quotedCount },
          { label: "RFQs total", value: rfqsQuery.data?.total ?? rfqs.length },
          { label: "Orders", value: ordersQuery.data?.total ?? orders.length },
        ]}
      />

      <section>
        <SectionHeader title="Recent RFQs" href="/rfqs" cta="All RFQs" />
        <Card>
          <CardContent className="p-0">
            {rfqs.length === 0 ? (
              <div className="py-10 text-center text-sm text-muted-foreground">
                No RFQs yet. Run a{" "}
                <Link href="/search" className="text-primary hover:underline">
                  search
                </Link>{" "}
                and request a quote on a match.
              </div>
            ) : (
              <Table>
                <TableHead>
                  <Tr>
                    <Th>RFQ</Th>
                    <Th>Request</Th>
                    <Th numeric>Qty</Th>
                    <Th>Status</Th>
                    <Th numeric>Age</Th>
                  </Tr>
                </TableHead>
                <TableBody>
                  {rfqs.slice(0, 6).map((rfq) => (
                    <Tr key={rfq.id}>
                      <Td>
                        <Link href={`/rfqs/${encodeURIComponent(rfq.id)}`} className="hover:underline">
                          <CodeId code={rfq.id} size="sm" />
                        </Link>
                      </Td>
                      <Td className="max-w-[360px] truncate font-mono text-xs">{rfq.query_text}</Td>
                      <Td numeric>{rfq.qty || "—"}</Td>
                      <Td>
                        <StatusPill tone={RFQ_TONE[rfq.status]}>{rfq.status}</StatusPill>
                      </Td>
                      <Td numeric>{age(rfq.created_at)}</Td>
                    </Tr>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      </section>

      <section>
        <SectionHeader title="Recent orders" href="/orders" cta="All orders" />
        <Card>
          <CardContent className="p-0">
            {orders.length === 0 ? (
              <div className="py-10 text-center text-sm text-muted-foreground">
                No orders yet. Accept a quote on an RFQ to create one.
              </div>
            ) : (
              <Table>
                <TableHead>
                  <Tr>
                    <Th>Order</Th>
                    <Th>Status</Th>
                    <Th numeric>Total</Th>
                    <Th numeric>Age</Th>
                  </Tr>
                </TableHead>
                <TableBody>
                  {orders.slice(0, 6).map((order) => (
                    <Tr key={order.id}>
                      <Td>
                        <Link href={`/orders/${encodeURIComponent(order.id)}`} className="hover:underline">
                          <CodeId code={order.id} size="sm" />
                        </Link>
                      </Td>
                      <Td>
                        <StatusPill tone={ORDER_TONE[order.status] ?? "neutral"}>
                          {humanize(order.status)}
                        </StatusPill>
                      </Td>
                      <Td numeric>{usd(order.total_usd)}</Td>
                      <Td numeric>{age(order.created_at)}</Td>
                    </Tr>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      </section>
    </Shell>
  );
}

function ManufacturerDashboard({ email }: { email?: string }) {
  const openRfqsQuery = useQuery({
    queryKey: ["rfqs", "open", "dashboard"],
    queryFn: () => listOpenRFQs({ limit: 50 }),
  });
  const ordersQuery = useQuery({
    queryKey: ["orders", "dashboard"],
    queryFn: () => listOrders({ limit: 50 }),
  });

  const openRfqs = openRfqsQuery.data?.rfqs ?? [];
  const orders = ordersQuery.data?.orders ?? [];

  return (
    <Shell
      eyebrow="Supplier workspace"
      title="Dashboard"
      description={`Incoming demand${email ? ` for ${email}` : ""}. Quote open requests and fulfil accepted orders.`}
      action={
        <Button asChild size="lg" variant="outline" className="font-mono uppercase tracking-[0.08em]">
          <Link href="/rfqs">
            Browse open RFQs <ArrowRight className="size-4" />
          </Link>
        </Button>
      }
    >
      <StatGrid
        items={[
          { label: "Open RFQs", value: openRfqsQuery.data?.total ?? openRfqs.length },
          { label: "My orders", value: ordersQuery.data?.total ?? orders.length },
        ]}
      />

      <section>
        <SectionHeader title="Open quote requests" href="/rfqs" cta="Open RFQs" />
        <Card>
          <CardContent className="p-0">
            {openRfqs.length === 0 ? (
              <div className="py-10 text-center text-sm text-muted-foreground">
                No open RFQs right now. New quote requests from buyers appear here.
              </div>
            ) : (
              <Table>
                <TableHead>
                  <Tr>
                    <Th>RFQ</Th>
                    <Th>Request</Th>
                    <Th numeric>Qty</Th>
                    <Th numeric>Age</Th>
                  </Tr>
                </TableHead>
                <TableBody>
                  {openRfqs.slice(0, 6).map((rfq) => (
                    <Tr key={rfq.id}>
                      <Td>
                        <CodeId code={rfq.id} size="sm" />
                      </Td>
                      <Td className="max-w-[360px] truncate font-mono text-xs">{rfq.query_text}</Td>
                      <Td numeric>{rfq.qty || "—"}</Td>
                      <Td numeric>{age(rfq.created_at)}</Td>
                    </Tr>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      </section>

      <section>
        <SectionHeader title="My orders" href="/orders" cta="All orders" />
        <Card>
          <CardContent className="p-0">
            {orders.length === 0 ? (
              <div className="py-10 text-center text-sm text-muted-foreground">
                No orders yet. Accepted quotes become orders here.
              </div>
            ) : (
              <Table>
                <TableHead>
                  <Tr>
                    <Th>Order</Th>
                    <Th>Status</Th>
                    <Th numeric>Total</Th>
                    <Th numeric>Age</Th>
                  </Tr>
                </TableHead>
                <TableBody>
                  {orders.slice(0, 6).map((order) => (
                    <Tr key={order.id}>
                      <Td>
                        <Link href={`/orders/${encodeURIComponent(order.id)}`} className="hover:underline">
                          <CodeId code={order.id} size="sm" />
                        </Link>
                      </Td>
                      <Td>
                        <StatusPill tone={ORDER_TONE[order.status] ?? "neutral"}>
                          {humanize(order.status)}
                        </StatusPill>
                      </Td>
                      <Td numeric>{usd(order.total_usd)}</Td>
                      <Td numeric>{age(order.created_at)}</Td>
                    </Tr>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      </section>
    </Shell>
  );
}

export function DashboardClient() {
  const user = useAuthStore((state) => state.user);
  if (user?.role === "manufacturer") {
    return <ManufacturerDashboard email={user?.email} />;
  }
  return <BuyerDashboard email={user?.email} />;
}
