"use client";

import Link from "next/link";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { CodeId } from "@/components/ui/code-id";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { StatusPill } from "@/components/ui/pill";
import { listManufacturers } from "@/lib/manufacturers/client";
import type { Manufacturer } from "@/lib/manufacturers/types";

const SERVICE_TYPES = ["pcb_fab", "pcba", "cable_assembly", "enclosure", "box_build"];
const ASSEMBLY_TYPES = ["smt", "tht", "mixed"];

type Filters = {
  cluster: string;
  service_type: string;
  assembly_type: string;
  verified_only: boolean;
};

function ManufacturerCard({ m }: { m: Manufacturer }) {
  return (
    <Link href={`/manufacturers/${encodeURIComponent(m.id)}`} className="block">
      <Card className="h-full transition-colors hover:border-primary">
        <CardHeader>
          <div className="flex flex-wrap items-start justify-between gap-2">
            <div>
              <CardTitle className="font-mono text-base">{m.name}</CardTitle>
              {m.name_zh ? (
                <CardDescription className="mt-0.5">{m.name_zh}</CardDescription>
              ) : null}
            </div>
            {m.verified ? (
              <StatusPill tone="success">Verified</StatusPill>
            ) : (
              <StatusPill tone="neutral">Unverified</StatusPill>
            )}
          </div>
          <div className="mt-2 flex flex-wrap items-center gap-2">
            {m.cluster ? <Badge variant="outline">{m.cluster}</Badge> : null}
            {m.city ? (
              <span className="font-mono text-xs text-muted-foreground">{m.city}</span>
            ) : null}
          </div>
        </CardHeader>
        <CardContent className="space-y-3">
          {m.service_types.length > 0 || m.assembly_types.length > 0 ? (
            <div className="flex flex-wrap gap-1.5">
              {m.service_types.map((s) => (
                <Badge key={`svc-${s}`} variant="primary">{s}</Badge>
              ))}
              {m.assembly_types.map((a) => (
                <Badge key={`asm-${a}`}>{a}</Badge>
              ))}
            </div>
          ) : null}
          <dl className="grid grid-cols-3 gap-2 font-mono text-xs">
            <div>
              <dt className="uppercase tracking-[0.08em] text-muted-foreground">Layers</dt>
              <dd className="mt-0.5">
                {m.min_layers || m.max_layers ? `${m.min_layers}–${m.max_layers}` : "—"}
              </dd>
            </div>
            <div>
              <dt className="uppercase tracking-[0.08em] text-muted-foreground">MOQ</dt>
              <dd className="mt-0.5">{m.min_order_qty ? m.min_order_qty.toLocaleString() : "—"}</dd>
            </div>
            <div>
              <dt className="uppercase tracking-[0.08em] text-muted-foreground">Lead</dt>
              <dd className="mt-0.5">{m.lead_time_days ? `${m.lead_time_days}d` : "—"}</dd>
            </div>
          </dl>
          {m.certifications.length > 0 ? (
            <div className="flex flex-wrap gap-1.5">
              {m.certifications.map((c) => (
                <Badge key={`cert-${c}`} variant="outline">{c}</Badge>
              ))}
            </div>
          ) : null}
          <CodeId code={m.id} size="sm" />
        </CardContent>
      </Card>
    </Link>
  );
}

export function ManufacturersListClient() {
  const [filters, setFilters] = useState<Filters>({
    cluster: "",
    service_type: "",
    assembly_type: "",
    verified_only: false,
  });

  const query = useQuery({
    queryKey: ["manufacturers", filters],
    queryFn: () =>
      listManufacturers({
        cluster: filters.cluster || undefined,
        service_type: filters.service_type || undefined,
        assembly_type: filters.assembly_type || undefined,
        verified_only: filters.verified_only || undefined,
        limit: 60,
      }),
  });

  return (
    <main className="mx-auto flex w-full max-w-[1200px] flex-col gap-6 px-6 py-6">
      <div>
        <p className="font-mono text-xs uppercase tracking-[0.18em] text-muted-foreground">Manufacturers</p>
        <h1 className="mt-2 text-lg">Verified photonics fabs</h1>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Filters</CardTitle>
          <CardDescription>Narrow by cluster, capability, and verification status.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div className="space-y-2">
            <Label htmlFor="f-cluster">Cluster</Label>
            <Input
              id="f-cluster"
              placeholder="e.g. Shenzhen"
              value={filters.cluster}
              onChange={(e) => setFilters((f) => ({ ...f, cluster: e.target.value }))}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="f-service">Service type</Label>
            <select
              id="f-service"
              className="h-9 w-full rounded-sm px-3 text-sm"
              value={filters.service_type}
              onChange={(e) => setFilters((f) => ({ ...f, service_type: e.target.value }))}
            >
              <option value="">Any</option>
              {SERVICE_TYPES.map((s) => (
                <option key={s} value={s}>{s}</option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="f-assembly">Assembly type</Label>
            <select
              id="f-assembly"
              className="h-9 w-full rounded-sm px-3 text-sm"
              value={filters.assembly_type}
              onChange={(e) => setFilters((f) => ({ ...f, assembly_type: e.target.value }))}
            >
              <option value="">Any</option>
              {ASSEMBLY_TYPES.map((a) => (
                <option key={a} value={a}>{a}</option>
              ))}
            </select>
          </div>
          <div className="flex items-end">
            <label className="flex items-center gap-2 font-mono text-xs uppercase tracking-[0.08em]">
              <input
                type="checkbox"
                className="size-4"
                checked={filters.verified_only}
                onChange={(e) => setFilters((f) => ({ ...f, verified_only: e.target.checked }))}
              />
              Verified only
            </label>
          </div>
        </CardContent>
      </Card>

      {query.isLoading ? (
        <p className="py-8 text-center font-mono text-sm text-muted-foreground">Loading manufacturers…</p>
      ) : null}

      {query.error ? (
        <p className="text-sm text-destructive">{(query.error as Error).message}</p>
      ) : null}

      {query.data && query.data.manufacturers.length === 0 ? (
        <p className="py-8 text-center text-sm text-muted-foreground">
          No manufacturers match these filters.
        </p>
      ) : null}

      {query.data && query.data.manufacturers.length > 0 ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {query.data.manufacturers.map((m) => (
            <ManufacturerCard key={m.id} m={m} />
          ))}
        </div>
      ) : null}
    </main>
  );
}
