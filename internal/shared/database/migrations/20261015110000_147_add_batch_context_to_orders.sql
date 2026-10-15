-- +goose Up

-- ============================================================
-- Certificate batch orders
--
-- The attendee-pays model is replaced by an organizer-pays batch
-- model. A certificate_batch order carries:
--   - batch_context       — the batch metadata + recipient snapshot
--   - batch_attendee_count — number of certificates the batch covers
--
-- The old certificate_context column stays for backward compatibility
-- with any existing (now dormant) certificate orders.
-- ============================================================

ALTER TABLE orders
  ADD COLUMN batch_context        JSONB NULL,
  ADD COLUMN batch_attendee_count INT   NOT NULL DEFAULT 0;

-- Enforce consistency for batch orders.
ALTER TABLE orders
  ADD CONSTRAINT chk_orders_batch_context
  CHECK (
    kind <> 'certificate_batch'
    OR (
      batch_context IS NOT NULL
      AND batch_attendee_count > 0
    )
  );

-- Extend the kind check to allow the new value.
ALTER TABLE orders
  DROP CONSTRAINT IF EXISTS chk_orders_kind;

ALTER TABLE orders
  ADD CONSTRAINT chk_orders_kind
  CHECK (kind IN ('registration', 'certificate', 'certificate_batch'));

-- Update the context consistency check to cover the new kind.
ALTER TABLE orders
  DROP CONSTRAINT IF EXISTS chk_orders_kind_context;

ALTER TABLE orders
  ADD CONSTRAINT chk_orders_kind_context
  CHECK (
    (kind = 'registration'      AND registration_id IS NOT NULL) OR
    (kind = 'certificate'       AND certificate_context IS NOT NULL) OR
    (kind = 'certificate_batch' AND batch_context IS NOT NULL)
  );

-- One pending batch order per event (prevents double-charging on retry).
CREATE UNIQUE INDEX uniq_orders_pending_batch
  ON orders ((batch_context->>'event_id'))
  WHERE status = 'pending' AND kind = 'certificate_batch';


-- +goose Down

DROP INDEX IF EXISTS uniq_orders_pending_batch;

ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_kind_context;

ALTER TABLE orders
  ADD CONSTRAINT chk_orders_kind_context
  CHECK (
    (kind = 'registration' AND registration_id IS NOT NULL AND certificate_context IS NULL) OR
    (kind = 'certificate'  AND registration_id IS NULL     AND certificate_context IS NOT NULL)
  );

ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_kind;
ALTER TABLE orders
  ADD CONSTRAINT chk_orders_kind
  CHECK (kind IN ('registration', 'certificate'));

ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_batch_context;
ALTER TABLE orders
  DROP COLUMN IF EXISTS batch_attendee_count,
  DROP COLUMN IF EXISTS batch_context;