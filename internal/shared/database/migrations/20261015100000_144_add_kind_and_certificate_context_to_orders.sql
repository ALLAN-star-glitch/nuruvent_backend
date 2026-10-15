-- +goose Up

-- ============================================================
-- Extend `orders` to support certificate purchases
--
-- Before this migration, every order was tied to a registration.
-- Certificates introduce a second order kind: a single-amount
-- purchase by a recipient, with no registration and no items.
--
-- Changes:
--   1. registration_id becomes nullable
--   2. kind discriminates between "registration" and "certificate"
--   3. certificate_context carries the certificate-specific input
-- ============================================================

-- 1. Make registration_id nullable.
ALTER TABLE orders
  ALTER COLUMN registration_id DROP NOT NULL;

-- 2. Add kind. Existing rows are all registration orders.
ALTER TABLE orders
  ADD COLUMN kind VARCHAR(20) NOT NULL DEFAULT 'registration';

ALTER TABLE orders
  ADD CONSTRAINT chk_orders_kind
  CHECK (kind IN ('registration', 'certificate'));

-- 3. Add certificate_context (nullable — only set for certificate orders).
ALTER TABLE orders
  ADD COLUMN certificate_context JSONB NULL;

-- 4. Consistency: registration orders must have a registration_id.
--    Certificate orders must have a certificate_context.
ALTER TABLE orders
  ADD CONSTRAINT chk_orders_kind_context
  CHECK (
    (kind = 'registration' AND registration_id IS NOT NULL AND certificate_context IS NULL) OR
    (kind = 'certificate'  AND registration_id IS NULL     AND certificate_context IS NOT NULL)
  );

-- 5. Idempotency for certificate orders: one pending order per
--    (context_type, context_id, recipient_email).
CREATE UNIQUE INDEX uniq_orders_pending_certificate
  ON orders (
    (certificate_context->>'context_type'),
    (certificate_context->>'context_id'),
    (certificate_context->>'recipient_email')
  )
  WHERE status = 'pending' AND kind = 'certificate';


-- +goose Down

DROP INDEX IF EXISTS uniq_orders_pending_certificate;

ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_kind_context;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_kind;
ALTER TABLE orders DROP COLUMN IF EXISTS certificate_context;
ALTER TABLE orders DROP COLUMN IF EXISTS kind;

ALTER TABLE orders
  ALTER COLUMN registration_id SET NOT NULL;