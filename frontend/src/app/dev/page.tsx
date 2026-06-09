"use client";

import * as React from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { CodeId } from "@/components/ui/code-id";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Modal } from "@/components/ui/modal";
import { StatusPill } from "@/components/ui/pill";
import { RouteIndicator } from "@/components/ui/route-indicator";
import {
  Table,
  TableBody,
  TableHead,
  Td,
  Th,
  Tr,
  type SortDirection,
} from "@/components/ui/table";

/**
 * /dev — design-token + primitive showcase. Not linked from the app; exists to
 * satisfy GAU-241 (tokens render) and GAU-242 (every primitive variant renders)
 * and to spot regressions in either theme.
 */

const COLORS: { name: string; varName: string }[] = [
  { name: "background", varName: "--background" },
  { name: "card", varName: "--card" },
  { name: "foreground", varName: "--foreground" },
  { name: "muted", varName: "--muted" },
  { name: "muted-foreground", varName: "--muted-foreground" },
  { name: "primary (accent)", varName: "--primary" },
  { name: "border", varName: "--border" },
  { name: "border-strong", varName: "--border-strong" },
  { name: "marker-cn", varName: "--marker-cn" },
  { name: "marker-us", varName: "--marker-us" },
  { name: "success", varName: "--success" },
  { name: "warning", varName: "--warning" },
  { name: "destructive", varName: "--destructive" },
];

const TYPE_SCALE: { label: string; cls: string }[] = [
  { label: "h1 · 32px mono", cls: "text-[2rem]" },
  { label: "h2 · 24px mono", cls: "text-[1.5rem]" },
  { label: "h3 · 18px mono", cls: "text-[1.125rem]" },
  { label: "body · 15px sans", cls: "text-base font-sans" },
  { label: "sm · 14px sans", cls: "text-sm font-sans" },
  { label: "xs · 12px sans", cls: "text-xs font-sans" },
];

const RADII: { label: string; cls: string }[] = [
  { label: "sm", cls: "rounded-sm" },
  { label: "md (default)", cls: "rounded-md" },
  { label: "lg", cls: "rounded-lg" },
];

function Section({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section className="border-t border-border py-8">
      <h2 className="mb-5 font-mono text-xs uppercase tracking-[0.16em] text-muted-foreground">
        {title}
      </h2>
      {children}
    </section>
  );
}

export default function DevShowcasePage() {
  const [modalOpen, setModalOpen] = React.useState(false);
  const [sort, setSort] = React.useState<{ key: string; dir: SortDirection }>({
    key: "price",
    dir: "asc",
  });

  const cycleSort = (key: string) =>
    setSort((prev) =>
      prev.key !== key
        ? { key, dir: "asc" }
        : { key, dir: prev.dir === "asc" ? "desc" : prev.dir === "desc" ? null : "asc" },
    );

  const dirFor = (key: string): SortDirection =>
    sort.key === key ? sort.dir : null;

  return (
    <main className="mx-auto max-w-4xl px-6 py-10">
      <header className="mb-2 flex items-baseline justify-between">
        <h1>FIBERLANE / DEV</h1>
        <span className="font-mono text-xs uppercase tracking-[0.12em] text-muted-foreground">
          design tokens + primitives
        </span>
      </header>
      <p className="text-sm text-muted-foreground">
        Token + component reference. Toggle the app theme to verify both render.
      </p>

      <Section title="Route Indicator — the brand">
        <div className="flex flex-col gap-6">
          <RouteIndicator size="sm" />
          <RouteIndicator size="md" />
          <RouteIndicator size="lg" />
          <RouteIndicator
            size="md"
            to={{ city: "AUSTIN", coords: "30.27°N 97.74°W" }}
            showCoords
          />
        </div>
      </Section>

      <Section title="Color tokens">
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
          {COLORS.map((c) => (
            <div
              key={c.varName}
              className="flex items-center gap-3 rounded-md border border-border p-2"
            >
              <span
                className="size-9 shrink-0 rounded-sm border border-border"
                style={{ background: `var(${c.varName})` }}
              />
              <span className="font-mono text-xs leading-tight">
                {c.name}
                <br />
                <span className="text-muted-foreground">{c.varName}</span>
              </span>
            </div>
          ))}
        </div>
      </Section>

      <Section title="Type scale">
        <div className="flex flex-col gap-3">
          {TYPE_SCALE.map((t) => (
            <div key={t.label} className="flex items-baseline gap-4">
              <span className="w-40 shrink-0 font-mono text-xs text-muted-foreground">
                {t.label}
              </span>
              <span className={t.cls}>Shenzhen → San Francisco</span>
            </div>
          ))}
        </div>
      </Section>

      <Section title="Radius (max 12px — no pills/full)">
        <div className="flex gap-4">
          {RADII.map((r) => (
            <div key={r.label} className="flex flex-col items-center gap-2">
              <span className={`size-16 border border-border-strong ${r.cls}`} />
              <span className="font-mono text-xs text-muted-foreground">
                {r.label}
              </span>
            </div>
          ))}
        </div>
      </Section>

      <Section title="Buttons">
        <div className="flex flex-col gap-4">
          <div className="flex flex-wrap items-center gap-3">
            <Button>Primary</Button>
            <Button variant="secondary">Secondary</Button>
            <Button variant="outline">Outline</Button>
            <Button variant="ghost">Ghost</Button>
            <Button variant="destructive">Destructive</Button>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <Button size="sm">Small</Button>
            <Button size="default">Default</Button>
            <Button size="lg">Large</Button>
            <Button disabled>Disabled</Button>
          </div>
        </div>
      </Section>

      <Section title="Badges">
        <div className="flex flex-wrap items-center gap-3">
          <Badge>Default</Badge>
          <Badge variant="outline">Outline</Badge>
          <Badge variant="primary">Verified</Badge>
          <Badge variant="success">In stock</Badge>
          <Badge variant="warning">Low MOQ</Badge>
          <Badge variant="destructive">EOL</Badge>
        </div>
      </Section>

      <Section title="Status pills">
        <div className="flex flex-wrap items-center gap-3">
          <StatusPill tone="neutral">Draft</StatusPill>
          <StatusPill tone="info">Quoted</StatusPill>
          <StatusPill tone="success">Confirmed</StatusPill>
          <StatusPill tone="warning">Pending</StatusPill>
          <StatusPill tone="destructive">Cancelled</StatusPill>
        </div>
      </Section>

      <Section title="Code IDs">
        <div className="flex flex-col gap-3">
          <div className="flex flex-wrap items-center gap-4">
            <CodeId code="RFQ-20260429-0142-SZX" />
            <CodeId code="ORD-20260429-0087" />
            <CodeId code="SUP-SZ-0031" />
            <CodeId code="QUOTE-0142-A" />
          </div>
          <div className="flex items-center gap-2">
            <span className="font-mono text-xs text-muted-foreground">
              copyable:
            </span>
            <CodeId code="RFQ-20260429-0142-SZX" copyable />
          </div>
        </div>
      </Section>

      <Section title="Inputs">
        <div className="flex max-w-sm flex-col gap-1.5">
          <Label htmlFor="demo-input">Part number</Label>
          <Input id="demo-input" placeholder="e.g. QSFP28-100G-LR4" />
          <span className="text-xs text-muted-foreground">
            Helper text sits below in the subtle tone.
          </span>
        </div>
      </Section>

      <Section title="Card">
        <Card className="max-w-sm">
          <CardHeader>
            <CardTitle>100G QSFP28 LR4</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-2 text-sm">
            <RouteIndicator size="sm" />
            <p className="text-muted-foreground">
              1310nm · 10km · single-mode · LC duplex
            </p>
          </CardContent>
        </Card>
      </Section>

      <Section title="Table (sortable header, numeric column)">
        <Table>
          <TableHead>
            <Tr>
              <Th sortable sortDirection={dirFor("part")} onClick={() => cycleSort("part")}>
                Part
              </Th>
              <Th>Supplier</Th>
              <Th numeric sortable sortDirection={dirFor("price")} onClick={() => cycleSort("price")}>
                Price (USD)
              </Th>
              <Th numeric>Lead time</Th>
            </Tr>
          </TableHead>
          <TableBody>
            <Tr>
              <Td className="font-mono">QSFP28-100G-LR4</Td>
              <Td>Shenzhen Gigalight</Td>
              <Td numeric>$148.00</Td>
              <Td numeric>12 d</Td>
            </Tr>
            <Tr>
              <Td className="font-mono">SFP-10G-SR</Td>
              <Td>HiOSO Technology</Td>
              <Td numeric>$11.50</Td>
              <Td numeric>7 d</Td>
            </Tr>
            <Tr>
              <Td className="font-mono">QSFP-40G-SR4</Td>
              <Td>Eoptolink</Td>
              <Td numeric>$62.00</Td>
              <Td numeric>15 d</Td>
            </Tr>
          </TableBody>
        </Table>
      </Section>

      <Section title="Modal">
        <Button variant="secondary" onClick={() => setModalOpen(true)}>
          Open modal
        </Button>
        <Modal
          open={modalOpen}
          onClose={() => setModalOpen(false)}
          title="Request quote"
          description="Sends a magic-link RFQ to the supplier."
        >
          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="demo-qty">Quantity</Label>
              <Input id="demo-qty" type="number" defaultValue={100} />
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="ghost" onClick={() => setModalOpen(false)}>
                Cancel
              </Button>
              <Button onClick={() => setModalOpen(false)}>Request quote</Button>
            </div>
          </div>
        </Modal>
      </Section>
    </main>
  );
}
