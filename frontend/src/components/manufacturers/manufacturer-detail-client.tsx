"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { CodeId } from "@/components/ui/code-id";
import { StatusPill } from "@/components/ui/pill";
import { getManufacturer } from "@/lib/manufacturers/client";
import type { Manufacturer } from "@/lib/manufacturers/types";

function Field({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="font-mono text-[11px] uppercase tracking-[0.08em] text-muted-foreground">{label}</div>
      <div className="mt-1 font-mono text-sm">{value || "—"}</div>
    </div>
  );
}

function Chips({ label, items, variant }: { label: string; items: string[]; variant?: "primary" | "outline" | "default" }) {
  return (
    <div>
      <div className="font-mono text-[11px] uppercase tracking-[0.08em] text-muted-foreground">{label}</div>
      <div className="mt-2 flex flex-wrap gap-1.5">
        {items.length > 0 ? (
          items.map((x) => (
            <Badge key={x} variant={variant ?? "default"}>{x}</Badge>
          ))
        ) : (
          <span className="font-mono text-sm text-muted-foreground">—</span>
        )}
      </div>
    </div>
  );
}

export function ManufacturerDetailClient({ manufacturerId }: { manufacturerId: string }) {
  const query = useQuery({
    queryKey: ["manufacturer", manufacturerId],
    queryFn: () => getManufacturer(manufacturerId),
  });

  const m: Manufacturer | undefined = query.data;

  return (
    <main className="mx-auto flex w-full max-w-[1200px] flex-col gap-6 px-6 py-6">
      <div>
        <p className="font-mono text-xs uppercase tracking-[0.18em] text-muted-foreground">Manufacturer</p>
        <div className="mt-2">
          <CodeId code={manufacturerId} size="lg" copyable />
        </div>
      </div>

      {query.isLoading ? (
        <Card>
          <CardContent className="py-10 text-center font-mono text-sm text-muted-foreground">
            Loading manufacturer…
          </CardContent>
        </Card>
      ) : null}

      {query.error ? (
        <Card className="border-destructive/50">
          <CardHeader>
            <CardTitle>Could not load manufacturer</CardTitle>
            <CardDescription>{(query.error as Error).message}</CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/manufacturers" className="font-mono text-xs uppercase tracking-[0.08em] text-primary hover:underline">
              ← Back to manufacturers
            </Link>
          </CardContent>
        </Card>
      ) : null}

      {m ? (
        <>
          <Card>
            <CardHeader>
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <CardTitle className="font-mono text-lg">{m.name}</CardTitle>
                  {m.name_zh ? <CardDescription className="mt-0.5">{m.name_zh}</CardDescription> : null}
                  <div className="mt-2 flex flex-wrap items-center gap-2">
                    {m.cluster ? <Badge variant="outline">{m.cluster}</Badge> : null}
                    {m.city ? <span className="font-mono text-xs text-muted-foreground">{m.city}</span> : null}
                  </div>
                </div>
                {m.verified ? (
                  <StatusPill tone="success">Verified</StatusPill>
                ) : (
                  <StatusPill tone="neutral">Unverified</StatusPill>
                )}
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              {m.description ? (
                <p className="text-sm text-muted-foreground">{m.description}</p>
              ) : null}
              {m.website ? (
                <a
                  href={m.website}
                  target="_blank"
                  rel="noreferrer"
                  className="font-mono text-sm text-primary hover:underline"
                >
                  {m.website}
                </a>
              ) : null}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Capabilities</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
              <Field
                label="Layers"
                value={m.min_layers || m.max_layers ? `${m.min_layers}–${m.max_layers}` : ""}
              />
              <Field label="Min order qty" value={m.min_order_qty ? m.min_order_qty.toLocaleString() : ""} />
              <Field label="Max order qty" value={m.max_order_qty ? m.max_order_qty.toLocaleString() : ""} />
              <Field label="Lead time" value={m.lead_time_days ? `${m.lead_time_days} days` : ""} />
              <Field label="Monthly capacity" value={m.monthly_capacity ? m.monthly_capacity.toLocaleString() : ""} />
              <Field label="Smallest package" value={m.smallest_package} />
              <div className="sm:col-span-2 lg:col-span-4">
                <Chips label="Service types" items={m.service_types} variant="primary" />
              </div>
              <div className="sm:col-span-2 lg:col-span-4">
                <Chips label="Assembly types" items={m.assembly_types} />
              </div>
              <div className="sm:col-span-2 lg:col-span-4">
                <Chips label="Materials" items={m.materials} variant="outline" />
              </div>
              <div className="sm:col-span-2 lg:col-span-4">
                <Chips label="Surface finishes" items={m.surface_finishes} variant="outline" />
              </div>
              <div className="sm:col-span-2 lg:col-span-4">
                <Chips label="Certifications" items={m.certifications} variant="outline" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Contact</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
              <Field label="Email" value={m.contact_email} />
              <Field label="WeChat" value={m.contact_wechat} />
              <Field label="Verified at" value={m.verified_at ? new Date(m.verified_at).toLocaleString() : ""} />
              <Field label="Joined" value={m.created_at ? new Date(m.created_at).toLocaleString() : ""} />
            </CardContent>
          </Card>
        </>
      ) : null}
    </main>
  );
}
