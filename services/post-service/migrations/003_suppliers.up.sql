CREATE TABLE suppliers (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name              text NOT NULL,
    name_zh           text,
    city              text NOT NULL,
    country           text NOT NULL DEFAULT 'CN',
    cluster           text,                       -- 'Shenzhen', 'Dongguan', 'Wuhan'
    capabilities      text[] NOT NULL DEFAULT '{}',
    certifications    text[] NOT NULL DEFAULT '{}',
    founded_year      int,
    employees         int,
    facility_size_m2  int,
    annual_output     text,
    on_time_rate      numeric(5,2),               -- 0..100
    rating            numeric(3,2),               -- 0..5
    order_count       int NOT NULL DEFAULT 0,
    verified_at       timestamptz,
    audit_report_url  text,
    photo_url         text,
    contact_email     text,
    contact_wechat    text,
    created_at        timestamptz NOT NULL DEFAULT now()
);
