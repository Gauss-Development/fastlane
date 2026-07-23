ALTER TABLE quotes ALTER COLUMN manufacturer_id TYPE uuid USING manufacturer_id::uuid;
