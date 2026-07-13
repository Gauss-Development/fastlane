CREATE TABLE IF NOT EXISTS projects (
    id              text PRIMARY KEY,
    owner_id        text NOT NULL,
    title           text NOT NULL,
    description     text,
    category        text NOT NULL DEFAULT 'pcba' CHECK (category IN ('pcb','pcba','cable_assembly','enclosure','other')),
    status          text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','active','archived')),
    owner_email     text,
    owner_company   text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS design_files (
    id              text PRIMARY KEY,
    project_id      text NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    kind            text NOT NULL CHECK (kind IN ('gerber','bom','assembly_drawing','pick_place','datasheet','nda','other')),
    filename        text NOT NULL,
    version         int  NOT NULL DEFAULT 1,
    content_sha256  text,
    object_key      text NOT NULL,
    size_bytes      bigint DEFAULT 0,
    content_type    text,
    uploaded_by     text NOT NULL,
    status          text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','committed')),
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS design_files_project_id_idx ON design_files (project_id);

CREATE TABLE IF NOT EXISTS ndas (
    id               text PRIMARY KEY,
    project_id       text NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    manufacturer_id  text NOT NULL,
    status           text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','accepted')),
    nda_version      text,
    accepted_ip      text,
    accepted_at      timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now(),
    UNIQUE (project_id, manufacturer_id)
);

CREATE INDEX IF NOT EXISTS ndas_project_id_idx ON ndas (project_id);

-- Sequences for human-readable IDs (grow past 4 digits instead of wrapping).
CREATE SEQUENCE IF NOT EXISTS project_id_seq;
CREATE SEQUENCE IF NOT EXISTS file_id_seq;
