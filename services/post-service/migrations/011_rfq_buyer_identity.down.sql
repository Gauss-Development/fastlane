ALTER TABLE rfqs
    DROP COLUMN IF EXISTS buyer_email,
    DROP COLUMN IF EXISTS buyer_company;

DROP SEQUENCE IF EXISTS rfq_id_seq;
DROP SEQUENCE IF EXISTS quote_id_seq;
