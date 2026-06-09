CREATE TABLE supplier_applications (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    payload     jsonb NOT NULL,
    status      text NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('pending','approved','rejected')),
    created_at  timestamptz NOT NULL DEFAULT now()
);
