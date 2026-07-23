-- manufacturer_id stores the catalog manufacturer's formatted id (MFR-YYYYMMDD-NNNN),
-- not a uuid — mirror supplier_id's "store the referenced entity's own id" semantics.
-- (Migration 013 wrongly typed it uuid; catalog manufacturer ids are text.)
ALTER TABLE quotes ALTER COLUMN manufacturer_id TYPE text USING manufacturer_id::text;
