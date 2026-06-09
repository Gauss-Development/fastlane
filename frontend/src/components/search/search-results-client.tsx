"use client";

import { FormEvent, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Modal } from "@/components/ui/modal";
import { RouteIndicator } from "@/components/ui/route-indicator";
import { StatusPill } from "@/components/ui/pill";
import { search, type ParsedSpecs, type ProductHit, type ProductSpecs } from "@/lib/search/client";
import { cn } from "@/lib/utils";

const SPEC_LABELS: Record<keyof ParsedSpecs, string> = {
  data_rate: "data rate",
  form_factor: "form factor",
  reach_km: "reach",
  wavelength_nm: "wavelength",
  compatibility: "compatibility",
  fiber_type: "fiber",
  qty_estimated: "qty",
  free_text: "intent",
};

const ADDABLE_FIELDS: Array<keyof ParsedSpecs> = [
  "data_rate",
  "form_factor",
  "reach_km",
  "wavelength_nm",
  "compatibility",
  "fiber_type",
  "free_text",
];

function isSpecsEmpty(specs: ParsedSpecs | null | undefined) {
  if (!specs) return true;
  return Object.entries(specs).every(([, value]) => {
    if (Array.isArray(value)) return value.length === 0;
    return value === undefined || value === null || value === "" || value === 0;
  });
}

function formatSpecValue(key: keyof ParsedSpecs, value: ParsedSpecs[keyof ParsedSpecs]) {
  if (Array.isArray(value)) return value.join(", ");
  if (key === "reach_km" && typeof value === "number") return `${value}km`;
  if (key === "wavelength_nm" && typeof value === "number") return `${value}nm`;
  return String(value);
}

function visibleSpecEntries(specs: ParsedSpecs | null | undefined) {
  if (!specs) return [] as Array<[keyof ParsedSpecs, ParsedSpecs[keyof ParsedSpecs]]>;
  return (Object.entries(specs) as Array<[keyof ParsedSpecs, ParsedSpecs[keyof ParsedSpecs]]>).filter(([, value]) => {
    if (Array.isArray(value)) return value.length > 0;
    return value !== undefined && value !== null && value !== "" && value !== 0;
  });
}

function withoutSpec(specs: ParsedSpecs | null | undefined, key: keyof ParsedSpecs): ParsedSpecs {
  const next: ParsedSpecs = { ...(specs ?? {}) };
  delete next[key];
  return next;
}

function parseAddedSpec(field: keyof ParsedSpecs, raw: string) {
  const value = raw.trim();
  if (!value) return undefined;
  if (field === "reach_km" || field === "wavelength_nm" || field === "qty_estimated") {
    const n = Number(value.replace(/[^\d.]/g, ""));
    return Number.isFinite(n) ? n : undefined;
  }
  if (field === "compatibility") {
    return value.split(",").map((item) => item.trim()).filter(Boolean);
  }
  return value;
}

function readSpec(specs: ProductSpecs | null, key: string) {
  const value = specs?.[key];
  if (Array.isArray(value)) return value.join(", ");
  if (typeof value === "boolean") return value ? "yes" : "no";
  if (value === null || value === undefined || value === "") return "—";
  return String(value);
}

function formatMoney(value: number) {
  if (!Number.isFinite(value) || value <= 0) return "Quote";
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 2,
  }).format(value);
}

function ProductResultRow({
  hit,
  onQuote,
}: {
  hit: ProductHit;
  onQuote: (hit: ProductHit) => void;
}) {
  const specs = hit.specs;
  const compatibility = readSpec(specs, "compatibility");
  const stockTone = hit.stock_qty > 500 ? "success" : hit.stock_qty > 0 ? "warning" : "neutral";

  return (
    <Card className="transition-colors hover:border-primary/60">
      <CardContent className="grid gap-4 p-4 lg:grid-cols-[minmax(0,1fr)_220px_120px] lg:items-center">
        <Link href={`/products/${hit.id}`} className="min-w-0 space-y-2">
          <div className="flex min-w-0 flex-wrap items-center gap-3">
            <span aria-hidden className="size-2 bg-marker-cn" />
            <span className="font-mono text-base font-medium tracking-[0.04em]">
              {hit.sku}
            </span>
            {hit.supplier_verified ? (
              <StatusPill tone="success">Verified</StatusPill>
            ) : null}
            <span className="font-mono text-xs text-muted-foreground">
              SCORE {Math.round(hit.match_score)}
            </span>
          </div>
          <div>
            <h2 className="truncate text-base">{hit.name}</h2>
            {hit.name_zh ? (
              <p className="mt-1 font-mono text-xs text-muted-foreground">{hit.name_zh}</p>
            ) : null}
          </div>
          <div className="flex flex-wrap gap-x-3 gap-y-1 font-mono text-xs text-muted-foreground">
            <span>{readSpec(specs, "form_factor")}</span>
            <span>{readSpec(specs, "wavelength_nm")}nm</span>
            <span>{readSpec(specs, "reach_km")}km</span>
            <span>{readSpec(specs, "fiber_type")}</span>
            <span>{readSpec(specs, "connector")}</span>
          </div>
          <p className="truncate text-sm text-muted-foreground">
            {compatibility !== "—" ? `${compatibility} compatible` : "Compatibility available on datasheet"}
          </p>
          <p className="text-sm italic text-muted-foreground">
            {hit.match_explanation || "Vector match from the seeded catalog; Claude rationale appears when ANTHROPIC_API_KEY is configured."}
          </p>
        </Link>

        <div className="grid grid-cols-2 gap-3 font-mono text-xs lg:grid-cols-1">
          <div>
            <div className="text-muted-foreground">SUPPLIER</div>
            <div className="truncate text-foreground">{hit.supplier_name}</div>
            <div className="text-muted-foreground">{hit.supplier_city}</div>
          </div>
          <div className="flex flex-wrap gap-2">
            <StatusPill tone={stockTone}>{hit.stock_qty.toLocaleString()} stock</StatusPill>
            <StatusPill tone="info">{hit.lead_time_days}d lead</StatusPill>
          </div>
        </div>

        <div className="flex flex-row items-center justify-between gap-3 lg:flex-col lg:items-end">
          <div className="text-right">
            <div className="font-mono text-xl tabular-nums">{formatMoney(hit.price_usd)}</div>
            <div className="font-mono text-xs uppercase text-muted-foreground">
              MOQ {hit.moq || "—"}
            </div>
          </div>
          <Button
            type="button"
            onClick={() => onQuote(hit)}
            className="font-mono uppercase tracking-[0.08em]"
          >
            Quote →
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function QuoteRequestModal({
  hit,
  onClose,
}: {
  hit: ProductHit | null;
  onClose: () => void;
}) {
  const [submitted, setSubmitted] = useState(false);

  return (
    <Modal
      open={Boolean(hit)}
      onClose={() => {
        setSubmitted(false);
        onClose();
      }}
      title="Request quote"
      description="GAU-249 will persist this RFQ and email suppliers with magic links. This pass wires the buyer search surface."
    >
      {hit ? (
        submitted ? (
          <div className="space-y-4">
            <StatusPill tone="warning">Not persisted yet</StatusPill>
            <p className="text-sm text-muted-foreground">
              Quote intent captured in the UI for <span className="font-mono text-foreground">{hit.sku}</span>.
              The RFQ database record and supplier email dispatch are the next Linear issue.
            </p>
            <Button type="button" onClick={onClose}>Close</Button>
          </div>
        ) : (
          <form
            className="space-y-4"
            onSubmit={(event) => {
              event.preventDefault();
              setSubmitted(true);
            }}
          >
            <div className="rounded-sm border border-border bg-muted/30 p-3">
              <div className="font-mono text-sm">{hit.sku}</div>
              <div className="mt-1 text-sm text-muted-foreground">
                {hit.name} • {hit.supplier_name}
              </div>
            </div>
            <div className="grid gap-3 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="quote-qty">Quantity</Label>
                <Input id="quote-qty" name="quantity" type="number" min={1} defaultValue={Math.max(hit.moq || 1, 100)} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="quote-date">Target date</Label>
                <Input id="quote-date" name="target_date" type="date" />
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="quote-notes">Notes</Label>
              <textarea
                id="quote-notes"
                name="notes"
                rows={4}
                className="w-full rounded-sm border border-input bg-input-background px-3 py-2 text-sm"
                placeholder="Compatibility, labeling, warranty, or shipping constraints."
              />
            </div>
            <div className="flex gap-2">
              <Button type="submit">Preview RFQ</Button>
              <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
            </div>
          </form>
        )
      ) : null}
    </Modal>
  );
}

export function SearchResultsClient({ initialQuery }: { initialQuery: string }) {
  const router = useRouter();
  const [query, setQuery] = useState(initialQuery);
  const [submittedQuery, setSubmittedQuery] = useState(initialQuery);
  const [specOverrides, setSpecOverrides] = useState<ParsedSpecs | null>(null);
  const [newField, setNewField] = useState<keyof ParsedSpecs>("data_rate");
  const [newValue, setNewValue] = useState("");
  const [quoteHit, setQuoteHit] = useState<ProductHit | null>(null);

  const searchQuery = useQuery({
    queryKey: ["fiberlane-search", submittedQuery, specOverrides],
    queryFn: () => search({ query: submittedQuery, limit: 20, specOverrides }),
    enabled: submittedQuery.trim().length > 0,
  });

  const effectiveSpecs = useMemo(
    () => specOverrides ?? searchQuery.data?.parsed_specs ?? null,
    [searchQuery.data?.parsed_specs, specOverrides],
  );

  function handleSearchSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmed = query.trim();
    if (!trimmed) return;
    setSpecOverrides(null);
    setSubmittedQuery(trimmed);
    router.replace(`/search?q=${encodeURIComponent(trimmed)}`);
  }

  function removeSpec(key: keyof ParsedSpecs) {
    setSpecOverrides(withoutSpec(effectiveSpecs, key));
  }

  function addSpec(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const parsed = parseAddedSpec(newField, newValue);
    if (parsed === undefined) return;
    setSpecOverrides({ ...(effectiveSpecs ?? {}), [newField]: parsed });
    setNewValue("");
  }

  return (
    <main className="mx-auto flex w-full max-w-[1320px] flex-col gap-6 px-6 py-6">
      <section className="space-y-4">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="min-w-0 flex-1">
            <p className="font-mono text-xs uppercase tracking-[0.18em] text-muted-foreground">
              Section B — Search Results
            </p>
            <h1 className="mt-2 break-words font-mono text-xl md:text-2xl">
              <span className="text-primary">&gt;</span> {submittedQuery || "Describe the part you need"}
            </h1>
          </div>
          <div className="flex flex-col items-start gap-2 sm:items-end">
            <RouteIndicator size="sm" />
            {searchQuery.data?.query_id ? (
              <span className="font-mono text-xs uppercase text-muted-foreground">
                Query {searchQuery.data.query_id}
              </span>
            ) : null}
          </div>
        </div>

        <form onSubmit={handleSearchSubmit} className="flex flex-col gap-2 md:flex-row">
          <Input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            className="h-11 font-mono"
            placeholder="100G QSFP28 transceiver, 10km, Cisco compatible"
          />
          <Button type="submit" className="font-mono uppercase tracking-[0.08em]">
            Run search
          </Button>
        </form>

        <Card>
          <CardContent className="space-y-3 p-4">
            <div className="flex flex-wrap items-center gap-2">
              {visibleSpecEntries(effectiveSpecs).map(([key, value]) => (
                <button
                  key={key}
                  type="button"
                  onClick={() => removeSpec(key)}
                  className="rounded-sm border border-border bg-secondary px-3 py-1.5 font-mono text-xs uppercase tracking-[0.06em] hover:border-primary hover:text-primary"
                  title="Remove and re-run search"
                >
                  {SPEC_LABELS[key]}: {formatSpecValue(key, value)} ×
                </button>
              ))}
              {isSpecsEmpty(effectiveSpecs) ? (
                <span className="font-mono text-xs uppercase tracking-[0.08em] text-muted-foreground">
                  No structured specs extracted yet. Search still runs vector ranking.
                </span>
              ) : null}
              {specOverrides ? (
                <Button type="button" variant="ghost" size="sm" onClick={() => setSpecOverrides(null)}>
                  Reset AI specs
                </Button>
              ) : null}
            </div>
            <form onSubmit={addSpec} className="flex flex-col gap-2 sm:flex-row">
              <select
                value={newField}
                onChange={(event) => setNewField(event.target.value as keyof ParsedSpecs)}
                className="h-9 rounded-sm border border-input bg-input-background px-3 font-mono text-sm"
                aria-label="Spec field"
              >
                {ADDABLE_FIELDS.map((field) => (
                  <option key={field} value={field}>{SPEC_LABELS[field]}</option>
                ))}
              </select>
              <Input
                value={newValue}
                onChange={(event) => setNewValue(event.target.value)}
                placeholder="Add or override a filter"
                className="h-9 font-mono"
              />
              <Button type="submit" variant="secondary" size="sm">
                + add filter
              </Button>
            </form>
          </CardContent>
        </Card>
      </section>

      {searchQuery.isLoading ? (
        <div className="space-y-3">
          {[1, 2, 3, 4].map((i) => (
            <Card key={i} className="animate-pulse">
              <CardContent className="h-32 p-4">
                <div className="h-4 w-1/3 rounded-sm bg-muted" />
                <div className="mt-4 h-3 w-2/3 rounded-sm bg-muted" />
                <div className="mt-3 h-3 w-1/2 rounded-sm bg-muted" />
              </CardContent>
            </Card>
          ))}
        </div>
      ) : null}

      {searchQuery.error ? (
        <Card className="border-destructive/50">
          <CardHeader>
            <CardTitle>Search failed</CardTitle>
            <CardDescription>
              {(searchQuery.error as Error).message}
            </CardDescription>
          </CardHeader>
        </Card>
      ) : null}

      {searchQuery.data && !searchQuery.isLoading ? (
        <section className="space-y-3">
          <div className="flex items-center justify-between gap-3">
            <p className="font-mono text-xs uppercase tracking-[0.12em] text-muted-foreground">
              {searchQuery.data.results.length} ranked matches
            </p>
            <p className={cn("font-mono text-xs uppercase tracking-[0.12em]", specOverrides ? "text-primary" : "text-muted-foreground")}>
              {specOverrides ? "Manual spec override active" : "AI extracted specs active"}
            </p>
          </div>

          {searchQuery.data.results.length === 0 ? (
            <Card>
              <CardContent className="py-10 text-center text-sm text-muted-foreground">
                No products found. Remove a chip or broaden the query.
              </CardContent>
            </Card>
          ) : (
            searchQuery.data.results.map((hit) => (
              <ProductResultRow key={hit.id} hit={hit} onQuote={setQuoteHit} />
            ))
          )}
        </section>
      ) : null}

      {!submittedQuery ? (
        <Card>
          <CardContent className="py-10 text-center text-sm text-muted-foreground">
            Start from the dashboard or enter a query above.
          </CardContent>
        </Card>
      ) : null}

      <QuoteRequestModal hit={quoteHit} onClose={() => setQuoteHit(null)} />
    </main>
  );
}
