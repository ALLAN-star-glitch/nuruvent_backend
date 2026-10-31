-- +goose Up
-- +goose StatementBegin
ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_identity;
ALTER TABLE orders ADD CONSTRAINT chk_orders_identity
CHECK (
    (
        (kind)::text = 'registration'::text
        AND (user_id IS NOT NULL OR guest_email IS NOT NULL)
    )
    OR (
        (kind)::text = 'certificate'::text
        AND (user_id IS NOT NULL OR guest_email IS NOT NULL)
    )
    OR (kind)::text = 'certificate_batch'::text
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_identity;
ALTER TABLE orders ADD CONSTRAINT chk_orders_identity
CHECK (
    (
        (kind)::text = 'registration'::text
        AND (
            (user_id IS NOT NULL AND guest_email IS NULL)
            OR (user_id IS NULL AND guest_email IS NOT NULL)
        )
    )
    OR (
        (kind)::text = 'certificate'::text
        AND (
            (user_id IS NOT NULL AND guest_email IS NULL)
            OR (user_id IS NULL AND guest_email IS NOT NULL)
        )
    )
    OR (kind)::text = 'certificate_batch'::text
);
-- +goose StatementEnd