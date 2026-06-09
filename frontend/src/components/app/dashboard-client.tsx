"use client";

import { FormEvent, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { CodeId } from "@/components/ui/code-id";
import { Input } from "@/components/ui/input";
import { RouteIndicator } from "@/components/ui/route-indicator";
import { StatusPill } from "@/components/ui/pill";
import { Table, TableBody, Td, Th, TableHead, Tr } from "@/components/ui/table";
import {
  exampleQueries,
  featuredSuppliers,
  liveQuotes,
  recentRFQs,
} from "@/components/search/demo-data";
import { useAuthStore } from "@/lib/stores/auth-store";

function buyerDomain(email?: string) {
  if (!email?.includes("@")) {
    return "acme-corp.com";
  }
  return email.split("@")[1]?.toLowerCase() || "acme-corp.com";
}

export function DashboardClient() {
  const router = useRouter();
  const user = useAuthStore((state) => state.user);
  const [query, setQuery] = useState<string>(exampleQueries[0]);

  function submitSearch(nextQuery: string = query) {
    const trimmed = nextQuery.trim();
    if (!trimmed) return;
    router.push(`/search?q=${encodeURIComponent(trimmed)}`);
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    submitSearch();
  }

  return (
    <div className="min-h-dvh bg-background">
      <header className="sticky top-0 z-10 border-b border-border bg-background/95 px-6 py-3 backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="font-mono text-xs uppercase tracking-[0.12em] text-muted-foreground">
            BUYER • {buyerDomain(user?.email)} • USD • PST
          </div>
          <RouteIndicator size="sm" />
        </div>
      </header>

      <main className="mx-auto flex w-full max-w-[1440px] flex-col gap-6 px-6 py-6">
        <section className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_360px]">
          <Card className="border-border-strong">
            <CardHeader className="gap-5">
              <div>
                <p className="mb-2 font-mono text-xs uppercase tracking-[0.18em] text-muted-foreground">
                  Section A — Hybrid Search
                </p>
                <CardTitle className="max-w-3xl text-2xl">
                  Describe the optical component. Fiberlane extracts the specs
                  and ranks verified Chinese catalog matches.
                </CardTitle>
              </div>
              <RouteIndicator size="lg" />
            </CardHeader>
            <CardContent className="space-y-4">
              <form onSubmit={handleSubmit} className="flex flex-col gap-3 md:flex-row">
                <div className="relative min-w-0 flex-1">
                  <span className="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 font-mono text-lg text-primary">
                    &gt;
                  </span>
                  <Input
                    value={query}
                    onChange={(event) => setQuery(event.target.value)}
                    placeholder="Describe the part you need..."
                    className="h-14 pl-10 font-mono text-base"
                    aria-label="Search product catalog"
                  />
                </div>
                <Button type="submit" size="lg" className="h-14 font-mono uppercase tracking-[0.08em]">
                  Search catalog
                </Button>
              </form>

              <div className="flex flex-wrap gap-2">
                {exampleQueries.map((example) => (
                  <button
                    key={example}
                    type="button"
                    onClick={() => {
                      setQuery(example);
                      submitSearch(example);
                    }}
                    className="rounded-sm border border-border bg-secondary px-3 py-2 font-mono text-xs uppercase tracking-[0.06em] text-secondary-foreground transition-colors hover:border-primary hover:text-primary"
                  >
                    {example}
                  </button>
                ))}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Live quote feed</CardTitle>
              <CardDescription>
                Supplier responses from seeded demo RFQs. Real-time transport comes after RFQ persistence.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {liveQuotes.map((quote) => (
                <div key={`${quote.rfqId}-${quote.supplier}`} className="border-b border-border pb-4 last:border-b-0 last:pb-0">
                  <div className="mb-1 flex items-center justify-between gap-3">
                    <span className="font-mono text-xs uppercase text-muted-foreground">
                      {quote.receivedAt}
                    </span>
                    <span className="font-mono text-sm tabular-nums text-primary">
                      {quote.price}
                    </span>
                  </div>
                  <p className="font-mono text-sm text-foreground">{quote.part}</p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {quote.supplier} • {quote.city} • {quote.leadTime}
                  </p>
                  <CodeId code={quote.rfqId} size="sm" className="mt-2" />
                </div>
              ))}
            </CardContent>
          </Card>
        </section>

        <section className="grid gap-6 xl:grid-cols-[minmax(0,2fr)_minmax(380px,1fr)]">
          <Card>
            <CardHeader>
              <CardTitle>Recent RFQs</CardTitle>
              <CardDescription>
                Dense demo workbench. Rows will link into GAU-251 quote comparison when RFQ persistence lands.
              </CardDescription>
            </CardHeader>
            <CardContent>
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
                  {recentRFQs.map((rfq) => (
                    <Tr key={rfq.id}>
                      <Td>
                        <CodeId code={rfq.id} size="sm" />
                      </Td>
                      <Td className="max-w-[360px] truncate font-mono text-xs">
                        {rfq.query}
                      </Td>
                      <Td numeric>{rfq.qty}</Td>
                      <Td>
                        <StatusPill tone={rfq.tone}>{rfq.status}</StatusPill>
                      </Td>
                      <Td numeric>{rfq.age}</Td>
                    </Tr>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Operating snapshot</CardTitle>
              <CardDescription>
                Cross-border sourcing health for the seeded photonics catalog.
              </CardDescription>
            </CardHeader>
            <CardContent className="grid grid-cols-2 gap-3">
              {[
                ["Verified suppliers", "7"],
                ["Seeded SKUs", "80+"],
                ["Median lead time", "10d"],
                ["Avg on-time rate", "95.9%"],
              ].map(([label, value]) => (
                <div key={label} className="border border-border bg-muted/30 p-3">
                  <div className="font-mono text-2xl tabular-nums">{value}</div>
                  <div className="mt-1 font-mono text-[11px] uppercase tracking-[0.08em] text-muted-foreground">
                    {label}
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        </section>

        <section>
          <div className="mb-3 flex items-end justify-between gap-3">
            <div>
              <h2 className="text-lg">Featured suppliers</h2>
              <p className="mt-1 text-sm text-muted-foreground">
                Real Chinese photonics suppliers from the catalog seed.
              </p>
            </div>
            <Link
              href="/suppliers"
              className="font-mono text-xs uppercase tracking-[0.08em] text-primary hover:underline"
            >
              View supplier surface →
            </Link>
          </div>
          <div className="flex gap-3 overflow-x-auto pb-2">
            {featuredSuppliers.map((supplier) => (
              <Card key={supplier.code} className="min-w-[260px]">
                <CardHeader className="pb-3">
                  <CodeId code={supplier.code} size="sm" />
                  <div>
                    <CardTitle className="text-base">{supplier.name}</CardTitle>
                    <CardDescription className="font-mono">
                      {supplier.nameZh} • {supplier.city}
                    </CardDescription>
                  </div>
                </CardHeader>
                <CardContent className="space-y-3 text-sm">
                  <p className="font-mono text-xs uppercase tracking-[0.08em] text-muted-foreground">
                    {supplier.capability}
                  </p>
                  <div className="grid grid-cols-2 gap-2 font-mono text-xs">
                    <div>
                      <div className="text-muted-foreground">ON-TIME</div>
                      <div className="text-success">{supplier.onTimeRate}</div>
                    </div>
                    <div>
                      <div className="text-muted-foreground">ORDERS</div>
                      <div>{supplier.orders}</div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </section>
      </main>
    </div>
  );
}
