"use client";

import { useRouter } from "next/navigation";
import { ChangeEvent, FormEvent, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { confirmUpload, createProject, requestUploadUrl } from "@/lib/design/client";
import type { FileKind } from "@/lib/design/types";
import { createRFQ } from "@/lib/rfqs/client";

const CATEGORIES = ["pcb", "pcba", "cable_assembly", "enclosure", "other"] as const;
type Category = (typeof CATEGORIES)[number];

async function sha256Hex(file: File): Promise<string> {
  const digest = await crypto.subtle.digest("SHA-256", await file.arrayBuffer());
  return Array.from(new Uint8Array(digest))
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
}

function inferKind(filename: string): FileKind {
  return filename.toLowerCase().endsWith(".zip") ? "gerber" : "other";
}

export function RFQNewClient() {
  const router = useRouter();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [description, setDescription] = useState("");
  const [category, setCategory] = useState<Category>("pcba");
  const [qty, setQty] = useState("");
  const [targetDate, setTargetDate] = useState("");
  const [shippingAddress, setShippingAddress] = useState("");
  const [notes, setNotes] = useState("");
  const [files, setFiles] = useState<File[]>([]);

  const [submitting, setSubmitting] = useState(false);
  const [uploadProgress, setUploadProgress] = useState("");
  const [errors, setErrors] = useState<{ description?: string; qty?: string; general?: string }>({});

  function handleFileChange(e: ChangeEvent<HTMLInputElement>) {
    const selected = Array.from(e.target.files ?? []);
    setFiles((prev) => [...prev, ...selected]);
    e.target.value = "";
  }

  function removeFile(index: number) {
    setFiles((prev) => prev.filter((_, i) => i !== index));
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();

    const newErrors: typeof errors = {};
    if (!description.trim()) newErrors.description = "Description is required.";
    const qtyNum = parseInt(qty, 10);
    if (!qty || !Number.isFinite(qtyNum) || qtyNum < 1) newErrors.qty = "Enter a valid quantity (≥ 1).";
    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }

    setErrors({});
    setSubmitting(true);
    setUploadProgress("");

    try {
      let projectId: string | undefined;

      if (files.length > 0) {
        setUploadProgress("Creating design project…");
        const project = await createProject({
          title: description.trim().slice(0, 60),
          category,
          description: description.trim(),
        });
        projectId = project.id;

        for (let i = 0; i < files.length; i++) {
          const file = files[i];
          const contentType = file.type || "application/octet-stream";
          setUploadProgress(`Uploading ${file.name} (${i + 1}/${files.length})…`);

          const { file: created, uploadUrl } = await requestUploadUrl(projectId, {
            kind: inferKind(file.name),
            filename: file.name,
            contentType,
          });

          const put = await fetch(uploadUrl, {
            method: "PUT",
            body: file,
            headers: { "Content-Type": contentType },
          });
          if (!put.ok) throw new Error(`Upload failed for ${file.name} (${put.status}).`);

          const contentSha256 = await sha256Hex(file);
          await confirmUpload(created.id, { contentSha256, sizeBytes: file.size });
        }
        setUploadProgress("Files uploaded. Creating RFQ…");
      }

      const rfq = await createRFQ({
        queryText: description.trim(),
        qty: qtyNum,
        targetDate: targetDate || undefined,
        shippingAddress: shippingAddress || undefined,
        notes: notes || undefined,
        matchedProductIds: [],
        projectId,
      });

      router.push(`/rfqs/${encodeURIComponent(rfq.id)}`);
    } catch (err) {
      setErrors({ general: err instanceof Error ? err.message : "Submission failed." });
      setUploadProgress("");
    } finally {
      setSubmitting(false);
    }
  }

  const labelClass = "font-mono text-[11px] uppercase tracking-[0.08em] text-muted-foreground";
  const textareaClass =
    "w-full resize-none rounded-sm border border-border bg-background px-3 py-2 font-mono text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 focus:ring-offset-background";
  const selectClass =
    "w-full rounded-sm border border-border bg-background px-3 py-2 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 focus:ring-offset-background";

  return (
    <main className="mx-auto w-full max-w-[880px] px-6 py-6">
      <div className="mb-6">
        <p className="mb-1 font-mono text-xs uppercase tracking-[0.16em] text-muted-foreground">
          Buyer workspace
        </p>
        <h1 className="text-2xl">New request</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Describe the part you need. Manufacturers on the open board will respond with quotes.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Request details</CardTitle>
          <CardDescription>No catalog product selected — this goes to the open market.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="flex flex-col gap-5">
            {/* Description */}
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="rn-desc" className={labelClass}>
                Description <span className="text-destructive">*</span>
              </Label>
              <textarea
                id="rn-desc"
                rows={4}
                required
                placeholder="e.g. Arduino-compatible dev board, ATmega328P, USB-C, 3.3V/5V selectable. Include specs, standards, or compatibility requirements."
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className={textareaClass}
                disabled={submitting}
              />
              {errors.description ? (
                <p className="font-mono text-xs text-destructive">{errors.description}</p>
              ) : null}
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              {/* Category */}
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="rn-cat" className={labelClass}>
                  Category
                </Label>
                <select
                  id="rn-cat"
                  value={category}
                  onChange={(e) => setCategory(e.target.value as Category)}
                  className={selectClass}
                  disabled={submitting}
                >
                  {CATEGORIES.map((c) => (
                    <option key={c} value={c}>
                      {c.replace("_", " ")}
                    </option>
                  ))}
                </select>
              </div>

              {/* Quantity */}
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="rn-qty" className={labelClass}>
                  Quantity <span className="text-destructive">*</span>
                </Label>
                <Input
                  id="rn-qty"
                  type="number"
                  min="1"
                  step="1"
                  placeholder="100"
                  required
                  value={qty}
                  onChange={(e) => setQty(e.target.value)}
                  className="font-mono"
                  disabled={submitting}
                />
                {errors.qty ? (
                  <p className="font-mono text-xs text-destructive">{errors.qty}</p>
                ) : null}
              </div>

              {/* Target date */}
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="rn-date" className={labelClass}>
                  Target date
                </Label>
                <Input
                  id="rn-date"
                  type="date"
                  value={targetDate}
                  onChange={(e) => setTargetDate(e.target.value)}
                  className="font-mono"
                  disabled={submitting}
                />
              </div>

              {/* Shipping address */}
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="rn-addr" className={labelClass}>
                  Shipping address
                </Label>
                <Input
                  id="rn-addr"
                  type="text"
                  placeholder="San Francisco, CA, USA"
                  value={shippingAddress}
                  onChange={(e) => setShippingAddress(e.target.value)}
                  className="font-mono"
                  disabled={submitting}
                />
              </div>
            </div>

            {/* Notes */}
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="rn-notes" className={labelClass}>
                Notes (optional)
              </Label>
              <textarea
                id="rn-notes"
                rows={2}
                placeholder="Certifications, packaging, MOQ, special requirements…"
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                className={textareaClass}
                disabled={submitting}
              />
            </div>

            {/* Design files */}
            <div className="flex flex-col gap-2">
              <Label className={labelClass}>Design files (optional)</Label>
              <p className="font-mono text-[11px] text-muted-foreground">
                Attach Gerber/BOM/drawings — optional. .zip files are treated as Gerbers.
              </p>
              <div className="flex flex-wrap items-center gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => fileInputRef.current?.click()}
                  disabled={submitting}
                >
                  Add files
                </Button>
                <input
                  ref={fileInputRef}
                  type="file"
                  multiple
                  className="hidden"
                  onChange={handleFileChange}
                  disabled={submitting}
                />
              </div>
              {files.length > 0 ? (
                <ul className="mt-1 space-y-1">
                  {files.map((f, i) => (
                    <li key={i} className="flex items-center justify-between gap-2 font-mono text-xs">
                      <span className="truncate text-foreground">{f.name}</span>
                      <button
                        type="button"
                        onClick={() => removeFile(i)}
                        disabled={submitting}
                        className="shrink-0 text-muted-foreground hover:text-destructive"
                        aria-label={`Remove ${f.name}`}
                      >
                        remove
                      </button>
                    </li>
                  ))}
                </ul>
              ) : null}
            </div>

            {errors.general ? (
              <p className="font-mono text-xs text-destructive">{errors.general}</p>
            ) : null}
            {uploadProgress ? (
              <p className="font-mono text-xs text-muted-foreground">{uploadProgress}</p>
            ) : null}

            <div className="flex justify-end gap-2 border-t border-border pt-4">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => router.back()}
                disabled={submitting}
              >
                Cancel
              </Button>
              <Button type="submit" size="sm" disabled={submitting}>
                {submitting ? "Submitting…" : "Submit request"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </main>
  );
}
