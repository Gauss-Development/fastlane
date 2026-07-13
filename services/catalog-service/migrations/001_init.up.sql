CREATE TABLE IF NOT EXISTS manufacturers (
    id                  text PRIMARY KEY,
    user_id             text NOT NULL UNIQUE,
    name                text NOT NULL,
    name_zh             text,
    city                text,
    country             text NOT NULL DEFAULT 'CN',
    cluster             text,
    description         text,
    website             text,
    service_types       text[] NOT NULL DEFAULT '{}',
    assembly_types      text[] NOT NULL DEFAULT '{}',
    min_layers          int  NOT NULL DEFAULT 0,
    max_layers          int  NOT NULL DEFAULT 0,
    materials           text[] NOT NULL DEFAULT '{}',
    surface_finishes    text[] NOT NULL DEFAULT '{}',
    min_order_qty       int  NOT NULL DEFAULT 0,
    max_order_qty       int  NOT NULL DEFAULT 0,
    lead_time_days      int  NOT NULL DEFAULT 0,
    monthly_capacity    int  NOT NULL DEFAULT 0,
    smallest_package    text,
    certifications      text[] NOT NULL DEFAULT '{}',
    verified            boolean NOT NULL DEFAULT false,
    verified_at         timestamptz,
    rating              double precision NOT NULL DEFAULT 0,
    order_count         int  NOT NULL DEFAULT 0,
    on_time_rate        double precision NOT NULL DEFAULT 0,
    contact_email       text,
    contact_wechat      text,
    status              text NOT NULL DEFAULT 'active' CHECK (status IN ('active','pending','archived')),
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS manufacturers_cluster_idx  ON manufacturers (cluster);
CREATE INDEX IF NOT EXISTS manufacturers_verified_idx ON manufacturers (verified);

CREATE SEQUENCE IF NOT EXISTS manufacturer_id_seq;
