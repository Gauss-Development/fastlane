-- Fiberlane: role enum + company name.
-- Existing users default to 'buyer' (the dominant Fiberlane persona).
-- Suppliers are inserted with role='supplier' and no password_hash — they
-- never log in directly; access is gated by magic-link tokens.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS role text NOT NULL DEFAULT 'buyer'
        CHECK (role IN ('buyer','supplier','admin'));

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS company text;

CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
