"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { CodeId } from "@/components/ui/code-id";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Modal } from "@/components/ui/modal";
import { StatusPill, type StatusTone } from "@/components/ui/pill";
import { RouteIndicator } from "@/components/ui/route-indicator";
import { Table, TableBody, Td, Th, TableHead, Tr } from "@/components/ui/table";
import { createProject, listProjects } from "@/lib/design/client";
import type { ProjectStatus } from "@/lib/design/types";

const STATUS_TONE: Record<ProjectStatus, StatusTone> = {
  draft: "neutral",
  active: "info",
  archived: "warning",
};

const CATEGORIES = ["pcb", "pcba", "cable_assembly", "enclosure", "other"] as const;

function NewProjectModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: createProject,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["projects"] });
      close();
    },
    onError: (err) => setError(err instanceof Error ? err.message : "Failed to create project."),
  });

  function close() {
    setError(null);
    mutation.reset();
    onClose();
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (mutation.isPending) return;
    const form = new FormData(event.currentTarget);
    const title = String(form.get("title") ?? "").trim();
    if (!title) {
      setError("Title is required.");
      return;
    }
    setError(null);
    mutation.mutate({
      title,
      description: String(form.get("description") ?? "").trim(),
      category: String(form.get("category") ?? "pcb"),
    });
  }

  return (
    <Modal
      open={open}
      onClose={close}
      title="New project"
      description="Create a design project. Add files and invite verified manufacturers once it exists."
    >
      <form className="space-y-4" onSubmit={handleSubmit}>
        <div className="space-y-2">
          <Label htmlFor="project-title">Title</Label>
          <Input id="project-title" name="title" required placeholder="4-layer control board rev A" />
        </div>
        <div className="space-y-2">
          <Label htmlFor="project-category">Category</Label>
          <select
            id="project-category"
            name="category"
            defaultValue="pcb"
            className="h-9 w-full rounded-sm border border-input bg-input-background px-3 text-sm"
          >
            {CATEGORIES.map((c) => (
              <option key={c} value={c}>
                {c}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-2">
          <Label htmlFor="project-description">Description</Label>
          <textarea
            id="project-description"
            name="description"
            rows={4}
            className="w-full rounded-sm border border-input bg-input-background px-3 py-2 text-sm"
            placeholder="Layer count, materials, quantities, and any compliance constraints."
          />
        </div>
        {error ? <p className="text-sm text-destructive">{error}</p> : null}
        <div className="flex gap-2">
          <Button type="submit" disabled={mutation.isPending}>
            {mutation.isPending ? "Creating..." : "Create Project"}
          </Button>
          <Button type="button" variant="outline" onClick={close} disabled={mutation.isPending}>
            Cancel
          </Button>
        </div>
      </form>
    </Modal>
  );
}

export function ProjectsListClient() {
  const [modalOpen, setModalOpen] = useState(false);
  const projectsQuery = useQuery({
    queryKey: ["projects"],
    queryFn: () => listProjects({ limit: 50 }),
  });

  return (
    <main className="mx-auto w-full max-w-[1200px] px-6 py-6">
      <div className="mb-6 flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.18em] text-muted-foreground">Projects</p>
          <RouteIndicator size="sm" className="mt-2" />
        </div>
        <Button onClick={() => setModalOpen(true)}>New Project</Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>My projects</CardTitle>
          <CardDescription>
            Design projects you own. Share files with verified manufacturers under NDA and collect quotes.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {projectsQuery.isLoading ? (
            <p className="py-8 text-center font-mono text-sm text-muted-foreground">Loading…</p>
          ) : null}

          {projectsQuery.error ? (
            <p className="py-8 text-center text-sm text-destructive">
              {(projectsQuery.error as Error).message}
            </p>
          ) : null}

          {projectsQuery.data && projectsQuery.data.projects.length === 0 ? (
            <div className="py-10 text-center">
              <p className="text-sm text-muted-foreground">
                No projects yet. Hit <span className="font-mono">New Project</span> to create one.
              </p>
            </div>
          ) : null}

          {projectsQuery.data && projectsQuery.data.projects.length > 0 ? (
            <Table>
              <TableHead>
                <Tr>
                  <Th>Project</Th>
                  <Th>Title</Th>
                  <Th>Category</Th>
                  <Th>Status</Th>
                  <Th>Created</Th>
                </Tr>
              </TableHead>
              <TableBody>
                {projectsQuery.data.projects.map((project) => (
                  <Tr key={project.id}>
                    <Td>
                      <Link href={`/projects/${encodeURIComponent(project.id)}`}>
                        <CodeId code={project.id} size="sm" />
                      </Link>
                    </Td>
                    <Td>
                      <Link href={`/projects/${encodeURIComponent(project.id)}`} className="hover:text-primary">
                        {project.title}
                      </Link>
                    </Td>
                    <Td>
                      <Badge>{project.category}</Badge>
                    </Td>
                    <Td>
                      <StatusPill tone={STATUS_TONE[project.status] ?? "neutral"}>{project.status}</StatusPill>
                    </Td>
                    <Td className="font-mono text-xs">
                      {project.created_at ? new Date(project.created_at).toLocaleDateString() : "—"}
                    </Td>
                  </Tr>
                ))}
              </TableBody>
            </Table>
          ) : null}
        </CardContent>
      </Card>

      <NewProjectModal open={modalOpen} onClose={() => setModalOpen(false)} />
    </main>
  );
}
