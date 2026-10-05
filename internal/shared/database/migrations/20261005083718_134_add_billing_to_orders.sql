-- +goose Up
-- +goose StatementBegin

ALTER TABLE orders
    ADD COLUMN billed_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    ADD COLUMN commission_rate   NUMERIC(5,4) NOT NULL DEFAULT 0.1000,
    ADD COLUMN settled_at        TIMESTAMP WITH TIME ZONE,
    ADD COLUMN payout_ref        VARCHAR(255);

-- Backfill through the ownership chain:
--   orders → registrations → event_registrations → events → teams → accounts
UPDATE orders o
SET billed_account_id = t.account_id
FROM registrations r
JOIN event_registrations er ON er.registration_id = r.id
JOIN events e               ON e.id = er.event_id
JOIN teams  t               ON t.id = e.team_id
WHERE r.id = o.registration_id
  AND t.deleted_at IS NULL
  AND o.billed_account_id IS NULL;

-- Fail loudly if any order couldn't be resolved.
ALTER TABLE orders
    ALTER COLUMN billed_account_id SET NOT NULL;

-- Direct lookup by account (used by the ledger and payout queries).
CREATE INDEX idx_orders_billed_account_id
    ON orders (billed_account_id)
    WHERE deleted_at IS NULL;

-- Paid orders for one account, ordered by paid_at — the hot path for
-- the payouts report and per-account revenue queries.
CREATE INDEX idx_orders_billed_account_status_paid_at
    ON orders (billed_account_id, status, paid_at)
    WHERE deleted_at IS NULL AND status = 'paid';

-- Unsettled paid orders — the manual payout workflow's primary query.
CREATE INDEX idx_orders_unsettled
    ON orders (billed_account_id)
    WHERE deleted_at IS NULL AND status = 'paid' AND settled_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_orders_unsettled;
DROP INDEX IF EXISTS idx_orders_billed_account_status_paid_at;
DROP INDEX IF EXISTS idx_orders_billed_account_id;

ALTER TABLE orders
    DROP COLUMN IF EXISTS payout_ref,
    DROP COLUMN IF EXISTS settled_at,
    DROP COLUMN IF EXISTS commission_rate,
    DROP COLUMN IF EXISTS billed_account_id;

-- +goose StatementEnd