-- +goose Up

-- ============================================================
-- certificate_batches
--
-- One row per batch of certificates issued for an event.
-- Links the payment order to the issuance operation so the webhook
-- can trigger issuance idempotently.
-- ============================================================

CREATE TABLE certificate_batches (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id      UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
  event_id        UUID NOT NULL,
  template_id     UUID NOT NULL REFERENCES certificate_templates(id) ON DELETE RESTRICT,
  order_id        UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,

  attendee_count  INT NOT NULL,
  total_amount_minor BIGINT NOT NULL,
  currency        VARCHAR(3) NOT NULL DEFAULT 'KES',

  status          VARCHAR(20) NOT NULL DEFAULT 'pending',
  issued_at       TIMESTAMPTZ NULL,

  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT chk_certificate_batches_status
    CHECK (status IN ('pending', 'paid', 'issued', 'failed', 'refunded'))
);

CREATE INDEX idx_certificate_batches_event
  ON certificate_batches (event_id);

CREATE INDEX idx_certificate_batches_order
  ON certificate_batches (order_id);

CREATE INDEX idx_certificate_batches_account
  ON certificate_batches (account_id);


-- +goose Down

DROP TABLE IF EXISTS certificate_batches;