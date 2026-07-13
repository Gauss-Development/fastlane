"use client";

import { BFF_BASE_URL } from "@/lib/auth/client-constants";
import { authenticatedFetch } from "@/lib/auth/client-api";
import type {
  DesignFile,
  DownloadURLResponse,
  FileKind,
  ListFilesResponse,
  ListProjectsResponse,
  NDA,
  Project,
  UploadURLResponse,
} from "@/lib/design/types";

export async function listProjects(params?: {
  status?: string;
  limit?: number;
  offset?: number;
}): Promise<ListProjectsResponse> {
  const query = new URLSearchParams();
  if (params?.status) query.set("status", params.status);
  if (params?.limit) query.set("limit", String(params.limit));
  if (params?.offset) query.set("offset", String(params.offset));
  const suffix = query.size > 0 ? `?${query.toString()}` : "";
  return authenticatedFetch<ListProjectsResponse>(`${BFF_BASE_URL}/projects${suffix}`);
}

export async function getProject(id: string): Promise<Project> {
  return authenticatedFetch<Project>(`${BFF_BASE_URL}/projects/${encodeURIComponent(id)}`);
}

export async function createProject(params: {
  title: string;
  description: string;
  category: string;
}): Promise<Project> {
  return authenticatedFetch<Project>(`${BFF_BASE_URL}/projects`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      title: params.title,
      description: params.description,
      category: params.category,
    }),
  });
}

export async function requestUploadUrl(
  projectId: string,
  params: { kind: FileKind; filename: string; contentType: string },
): Promise<UploadURLResponse> {
  return authenticatedFetch<UploadURLResponse>(
    `${BFF_BASE_URL}/projects/${encodeURIComponent(projectId)}/files/upload-url`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        kind: params.kind,
        filename: params.filename,
        content_type: params.contentType,
      }),
    },
  );
}

export async function confirmUpload(
  fileId: string,
  params: { contentSha256: string; sizeBytes: number },
): Promise<DesignFile> {
  return authenticatedFetch<DesignFile>(`${BFF_BASE_URL}/files/${encodeURIComponent(fileId)}/confirm`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      content_sha256: params.contentSha256,
      size_bytes: params.sizeBytes,
    }),
  });
}

export async function listFiles(projectId: string): Promise<ListFilesResponse> {
  return authenticatedFetch<ListFilesResponse>(
    `${BFF_BASE_URL}/projects/${encodeURIComponent(projectId)}/files`,
  );
}

export async function requestDownloadUrl(fileId: string): Promise<DownloadURLResponse> {
  return authenticatedFetch<DownloadURLResponse>(
    `${BFF_BASE_URL}/files/${encodeURIComponent(fileId)}/download-url`,
  );
}

export async function acceptNda(projectId: string, params: { ndaVersion: string }): Promise<NDA> {
  return authenticatedFetch<NDA>(
    `${BFF_BASE_URL}/projects/${encodeURIComponent(projectId)}/nda/accept`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ nda_version: params.ndaVersion }),
    },
  );
}

export async function getNdaStatus(projectId: string): Promise<NDA> {
  return authenticatedFetch<NDA>(`${BFF_BASE_URL}/projects/${encodeURIComponent(projectId)}/nda`);
}

export async function inviteManufacturer(
  projectId: string,
  params: { manufacturerId: string },
): Promise<void> {
  return authenticatedFetch<void>(
    `${BFF_BASE_URL}/projects/${encodeURIComponent(projectId)}/invite`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ manufacturer_id: params.manufacturerId }),
    },
  );
}
