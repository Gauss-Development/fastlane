"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { createManufacturer, getMyManufacturer, updateManufacturer } from "@/lib/manufacturers/client";
import type { Manufacturer } from "@/lib/manufacturers/types";
import { useAuthStore } from "@/lib/stores/auth-store";

type FormState = {
  name: string;
  name_zh: string;
  city: string;
  cluster: string;
  description: string;
  website: string;
  service_types: string;
  assembly_types: string;
  min_layers: string;
  max_layers: string;
  materials: string;
  surface_finishes: string;
  min_order_qty: string;
  max_order_qty: string;
  lead_time_days: string;
  monthly_capacity: string;
  smallest_package: string;
  certifications: string;
  contact_email: string;
  contact_wechat: string;
};

const EMPTY: FormState = {
  name: "",
  name_zh: "",
  city: "",
  cluster: "",
  description: "",
  website: "",
  service_types: "",
  assembly_types: "",
  min_layers: "",
  max_layers: "",
  materials: "",
  surface_finishes: "",
  min_order_qty: "",
  max_order_qty: "",
  lead_time_days: "",
  monthly_capacity: "",
  smallest_package: "",
  certifications: "",
  contact_email: "",
  contact_wechat: "",
};

const csv = (s: string) => s.split(",").map((x) => x.trim()).filter(Boolean);
const num = (s: string) => {
  const n = Number(s);
  return Number.isFinite(n) ? n : 0;
};

function fromManufacturer(m: Manufacturer): FormState {
  return {
    name: m.name ?? "",
    name_zh: m.name_zh ?? "",
    city: m.city ?? "",
    cluster: m.cluster ?? "",
    description: m.description ?? "",
    website: m.website ?? "",
    service_types: (m.service_types ?? []).join(", "),
    assembly_types: (m.assembly_types ?? []).join(", "),
    min_layers: m.min_layers ? String(m.min_layers) : "",
    max_layers: m.max_layers ? String(m.max_layers) : "",
    materials: (m.materials ?? []).join(", "),
    surface_finishes: (m.surface_finishes ?? []).join(", "),
    min_order_qty: m.min_order_qty ? String(m.min_order_qty) : "",
    max_order_qty: m.max_order_qty ? String(m.max_order_qty) : "",
    lead_time_days: m.lead_time_days ? String(m.lead_time_days) : "",
    monthly_capacity: m.monthly_capacity ? String(m.monthly_capacity) : "",
    smallest_package: m.smallest_package ?? "",
    certifications: (m.certifications ?? []).join(", "),
    contact_email: m.contact_email ?? "",
    contact_wechat: m.contact_wechat ?? "",
  };
}

function toBody(f: FormState): Partial<Manufacturer> {
  return {
    name: f.name,
    name_zh: f.name_zh,
    city: f.city,
    cluster: f.cluster,
    description: f.description,
    website: f.website,
    service_types: csv(f.service_types),
    assembly_types: csv(f.assembly_types),
    min_layers: num(f.min_layers),
    max_layers: num(f.max_layers),
    materials: csv(f.materials),
    surface_finishes: csv(f.surface_finishes),
    min_order_qty: num(f.min_order_qty),
    max_order_qty: num(f.max_order_qty),
    lead_time_days: num(f.lead_time_days),
    monthly_capacity: num(f.monthly_capacity),
    smallest_package: f.smallest_package,
    certifications: csv(f.certifications),
    contact_email: f.contact_email,
    contact_wechat: f.contact_wechat,
  };
}

function TextField({
  id,
  label,
  value,
  onChange,
  type,
  placeholder,
}: {
  id: string;
  label: string;
  value: string;
  onChange: (v: string) => void;
  type?: string;
  placeholder?: string;
}) {
  return (
    <div className="space-y-2">
      <Label htmlFor={id}>{label}</Label>
      <Input id={id} type={type} value={value} placeholder={placeholder} onChange={(e) => onChange(e.target.value)} />
    </div>
  );
}

export function ManufacturerProfileClient() {
  const user = useAuthStore((s) => s.user);
  const queryClient = useQueryClient();
  const [form, setForm] = useState<FormState>(EMPTY);
  const [prefilledId, setPrefilledId] = useState<string | null>(null);
  const [ok, setOk] = useState(false);

  const isManufacturer = user?.role === "manufacturer";

  const query = useQuery({
    queryKey: ["my-manufacturer"],
    queryFn: getMyManufacturer,
    enabled: isManufacturer,
    retry: false,
  });

  const existing = query.data;
  const existingId = existing?.id;

  // Prefill during render when a freshly-loaded profile hasn't been synced yet.
  // (Storing info from previous renders — https://react.dev/learn/you-might-not-need-an-effect)
  if (existing && existing.id !== prefilledId) {
    setForm(fromManufacturer(existing));
    setPrefilledId(existing.id);
  }

  const set = (k: keyof FormState) => (v: string) => setForm((f) => ({ ...f, [k]: v }));

  const mutation = useMutation({
    mutationFn: (body: Partial<Manufacturer>) =>
      existingId ? updateManufacturer(existingId, body) : createManufacturer(body),
    onSuccess: () => {
      setOk(true);
      queryClient.invalidateQueries({ queryKey: ["my-manufacturer"] });
    },
    onError: () => setOk(false),
  });

  if (!isManufacturer) {
    return (
      <main className="mx-auto w-full max-w-[960px] px-6 py-6">
        <Card>
          <CardHeader>
            <CardTitle>My Profile</CardTitle>
            <CardDescription>This page is for manufacturer accounts.</CardDescription>
          </CardHeader>
        </Card>
      </main>
    );
  }

  if (query.isLoading) {
    return (
      <main className="mx-auto w-full max-w-[960px] px-6 py-6">
        <p className="py-8 text-center font-mono text-sm text-muted-foreground">Loading profile…</p>
      </main>
    );
  }

  const mode = existingId ? "edit" : "create";

  return (
    <main className="mx-auto flex w-full max-w-[960px] flex-col gap-6 px-6 py-6">
      <div>
        <p className="font-mono text-xs uppercase tracking-[0.18em] text-muted-foreground">My Profile</p>
        <h1 className="mt-2 text-lg">{mode === "edit" ? "Edit manufacturer profile" : "Create manufacturer profile"}</h1>
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          setOk(false);
          mutation.mutate(toBody(form));
        }}
        className="flex flex-col gap-6"
      >
        <Card>
          <CardHeader>
            <CardTitle>Identity</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-2">
            <TextField id="p-name" label="Name" value={form.name} onChange={set("name")} placeholder="Shenzhen Optics Co." />
            <TextField id="p-name-zh" label="Name (中文)" value={form.name_zh} onChange={set("name_zh")} />
            <TextField id="p-city" label="City" value={form.city} onChange={set("city")} placeholder="Shenzhen" />
            <TextField id="p-cluster" label="Cluster" value={form.cluster} onChange={set("cluster")} placeholder="Bao'an / Longgang" />
            <TextField id="p-website" label="Website" value={form.website} onChange={set("website")} placeholder="https://" />
            <div className="space-y-2 sm:col-span-2">
              <Label htmlFor="p-description">Description</Label>
              <textarea
                id="p-description"
                rows={4}
                className="w-full rounded-sm px-3 py-2 text-sm"
                value={form.description}
                onChange={(e) => set("description")(e.target.value)}
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Capabilities</CardTitle>
            <CardDescription>Comma-separated lists for multi-value fields.</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-2">
            <TextField
              id="p-service-types"
              label="Service types"
              value={form.service_types}
              onChange={set("service_types")}
              placeholder="pcb_fab, pcba, cable_assembly, enclosure, box_build"
            />
            <TextField
              id="p-assembly-types"
              label="Assembly types"
              value={form.assembly_types}
              onChange={set("assembly_types")}
              placeholder="smt, tht, mixed"
            />
            <TextField id="p-min-layers" label="Min layers" type="number" value={form.min_layers} onChange={set("min_layers")} />
            <TextField id="p-max-layers" label="Max layers" type="number" value={form.max_layers} onChange={set("max_layers")} />
            <TextField id="p-materials" label="Materials" value={form.materials} onChange={set("materials")} placeholder="FR-4, Rogers, aluminum" />
            <TextField
              id="p-surface-finishes"
              label="Surface finishes"
              value={form.surface_finishes}
              onChange={set("surface_finishes")}
              placeholder="ENIG, HASL, OSP"
            />
            <TextField id="p-min-order-qty" label="Min order qty" type="number" value={form.min_order_qty} onChange={set("min_order_qty")} />
            <TextField id="p-max-order-qty" label="Max order qty" type="number" value={form.max_order_qty} onChange={set("max_order_qty")} />
            <TextField id="p-lead-time" label="Lead time (days)" type="number" value={form.lead_time_days} onChange={set("lead_time_days")} />
            <TextField id="p-capacity" label="Monthly capacity" type="number" value={form.monthly_capacity} onChange={set("monthly_capacity")} />
            <TextField id="p-package" label="Smallest package" value={form.smallest_package} onChange={set("smallest_package")} placeholder="0201" />
            <TextField
              id="p-certifications"
              label="Certifications"
              value={form.certifications}
              onChange={set("certifications")}
              placeholder="ISO9001, UL, RoHS"
            />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Contact</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-2">
            <TextField id="p-email" label="Contact email" type="email" value={form.contact_email} onChange={set("contact_email")} />
            <TextField id="p-wechat" label="Contact WeChat" value={form.contact_wechat} onChange={set("contact_wechat")} />
          </CardContent>
        </Card>

        <div className="flex flex-wrap items-center gap-3">
          <Button type="submit" disabled={mutation.isPending}>
            {mutation.isPending ? "Saving…" : mode === "edit" ? "Save changes" : "Create profile"}
          </Button>
          {ok ? (
            <p className="font-mono text-xs uppercase tracking-[0.08em] text-success">Saved</p>
          ) : null}
          {mutation.error ? (
            <p className="text-sm text-destructive">{(mutation.error as Error).message}</p>
          ) : null}
        </div>
      </form>
    </main>
  );
}
