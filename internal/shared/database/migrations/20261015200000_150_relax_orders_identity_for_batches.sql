-- +goose Up

-- ============================================================
-- Relax chk_orders_identity for certificate_batch orders.
--
-- The original constraint requires exactly one of user_id or
-- guest_email. That's correct for registration orders (an attendee
-- buys a ticket) and for legacy attendee-pays certificate orders.
--
-- Certificate batch orders are different: they're account-level
-- purchases made by an admin on behalf of the account. Neither a
-- user_id nor a guest_email applies — the order belongs to the
-- account via billed_account_id.
-- ============================================================

ALTER TABLE orders DROP CONSTRAINT chk_orders_identity;

ALTER TABLE orders
  ADD CONSTRAINT chk_orders_identity
  CHECK (
    -- Registration: exactly one of user_id / guest_email
    (kind = 'registration' AND (
      (user_id IS NOT NULL AND guest_email IS NULL) OR
      (user_id IS NULL AND guest_email IS NOT NULL)
    ))
    OR
    -- Legacy attendee-pays certificate: same rule
    (kind = 'certificate' AND (
      (user_id IS NOT NULL AND guest_email IS NULL) OR
      (user_id IS NULL AND guest_email IS NOT NULL)
    ))
    OR
    -- Certificate batch: no identity required
    (kind = 'certificate_batch')
  );


-- +goose Down

ALTER TABLE orders DROP CONSTRAINT chk_orders_identity;

ALTER TABLE orders
  ADD CONSTRAINT chk_orders_identity
  CHECK (
    (user_id IS NOT NULL AND guest_email IS NULL) OR
    (user_id IS NULL AND guest_email IS NOT NULL)
  );