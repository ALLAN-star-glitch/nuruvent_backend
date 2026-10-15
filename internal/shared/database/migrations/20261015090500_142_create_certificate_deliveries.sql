-- +goose Up

-- ============================================================
-- certificate_deliveries
--
-- Audit trail for every attempt to deliver a certificate —
-- email sends, downloads, retries. One row per attempt.
-- ============================================================

CREATE TABLE certificate_deliveries (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  certificate_id  UUID NOT NULL REFERENCES certificates(id) ON DELETE CASCADE,
  channel         VARCHAR(20) NOT NULL,
  recipient       VARCHAR(255) NOT NULL,
  status          VARCHAR(20) NOT NULL DEFAULT 'pending',
  error           TEXT NULL,
  attempts        INT NOT NULL DEFAULT 0,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT chk_certificate_deliveries_channel
    CHECK (channel IN ('email', 'download')),

  CONSTRAINT chk_certificate_deliveries_status
    CHECK (status IN ('pending', 'success', 'failed'))
);

CREATE INDEX idx_certificate_deliveries_certificate_id
  ON certificate_deliveries (certificate_id);

CREATE INDEX idx_certificate_deliveries_status
  ON certificate_deliveries (status)
  WHERE status IN ('pending', 'failed');


-- +goose Down

DROP INDEX IF EXISTS idx_certificate_deliveries_status;
DROP INDEX IF EXISTS idx_certificate_deliveries_certificate_id;
DROP TABLE IF EXISTS certificate_deliveries;