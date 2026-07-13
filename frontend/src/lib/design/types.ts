export type ProjectStatus = "draft" | "active" | "archived";
export type FileKind =
  | "gerber"
  | "bom"
  | "assembly_drawing"
  | "pick_place"
  | "datasheet"
  | "nda"
  | "other";
export type FileStatus = "pending" | "committed";
export type NDAStatus = "pending" | "accepted";

export interface Project {
  id: string;
  owner_id?: string;
  owner_email?: string;
  owner_company?: string;
  title: string;
  description: string;
  category: string;
  status: ProjectStatus;
  created_at: string;
  updated_at: string;
}

export interface DesignFile {
  id: string;
  project_id: string;
  kind: FileKind;
  version: number;
  status: FileStatus;
  filename: string;
  content_type: string;
  object_key: string;
  size_bytes: number;
  content_sha256: string;
  created_at: string;
}

export interface NDA {
  project_id: string;
  manufacturer_id: string;
  status: NDAStatus;
  nda_version: string;
  accepted_ip: string;
  accepted_at: string;
}

export interface ListProjectsResponse {
  projects: Project[];
  total: number;
}

export interface ListFilesResponse {
  files: DesignFile[];
}

export interface UploadURLResponse {
  file: DesignFile;
  uploadUrl: string;
  objectKey: string;
  expiresIn: number;
}

export interface DownloadURLResponse {
  downloadUrl: string;
  filename: string;
  expiresIn: number;
}
