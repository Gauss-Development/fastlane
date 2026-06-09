-- Buyer identity is denormalized onto the RFQ at creation time: the users
-- table lives in user-service's DB, but supplier emails and the magic-link
-- page need buyer context without a cross-service call.
ALTER TABLE rfqs
    ADD COLUMN buyer_email   text,
    ADD COLUMN buyer_company text;

-- Global sequences feed the NNNN segment of formatted ids
-- (RFQ-YYYYMMDD-NNNN-SZX / QUOTE-YYYYMMDD-NNNN-SZX). A single global
-- sequence keeps ids unique without per-day coordination.
CREATE SEQUENCE IF NOT EXISTS rfq_id_seq;
CREATE SEQUENCE IF NOT EXISTS quote_id_seq;
