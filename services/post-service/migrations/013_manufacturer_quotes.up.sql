ALTER TABLE quotes ALTER COLUMN supplier_id DROP NOT NULL;
ALTER TABLE quotes ADD COLUMN manufacturer_id uuid;
ALTER TABLE quotes ADD CONSTRAINT quotes_party_ck
    CHECK (supplier_id IS NOT NULL OR manufacturer_id IS NOT NULL);
