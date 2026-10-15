-- +goose Up

-- ============================================================
-- certificates
--
-- A certificate snapshots everything needed to render and verify
-- an issued certificate. It never consults the original context
-- (event/course) at render time.
-- ============================================================

CREATE TABLE certificates (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id           UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
  template_id          UUID NOT NULL REFERENCES certificate_templates(id) ON DELETE RESTRICT,
  verification_code    VARCHAR(64) NOT NULL UNIQUE,

  -- Recipient snapshot (user_id is nullable for guest recipients)
  recipient_user_id    UUID NULL REFERENCES users(id) ON DELETE SET NULL,
  recipient_name       VARCHAR(255) NOT NULL,
  recipient_email      VARCHAR(255) NOT NULL,

  -- Opaque context reference: "event", "course", "workshop", ...
  context_type         VARCHAR(50) NOT NULL,
  context_id           UUID NOT NULL,

  -- Free-form metadata snapshot (mirrors certificate_templates.design style)
  metadata             JSONB NOT NULL DEFAULT '{}'::jsonb,

  -- Completion vs issuance are distinct
  completed_at         TIMESTAMPTZ NOT NULL,
  issued_at            TIMESTAMPTZ NOT NULL,

  -- Payment (nullable — empty for free certificates)
  payment_reference    VARCHAR(255) NULL,

  -- Lifecycle
  status               VARCHAR(20) NOT NULL DEFAULT 'active',
  revoked_at           TIMESTAMPTZ NULL,
  revocation_reason    TEXT NULL,

  -- Cached PDF
  pdf_url              VARCHAR(500) NULL,
  pdf_generated_at     TIMESTAMPTZ NULL,

  created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT chk_certificates_status
    CHECK (status IN ('active', 'revoked')),

  CONSTRAINT chk_certificates_revoked_consistency
    CHECK (
      (status = 'active'  AND revoked_at IS NULL) OR
      (status = 'revoked' AND revoked_at IS NOT NULL)
    )
);

-- Idempotency: one certificate per (context, recipient).
-- Registered recipients use recipient_user_id; guests use recipient_email.
CREATE UNIQUE INDEX uniq_certificates_context_user
  ON certificates (context_type, context_id, recipient_user_id)
  WHERE recipient_user_id IS NOT NULL;

CREATE UNIQUE INDEX uniq_certificates_context_email
  ON certificates (context_type, context_id, recipient_email)
  WHERE recipient_user_id IS NULL;

-- Public verification lookup
CREATE UNIQUE INDEX uniq_certificates_verification_code
  ON certificates (verification_code);

-- Account listing
CREATE INDEX idx_certificates_account_id
  ON certificates (account_id);

CREATE INDEX idx_certificates_account_context
  ON certificates (account_id, context_type, context_id);

-- Recipient listing
CREATE INDEX idx_certificates_recipient_user_id
  ON certificates (recipient_user_id)
  WHERE recipient_user_id IS NOT NULL;

CREATE INDEX idx_certificates_recipient_email
  ON certificates (recipient_email)
  WHERE recipient_user_id IS NULL;

-- Context filtering (used by account view when context ID is known)
CREATE INDEX idx_certificates_context
  ON certificates (context_type, context_id);


-- +goose Down

DROP INDEX IF EXISTS idx_certificates_context;
DROP INDEX IF EXISTS idx_certificates_recipient_email;
DROP INDEX IF EXISTS idx_certificates_recipient_user_id;
DROP INDEX IF EXISTS idx_certificates_account_context;
DROP INDEX IF EXISTS idx_certificates_account_id;
DROP INDEX IF EXISTS uniq_certificates_verification_code;
DROP INDEX IF EXISTS uniq_certificates_context_email;
DROP INDEX IF EXISTS uniq_certificates_context_user;
DROP TABLE IF EXISTS certificates;