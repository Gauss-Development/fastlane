"use client";

import Link from "next/link";
import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { CodeId } from "@/components/ui/code-id";
import { StatusPill, type StatusTone } from "@/components/ui/pill";
import { Table, TableBody, Td, Th, TableHead, Tr } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Modal } from "@/components/ui/modal";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { listOpenRFQs, listRFQs, submitManufacturerQuote } from "@/lib/rfqs/client";
import type { RFQ, RFQStatus } from "@/lib/rfqs/types";
import { useAuthStore } from "@/lib/stores/auth-store";

const STATUS_TONE: Record<RFQStatus, StatusTone> = {
  open: "info",
  quoted: "warning",
  accepted: "success",
  closed: "neutral",
};

function age(createdAt: string) {
  const created = new Date(createdAt).getTime();
  if (!Number.isFinite(created)) return "—";
  const minutes = Math.max(0, Math.floor((Date.now() - created) / 60000));
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  if (hours < 48) return `${hours}h`;
  return `${Math.floor(hours / 24)}d`;
}

interface QuoteFormState {
  price_usd: string;
  lead_time_days: string;
  validity_date: string;
  notes: string;
}

const EMPTY_FORM: QuoteFormState = {
  price_usd: "",
  lead_time_days: "",
  validity_date: "",
  notes: "",
};

function QuoteModal({
  rfq,
  onClose,
}: {
  rfq: RFQ;
  onClose: () => void;
}) {
  const queryClient = useQueryClient();
  const [form, setForm] = useState<QuoteFormState>(EMPTY_FORM);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [submitted, setSubmitted] = useState(false);

  function set(field: keyof QuoteFormState) {
    return (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
      setForm((prev) => ({ ...prev, [field]: e.target.value }));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const price = parseFloat(form.price_usd);
    const leadTime = parseInt(form.lead_time_days, 10);
    if (!Number.isFinite(price) || price <= 0) {
      setError("Enter a valid price.");
      return;
    }
    if (!Number.isFinite(leadTime) || leadTime <= 0) {
      setError("Enter a valid lead time.");
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      await submitManufacturerQuote(rfq.id, {
        price_usd: price,
        lead_time_days: leadTime,
        validity_date: form.validity_date,
        notes: form.notes,
      });
      setSubmitted(true);
      await queryClient.invalidateQueries({ queryKey: ["rfqs", "open"] });
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal
      open
      onClose={onClose}
      title="Submit Quote"
      description={`RFQ: ${rfq.query_text.slice(0, 80)}${rfq.query_text.length > 80 ? "…" : ""}`}
    >
      {submitted ? (
        <div className="flex flex-col gap-4">
          <p className="font-mono text-sm text-foreground">Quote submitted successfully.</p>
          <Button variant="outline" size="sm" onClick={onClose}>Close</Button>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="qf-price" className="font-mono text-[11px] uppercase tracking-[0.08em] text-muted-foreground">
                Unit price (USD)
              </Label>
              <Input
                id="qf-price"
                type="number"
                step="0.01"
                min="0.01"
                placeholder="0.00"
                required
                value={form.price_usd}
                onChange={set("price_usd")}
                className="font-mono"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="qf-lead" className="font-mono text-[11px] uppercase tracking-[0.08em] text-muted-foreground">
                Lead time (days)
              </Label>
              <Input
                id="qf-lead"
                type="number"
                step="1"
                min="1"
                placeholder="30"
                required
                value={form.lead_time_days}
                onChange={set("lead_time_days")}
                className="font-mono"
              />
            </div>
            <div className="flex flex-col gap-1.5 sm:col-span-2">
              <Label htmlFor="qf-validity" className="font-mono text-[11px] uppercase tracking-[0.08em] text-muted-foreground">
                Valid until
              </Label>
              <Input
                id="qf-validity"
                type="date"
                value={form.validity_date}
                onChange={set("validity_date")}
                className="font-mono"
              />
            </div>
            <div className="flex flex-col gap-1.5 sm:col-span-2">
              <Label htmlFor="qf-notes" className="font-mono text-[11px] uppercase tracking-[0.08em] text-muted-foreground">
                Notes
              </Label>
              <textarea
                id="qf-notes"
                rows={3}
                placeholder="Certifications, packaging, MOQ…"
                value={form.notes}
                onChange={set("notes")}
                className="w-full resize-none rounded-sm border border-border bg-background px-3 py-2 font-mono text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 focus:ring-offset-background"
              />
            </div>
          </div>

          {error ? (
            <p className="font-mono text-xs text-destructive">{error}</p>
          ) : null}

          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" size="sm" onClick={onClose} disabled={submitting}>
              Cancel
            </Button>
            <Button type="submit" size="sm" disabled={submitting}>
              {submitting ? "Submitting…" : "Submit quote"}
            </Button>
          </div>
        </form>
      )}
    </Modal>
  );
}

export function RFQListClient() {
  const isManufacturer = useAuthStore((state) => state.user?.role) === "manufacturer";
  const [quotingRFQ, setQuotingRFQ] = useState<RFQ | null>(null);

  const rfqsQuery = useQuery({
    queryKey: isManufacturer ? ["rfqs", "open"] : ["rfqs"],
    queryFn: () => (isManufacturer ? listOpenRFQs({ limit: 50 }) : listRFQs({ limit: 50 })),
  });

  return (
    <main className="mx-auto w-full max-w-[1200px] px-6 py-6">
      {quotingRFQ ? (
        <QuoteModal rfq={quotingRFQ} onClose={() => setQuotingRFQ(null)} />
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle>{isManufacturer ? "Incoming RFQs" : "RFQs"}</CardTitle>
          <CardDescription>
            {isManufacturer
              ? "Open quote requests from buyers. Buyer contact and delivery address are shared once a quote is accepted."
              : "Quote requests sent to verified suppliers. Suppliers respond through magic links — no account needed on their side."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {rfqsQuery.isLoading ? (
            <p className="py-8 text-center font-mono text-sm text-muted-foreground">Loading RFQs…</p>
          ) : null}

          {rfqsQuery.error ? (
            <p className="py-8 text-center text-sm text-destructive">
              {(rfqsQuery.error as Error).message}
            </p>
          ) : null}

          {rfqsQuery.data && rfqsQuery.data.rfqs.length === 0 ? (
            <div className="py-10 text-center">
              {isManufacturer ? (
                <p className="text-sm text-muted-foreground">
                  No open RFQs right now. New quote requests from buyers appear here.
                </p>
              ) : (
                <>
                  <p className="text-sm text-muted-foreground">
                    No RFQs yet. Run a search and hit <span className="font-mono">Quote →</span> on a result.
                  </p>
                  <Link href="/dashboard" className="mt-3 inline-block font-mono text-xs uppercase tracking-[0.08em] text-primary hover:underline">
                    Go to search →
                  </Link>
                </>
              )}
            </div>
          ) : null}

          {rfqsQuery.data && rfqsQuery.data.rfqs.length > 0 ? (
            <Table>
              <TableHead>
                <Tr>
                  <Th>RFQ</Th>
                  <Th>Request</Th>
                  <Th numeric>Qty</Th>
                  <Th>Status</Th>
                  <Th numeric>Age</Th>
                  {isManufacturer ? <Th /> : null}
                </Tr>
              </TableHead>
              <TableBody>
                {rfqsQuery.data.rfqs.map((rfq) => (
                  <Tr key={rfq.id}>
                    <Td>
                      {isManufacturer ? (
                        <CodeId code={rfq.id} size="sm" />
                      ) : (
                        <Link href={`/rfqs/${encodeURIComponent(rfq.id)}`} className="hover:underline">
                          <CodeId code={rfq.id} size="sm" />
                        </Link>
                      )}
                    </Td>
                    <Td className="max-w-[360px] truncate font-mono text-xs">{rfq.query_text}</Td>
                    <Td numeric>{rfq.qty || "—"}</Td>
                    <Td>
                      <StatusPill tone={STATUS_TONE[rfq.status] ?? "neutral"}>{rfq.status}</StatusPill>
                    </Td>
                    <Td numeric>{age(rfq.created_at)}</Td>
                    {isManufacturer ? (
                      <Td>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => setQuotingRFQ(rfq)}
                        >
                          Quote
                        </Button>
                      </Td>
                    ) : null}
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
