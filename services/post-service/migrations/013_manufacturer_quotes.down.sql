ALTER TABLE quotes DROP CONSTRAINT IF EXISTS quotes_party_ck;
ALTER TABLE quotes DROP COLUMN IF EXISTS manufacturer_id;
-- best-effort: re-add NOT NULL only if no NULLs exist
ALTER TABLE quotes ALTER COLUMN supplier_id SET NOT NULL;
