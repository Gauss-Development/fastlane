"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { CodeId } from "@/components/ui/code-id";
import { StatusPill, type StatusTone } from "@/components/ui/pill";
import { Table, TableBody, Td, Th, TableHead, Tr } from "@/components/ui/table";
import { listRFQs } from "@/lib/rfqs/client";
import type { RFQStatus } from "@/lib/rfqs/types";

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

export function RFQListClient() {
  const rfqsQuery = useQuery({
    queryKey: ["rfqs"],
    queryFn: () => listRFQs({ limit: 50 }),
  });

  return (
    <main className="mx-auto w-full max-w-[1200px] px-6 py-6">
      <Card>
        <CardHeader>
          <CardTitle>RFQs</CardTitle>
          <CardDescription>
            Quote requests sent to verified suppliers. Suppliers respond through magic links — no account needed on their side.
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
              <p className="text-sm text-muted-foreground">
                No RFQs yet. Run a search and hit <span className="font-mono">Quote →</span> on a result.
              </p>
              <Link href="/dashboard" className="mt-3 inline-block font-mono text-xs uppercase tracking-[0.08em] text-primary hover:underline">
                Go to search →
              </Link>
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
                </Tr>
              </TableHead>
              <TableBody>
                {rfqsQuery.data.rfqs.map((rfq) => (
                  <Tr key={rfq.id}>
                    <Td>
                      <Link href={`/rfqs/${encodeURIComponent(rfq.id)}`} className="hover:underline">
                        <CodeId code={rfq.id} size="sm" />
                      </Link>
                    </Td>
                    <Td className="max-w-[360px] truncate font-mono text-xs">{rfq.query_text}</Td>
                    <Td numeric>{rfq.qty || "—"}</Td>
                    <Td>
                      <StatusPill tone={STATUS_TONE[rfq.status] ?? "neutral"}>{rfq.status}</StatusPill>
                    </Td>
                    <Td numeric>{age(rfq.created_at)}</Td>
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
