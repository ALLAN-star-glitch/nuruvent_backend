-- +goose Up

-- ============================================================
-- Account-level default certificate price
--
-- In minor units. Default 6500 = KES 65.00.
-- Organizers can change this later via the account settings UI.
-- ============================================================

ALTER TABLE accounts
  ADD COLUMN certificate_price_minor INT NOT NULL DEFAULT 6500;


-- +goose Down

ALTER TABLE accounts
  DROP COLUMN IF EXISTS certificate_price_minor;