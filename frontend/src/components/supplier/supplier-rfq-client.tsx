"use client";

import { FormEvent, useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { Button } from "@/components/ui/button";
import { CodeId } from "@/components/ui/code-id";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { StatusPill } from "@/components/ui/pill";
import { getSupplierRFQ, submitSupplierQuote } from "@/lib/rfqs/client";
import type { Quote } from "@/lib/rfqs/types";

// Bilingual field label: English over smaller Mandarin, per GAU-250.
function BiLabel({ htmlFor, en, zh }: { htmlFor?: string; en: string; zh: string }) {
  return (
    <Label htmlFor={htmlFor} className="flex flex-col gap-0.5">
      <span className="font-mono text-xs uppercase tracking-[0.08em]">{en}</span>
      <span className="font-mono text-[11px] text-muted-foreground">{zh}</span>
    </Label>
  );
}

function BiValueRow({ en, zh, children }: { en: string; zh: string; children: React.ReactNode }) {
  return (
    <div className="border-b border-border py-3 last:border-b-0">
      <div className="font-mono text-[11px] uppercase tracking-[0.08em] text-muted-foreground">
        {en} / {zh}
      </div>
      <div className="mt-1 text-sm">{children}</div>
    </div>
  );
}

export function SupplierRFQClient({ token }: { token: string }) {
  const [submitted, setSubmitted] = useState<Quote | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const viewQuery = useQuery({
    queryKey: ["supplier-rfq", token],
    queryFn: () => getSupplierRFQ(token),
    retry: false,
  });

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (submitting) return;

    const form = new FormData(event.currentTarget);
    const priceUsd = Number(form.get("price_usd") ?? 0);
    const leadTimeDays = Number(form.get("lead_time_days") ?? 0);
    if (!Number.isFinite(priceUsd) || priceUsd <= 0 || !Number.isFinite(leadTimeDays) || leadTimeDays <= 0) {
      setError("Unit price and lead time are required. / 请填写单价和交货期。");
      return;
    }

    setSubmitting(true);
    setError(null);
    try {
      const quote = await submitSupplierQuote(token, {
        priceUsd,
        leadTimeDays,
        validityDate: String(form.get("validity_date") ?? ""),
        notes: String(form.get("notes") ?? ""),
      });
      setSubmitted(quote);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to submit quote. / 提交失败。");
      setSubmitting(false);
    }
  }

  const view = viewQuery.data;
  const alreadySubmitted = view?.quote && view.quote.status !== "pending" ? view.quote : null;
  const finalQuote = submitted ?? alreadySubmitted;

  return (
    <main className="mx-auto flex min-h-dvh w-full max-w-[560px] flex-col gap-6 px-4 py-8">
      <div className="flex items-center justify-between border-b border-foreground pb-3">
        <span className="font-mono text-sm font-bold tracking-[0.08em]">FIBERLANE</span>
        <span className="font-mono text-xs text-muted-foreground">CN ───► US</span>
      </div>

      {viewQuery.isLoading ? (
        <p className="py-12 text-center font-mono text-sm text-muted-foreground">
          Loading RFQ… / 询盘加载中…
        </p>
      ) : null}

      {viewQuery.error ? (
        <div className="space-y-3 border border-destructive/50 p-6 text-center">
          <p className="font-mono text-sm font-bold">THIS LINK IS INVALID OR HAS EXPIRED</p>
          <p className="font-mono text-sm text-muted-foreground">此链接无效或已过期</p>
          <p className="text-xs text-muted-foreground">
            Ask the buyer to send a new RFQ invitation. / 请联系买家重新发送询盘邀请。
          </p>
        </div>
      ) : null}

      {view ? (
        <>
          <div>
            <p className="font-mono text-base font-bold">
              NEW RFQ FROM {(view.rfq.buyer_company || "A US BUYER").toUpperCase()} (United States)
            </p>
            <p className="mt-1 font-mono text-sm text-muted-foreground">
              来自 {view.rfq.buyer_company || "美国买家"}（美国）的新询盘
            </p>
            <div className="mt-3">
              <CodeId code={view.rfq.id} size="sm" />
            </div>
          </div>

          <div className="border border-border bg-card px-4">
            <BiValueRow en="PART REQUESTED" zh="询盘部件">
              <span className="font-mono">{view.rfq.query_text}</span>
            </BiValueRow>
            <BiValueRow en="QUANTITY" zh="数量">
              <span className="font-mono">{view.rfq.qty ? `${view.rfq.qty.toLocaleString()} units` : "—"}</span>
            </BiValueRow>
            <BiValueRow en="DELIVERY TO" zh="交货地点">
              <span className="font-mono">{view.rfq.shipping_address || "United States"}</span>
            </BiValueRow>
            <BiValueRow en="TARGET DATE" zh="目标日期">
              <span className="font-mono">{view.rfq.target_date || "—"}</span>
            </BiValueRow>
            {view.rfq.notes ? (
              <BiValueRow en="NOTES" zh="备注">
                <span className="text-muted-foreground">{view.rfq.notes}</span>
              </BiValueRow>
            ) : null}
          </div>

          {finalQuote ? (
            <div className="space-y-3 border border-border bg-card p-6 text-center">
              <StatusPill tone="success">Submitted / 已提交</StatusPill>
              <p className="font-mono text-sm font-bold">QUOTE RECEIVED — THANK YOU / 报价已收到，谢谢</p>
              <p className="text-sm text-muted-foreground">
                The buyer has been notified by email. / 买家已收到邮件通知。
              </p>
              <div className="flex justify-center">
                <CodeId code={finalQuote.id} size="sm" copyable />
              </div>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              <p className="border-t border-border pt-4 font-mono text-sm font-bold">
                YOUR QUOTE / 您的报价
              </p>
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <BiLabel htmlFor="price-usd" en="Unit price USD" zh="单价（美元）" />
                  <Input id="price-usd" name="price_usd" type="number" step="0.01" min="0.01" required inputMode="decimal" className="font-mono" />
                </div>
                <div className="space-y-2">
                  <BiLabel htmlFor="lead-time" en="Lead time days" zh="交货期（天）" />
                  <Input id="lead-time" name="lead_time_days" type="number" min="1" required inputMode="numeric" className="font-mono" />
                </div>
              </div>
              <div className="space-y-2">
                <BiLabel htmlFor="validity-date" en="Quote valid until (optional)" zh="报价有效期（可选）" />
                <Input id="validity-date" name="validity_date" type="date" className="font-mono" />
              </div>
              <div className="space-y-2">
                <BiLabel htmlFor="supplier-notes" en="Notes (optional)" zh="备注（可选）" />
                <textarea
                  id="supplier-notes"
                  name="notes"
                  rows={3}
                  className="w-full rounded-sm border border-input bg-input-background px-3 py-2 text-sm"
                />
              </div>
              {error ? <p className="text-sm text-destructive">{error}</p> : null}
              <Button type="submit" disabled={submitting} className="h-12 w-full font-mono text-sm font-bold uppercase tracking-[0.08em]">
                {submitting ? "Submitting… / 提交中…" : "Submit Quote / 提交报价"}
              </Button>
              <p className="text-center text-xs text-muted-foreground">
                No login required. This link is unique to {view.supplier_name || "your company"}.
                <br />
                无需登录。此链接仅适用于贵公司。
              </p>
            </form>
          )}
        </>
      ) : null}
    </main>
  );
}
