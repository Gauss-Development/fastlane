"use client";

import Link from "next/link";
import { ChangeEvent, FormEvent, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { CodeId } from "@/components/ui/code-id";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { StatusPill, type StatusTone } from "@/components/ui/pill";
import { RouteIndicator } from "@/components/ui/route-indicator";
import { Table, TableBody, Td, Th, TableHead, Tr } from "@/components/ui/table";
import {
  acceptNda,
  confirmUpload,
  getNdaStatus,
  getProject,
  inviteManufacturer,
  listFiles,
  requestDownloadUrl,
  requestUploadUrl,
} from "@/lib/design/client";
import type { DesignFile, FileKind, ProjectStatus } from "@/lib/design/types";
import { useAuthStore } from "@/lib/stores/auth-store";

const PROJECT_TONE: Record<ProjectStatus, StatusTone> = {
  draft: "neutral",
  active: "info",
  archived: "warning",
};

const FILE_KINDS: FileKind[] = [
  "gerber",
  "bom",
  "assembly_drawing",
  "pick_place",
  "datasheet",
  "nda",
  "other",
];

function formatSize(bytes: number) {
  if (!bytes) return "—";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

async function sha256Hex(file: File): Promise<string> {
  const digest = await crypto.subtle.digest("SHA-256", await file.arrayBuffer());
  return Array.from(new Uint8Array(digest))
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
}

function FilesTable({ files }: { files: DesignFile[] }) {
  const [downloadingId, setDownloadingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function download(file: DesignFile) {
    setError(null);
    setDownloadingId(file.id);
    try {
      const { downloadUrl } = await requestDownloadUrl(file.id);
      window.open(downloadUrl, "_blank", "noopener,noreferrer");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to get download link.");
    } finally {
      setDownloadingId(null);
    }
  }

  if (files.length === 0) {
    return <p className="py-6 text-center text-sm text-muted-foreground">No files uploaded yet.</p>;
  }

  return (
    <div className="space-y-3">
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
      <Table>
        <TableHead>
          <Tr>
            <Th>Filename</Th>
            <Th>Kind</Th>
            <Th>Status</Th>
            <Th numeric>Size</Th>
            <Th />
          </Tr>
        </TableHead>
        <TableBody>
          {files.map((file) => (
            <Tr key={file.id}>
              <Td className="font-mono text-xs">{file.filename}</Td>
              <Td>
                <Badge>{file.kind}</Badge>
              </Td>
              <Td>
                <StatusPill tone={file.status === "committed" ? "success" : "neutral"}>{file.status}</StatusPill>
              </Td>
              <Td numeric>{formatSize(file.size_bytes)}</Td>
              <Td numeric>
                <Button
                  size="sm"
                  variant="outline"
                  disabled={file.status !== "committed" || downloadingId === file.id}
                  onClick={() => download(file)}
                >
                  {downloadingId === file.id ? "…" : "Download"}
                </Button>
              </Td>
            </Tr>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function UploadCard({ projectId }: { projectId: string }) {
  const queryClient = useQueryClient();
  const [kind, setKind] = useState<FileKind>("gerber");
  const [status, setStatus] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function handleFile(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file || busy) return;

    setBusy(true);
    setError(null);
    const contentType = file.type || "application/octet-stream";
    try {
      setStatus("Requesting upload URL…");
      const { file: created, uploadUrl } = await requestUploadUrl(projectId, {
        kind,
        filename: file.name,
        contentType,
      });

      setStatus("Uploading…");
      const put = await fetch(uploadUrl, {
        method: "PUT",
        body: file,
        headers: { "Content-Type": contentType },
      });
      if (!put.ok) throw new Error(`Upload failed (${put.status}).`);

      setStatus("Confirming…");
      const contentSha256 = await sha256Hex(file);
      await confirmUpload(created.id, { contentSha256, sizeBytes: file.size });

      queryClient.invalidateQueries({ queryKey: ["project-files", projectId] });
      setStatus(`Uploaded ${file.name}.`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Upload failed.");
      setStatus(null);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
      <div className="space-y-2">
        <Label htmlFor="upload-kind">Kind</Label>
        <select
          id="upload-kind"
          value={kind}
          disabled={busy}
          onChange={(e) => setKind(e.target.value as FileKind)}
          className="h-9 w-full rounded-sm border border-input bg-input-background px-3 text-sm sm:w-48"
        >
          {FILE_KINDS.map((k) => (
            <option key={k} value={k}>
              {k}
            </option>
          ))}
        </select>
      </div>
      <div className="space-y-2">
        <Label htmlFor="upload-file">File</Label>
        <Input id="upload-file" type="file" disabled={busy} onChange={handleFile} className="cursor-pointer" />
      </div>
      <div className="flex-1 text-sm">
        {status ? <span className="text-muted-foreground">{status}</span> : null}
        {error ? <span className="text-destructive">{error}</span> : null}
      </div>
    </div>
  );
}

function InviteCard({ projectId }: { projectId: string }) {
  const [manufacturerId, setManufacturerId] = useState("");
  const [ok, setOk] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: () => inviteManufacturer(projectId, { manufacturerId: manufacturerId.trim() }),
    onSuccess: () => {
      setOk(true);
      setError(null);
      setManufacturerId("");
    },
    onError: (err) => {
      setOk(false);
      setError(err instanceof Error ? err.message : "Failed to invite manufacturer.");
    },
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (mutation.isPending || !manufacturerId.trim()) return;
    setOk(false);
    setError(null);
    mutation.mutate();
  }

  return (
    <form className="flex flex-col gap-3 sm:flex-row sm:items-end" onSubmit={handleSubmit}>
      <div className="flex-1 space-y-2">
        <Label htmlFor="invite-manufacturer">Manufacturer ID</Label>
        <Input
          id="invite-manufacturer"
          value={manufacturerId}
          onChange={(e) => setManufacturerId(e.target.value)}
          placeholder="SUP-…"
          className="font-mono"
        />
      </div>
      <Button type="submit" disabled={mutation.isPending || !manufacturerId.trim()}>
        {mutation.isPending ? "Inviting…" : "Invite"}
      </Button>
      <div className="text-sm sm:self-center">
        {ok ? <span className="text-success">Invited.</span> : null}
        {error ? <span className="text-destructive">{error}</span> : null}
      </div>
    </form>
  );
}

function OwnerView({ projectId }: { projectId: string }) {
  const filesQuery = useQuery({
    queryKey: ["project-files", projectId],
    queryFn: () => listFiles(projectId),
  });

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Design files</CardTitle>
          <CardDescription>Upload gerbers, BOMs, and drawings. Invited manufacturers download these under NDA.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <UploadCard projectId={projectId} />
          {filesQuery.isLoading ? (
            <p className="py-6 text-center font-mono text-sm text-muted-foreground">Loading…</p>
          ) : null}
          {filesQuery.error ? (
            <p className="py-6 text-center text-sm text-destructive">{(filesQuery.error as Error).message}</p>
          ) : null}
          {filesQuery.data ? <FilesTable files={filesQuery.data.files} /> : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Invite manufacturer</CardTitle>
          <CardDescription>Invite a verified manufacturer to review this project. They must accept the NDA before seeing files.</CardDescription>
        </CardHeader>
        <CardContent>
          <InviteCard projectId={projectId} />
        </CardContent>
      </Card>
    </>
  );
}

function ManufacturerView({ projectId }: { projectId: string }) {
  const queryClient = useQueryClient();
  const ndaQuery = useQuery({
    queryKey: ["project-nda", projectId],
    queryFn: () => getNdaStatus(projectId),
    retry: false,
  });
  const filesQuery = useQuery({
    queryKey: ["project-files", projectId],
    queryFn: () => listFiles(projectId),
    enabled: ndaQuery.data?.status === "accepted",
  });

  const acceptMutation = useMutation({
    mutationFn: () => acceptNda(projectId, { ndaVersion: "v1" }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["project-nda", projectId] });
      queryClient.invalidateQueries({ queryKey: ["project-files", projectId] });
    },
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>NDA & files</CardTitle>
        <CardDescription>Accept the NDA to unlock the owner&apos;s design files.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {ndaQuery.isLoading ? (
          <p className="py-6 text-center font-mono text-sm text-muted-foreground">Loading…</p>
        ) : null}

        {ndaQuery.error ? (
          <p className="py-6 text-center text-sm text-muted-foreground">
            You have not been invited to this project.
          </p>
        ) : null}

        {ndaQuery.data?.status === "pending" ? (
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">
              You have been invited. Accept the NDA (version {ndaQuery.data.nda_version || "v1"}) to view files.
            </p>
            {acceptMutation.error ? (
              <p className="text-sm text-destructive">{(acceptMutation.error as Error).message}</p>
            ) : null}
            <Button onClick={() => acceptMutation.mutate()} disabled={acceptMutation.isPending}>
              {acceptMutation.isPending ? "Accepting…" : "Accept NDA"}
            </Button>
          </div>
        ) : null}

        {ndaQuery.data?.status === "accepted" ? (
          <>
            <StatusPill tone="success">NDA accepted</StatusPill>
            {filesQuery.isLoading ? (
              <p className="py-6 text-center font-mono text-sm text-muted-foreground">Loading files…</p>
            ) : null}
            {filesQuery.error ? (
              <p className="py-6 text-center text-sm text-destructive">{(filesQuery.error as Error).message}</p>
            ) : null}
            {filesQuery.data ? <FilesTable files={filesQuery.data.files} /> : null}
          </>
        ) : null}
      </CardContent>
    </Card>
  );
}

export function ProjectDetailClient({ projectId }: { projectId: string }) {
  const user = useAuthStore((s) => s.user);
  const projectQuery = useQuery({
    queryKey: ["project", projectId],
    queryFn: () => getProject(projectId),
  });

  const project = projectQuery.data;
  const isManufacturer = user?.role === "manufacturer";

  return (
    <main className="mx-auto flex w-full max-w-[1200px] flex-col gap-6 px-6 py-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.18em] text-muted-foreground">Project</p>
          <div className="mt-2">
            <CodeId code={projectId} size="lg" copyable />
          </div>
        </div>
        <RouteIndicator size="sm" />
      </div>

      {projectQuery.isLoading ? (
        <Card>
          <CardContent className="py-10 text-center font-mono text-sm text-muted-foreground">Loading…</CardContent>
        </Card>
      ) : null}

      {projectQuery.error ? (
        <Card className="border-destructive/50">
          <CardHeader>
            <CardTitle>Could not load project</CardTitle>
            <CardDescription>{(projectQuery.error as Error).message}</CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/projects" className="font-mono text-xs uppercase tracking-[0.08em] text-primary hover:underline">
              ← Back to projects
            </Link>
          </CardContent>
        </Card>
      ) : null}

      {project ? (
        <Card>
          <CardHeader>
            <div className="flex flex-wrap items-center justify-between gap-3">
              <CardTitle>{project.title}</CardTitle>
              <StatusPill tone={PROJECT_TONE[project.status] ?? "neutral"}>{project.status}</StatusPill>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-wrap items-center gap-3">
              <Badge>{project.category}</Badge>
            </div>
            {project.description ? (
              <p className="text-sm text-muted-foreground">{project.description}</p>
            ) : null}
          </CardContent>
        </Card>
      ) : null}

      {project ? (
        isManufacturer ? (
          <ManufacturerView projectId={projectId} />
        ) : (
          <OwnerView projectId={projectId} />
        )
      ) : null}
    </main>
  );
}
