"use client";

import Link from "next/link";
import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { CodeId } from "@/components/ui/code-id";
import { RouteIndicator } from "@/components/ui/route-indicator";
import { StatusPill, type StatusTone } from "@/components/ui/pill";
import { Table, TableBody, Td, Th, TableHead, Tr } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { acceptQuote, getRFQ, listQuotes } from "@/lib/rfqs/client";
import type { QuoteStatus, RFQStatus } from "@/lib/rfqs/types";

const RFQ_TONE: Record<RFQStatus, StatusTone> = {
  open: "info",
  quoted: "warning",
  accepted: "success",
  closed: "neutral",
};

const QUOTE_TONE: Record<QuoteStatus, StatusTone> = {
  pending: "neutral",
  submitted: "info",
  accepted: "success",
  rejected: "destructive",
};

function formatMoney(value: number) {
  if (!Number.isFinite(value) || value <= 0) return "—";
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" }).format(value);
}

function formatSpecs(specs: Record<string, unknown> | null) {
  if (!specs) return [];
  return Object.entries(specs).filter(([, value]) => {
    if (Array.isArray(value)) return value.length > 0;
    return value !== null && value !== undefined && value !== "" && value !== 0;
  });
}

export function RFQDetailClient({ rfqId }: { rfqId: string }) {
  const queryClient = useQueryClient();
  const [accepting, setAccepting] = useState<string | null>(null);
  const [acceptedQuoteId, setAcceptedQuoteId] = useState<string | null>(null);
  const [acceptError, setAcceptError] = useState<string | null>(null);

  const rfqQuery = useQuery({
    queryKey: ["rfq", rfqId],
    queryFn: () => getRFQ(rfqId),
  });
  const quotesQuery = useQuery({
    queryKey: ["rfq-quotes", rfqId],
    queryFn: () => listQuotes(rfqId),
    // Quotes arrive asynchronously as suppliers respond; poll while open.
    refetchInterval: 30_000,
  });

  const rfq = rfqQuery.data;
  const isAccepted = rfq?.status === "accepted" || rfq?.status === "closed";
  const alreadyAcceptedId =
    acceptedQuoteId ??
    quotesQuery.data?.quotes.find((q) => q.status === "accepted")?.id ??
    null;

  async function handleAccept(quoteId: string) {
    setAccepting(quoteId);
    setAcceptError(null);
    try {
      await acceptQuote(rfqId, quoteId);
      setAcceptedQuoteId(quoteId);
      await queryClient.invalidateQueries({ queryKey: ["rfq", rfqId] });
      await queryClient.invalidateQueries({ queryKey: ["rfq-quotes", rfqId] });
    } catch (err) {
      setAcceptError((err as Error).message);
    } finally {
      setAccepting(null);
    }
  }

  const showAcceptSuccess = alreadyAcceptedId && (acceptedQuoteId !== null || isAccepted);

  return (
    <main className="mx-auto flex w-full max-w-[1200px] flex-col gap-6 px-6 py-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.18em] text-muted-foreground">RFQ Detail</p>
          <div className="mt-2">
            <CodeId code={rfqId} size="lg" copyable />
          </div>
        </div>
        <RouteIndicator size="sm" />
      </div>

      {rfqQuery.isLoading ? (
        <Card>
          <CardContent className="py-10 text-center font-mono text-sm text-muted-foreground">
            Loading RFQ…
          </CardContent>
        </Card>
      ) : null}

      {rfqQuery.error ? (
        <Card className="border-destructive/50">
          <CardHeader>
            <CardTitle>Could not load RFQ</CardTitle>
            <CardDescription>{(rfqQuery.error as Error).message}</CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/rfqs" className="font-mono text-xs uppercase tracking-[0.08em] text-primary hover:underline">
              ← Back to RFQs
            </Link>
          </CardContent>
        </Card>
      ) : null}

      {rfq ? (
        <Card>
          <CardHeader>
            <div className="flex flex-wrap items-center justify-between gap-3">
              <CardTitle className="font-mono text-base">
                <span className="text-primary">&gt;</span> {rfq.query_text}
              </CardTitle>
              <StatusPill tone={RFQ_TONE[rfq.status] ?? "neutral"}>{rfq.status}</StatusPill>
            </div>
          </CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {[
              ["Quantity", rfq.qty ? rfq.qty.toLocaleString() : "—"],
              ["Target date", rfq.target_date || "—"],
              ["Ship to", rfq.shipping_address || "—"],
              ["Created", rfq.created_at ? new Date(rfq.created_at).toLocaleString() : "—"],
            ].map(([label, value]) => (
              <div key={label as string}>
                <div className="font-mono text-[11px] uppercase tracking-[0.08em] text-muted-foreground">{label}</div>
                <div className="mt-1 font-mono text-sm">{value}</div>
              </div>
            ))}
            {formatSpecs(rfq.parsed_specs).length > 0 ? (
              <div className="sm:col-span-2 lg:col-span-4">
                <div className="font-mono text-[11px] uppercase tracking-[0.08em] text-muted-foreground">Extracted specs</div>
                <div className="mt-2 flex flex-wrap gap-2">
                  {formatSpecs(rfq.parsed_specs).map(([key, value]) => (
                    <span key={key} className="rounded-sm border border-border bg-secondary px-2 py-1 font-mono text-xs">
                      {key}: {Array.isArray(value) ? value.join(", ") : String(value)}
                    </span>
                  ))}
                </div>
              </div>
            ) : null}
            {rfq.notes ? (
              <div className="sm:col-span-2 lg:col-span-4">
                <div className="font-mono text-[11px] uppercase tracking-[0.08em] text-muted-foreground">Notes</div>
                <p className="mt-1 text-sm text-muted-foreground">{rfq.notes}</p>
              </div>
            ) : null}
          </CardContent>
        </Card>
      ) : null}

      {showAcceptSuccess ? (
        <div className="rounded-sm border border-success/40 bg-success/10 px-4 py-3 font-mono text-sm text-foreground">
          Quote accepted — order created.{" "}
          <Link href="/orders" className="font-medium text-primary hover:underline">
            View orders →
          </Link>
        </div>
      ) : null}

      {acceptError ? (
        <div className="rounded-sm border border-destructive/40 bg-destructive/10 px-4 py-3 font-mono text-sm text-destructive">
          {acceptError}
        </div>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle>Quotes</CardTitle>
          <CardDescription>
            Suppliers respond through their magic link. This table refreshes automatically; AI match scoring lands with the comparison screen (GAU-251).
          </CardDescription>
        </CardHeader>
        <CardContent>
          {quotesQuery.isLoading ? (
            <p className="py-6 text-center font-mono text-sm text-muted-foreground">Loading quotes…</p>
          ) : null}
          {quotesQuery.data && quotesQuery.data.quotes.length === 0 ? (
            <p className="py-6 text-center text-sm text-muted-foreground">
              No supplier responses yet. Suppliers are in CST (UTC+8); first quotes typically arrive within 24 hours.
            </p>
          ) : null}
          {quotesQuery.data && quotesQuery.data.quotes.length > 0 ? (
            <Table>
              <TableHead>
                <Tr>
                  <Th>Quote</Th>
                  <Th>Status</Th>
                  <Th numeric>Unit price</Th>
                  <Th numeric>Lead time</Th>
                  <Th>Valid until</Th>
                  <Th>Notes</Th>
                  {!isAccepted ? <Th /> : null}
                </Tr>
              </TableHead>
              <TableBody>
                {quotesQuery.data.quotes.map((quote) => {
                  const isThisAccepted = quote.id === alreadyAcceptedId || quote.status === "accepted";
                  const canAccept = !isAccepted && !alreadyAcceptedId && quote.status === "submitted";
                  return (
                    <Tr key={quote.id} className={isThisAccepted ? "bg-success/5" : undefined}>
                      <Td><CodeId code={quote.id} size="sm" /></Td>
                      <Td><StatusPill tone={QUOTE_TONE[quote.status] ?? "neutral"}>{quote.status}</StatusPill></Td>
                      <Td numeric className="font-mono">{formatMoney(quote.price_usd)}</Td>
                      <Td numeric className="font-mono">{quote.lead_time_days ? `${quote.lead_time_days}d` : "—"}</Td>
                      <Td className="font-mono text-xs">{quote.validity_date || "—"}</Td>
                      <Td className="max-w-[280px] truncate text-xs text-muted-foreground">{quote.supplier_notes || "—"}</Td>
                      {!isAccepted ? (
                        <Td>
                          {canAccept ? (
                            <Button
                              size="sm"
                              variant="default"
                              disabled={accepting === quote.id}
                              onClick={() => handleAccept(quote.id)}
                            >
                              {accepting === quote.id ? "Accepting…" : "Accept"}
                            </Button>
                          ) : null}
                        </Td>
                      ) : null}
                    </Tr>
                  );
                })}
              </TableBody>
            </Table>
          ) : null}
        </CardContent>
      </Card>
    </main>
  );
}
