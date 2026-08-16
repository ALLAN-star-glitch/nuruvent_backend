-- +goose Up
-- +goose StatementBegin

-- Rename business_id to account_id
ALTER TABLE events RENAME COLUMN business_id TO account_id;

-- Update index name
DROP INDEX IF EXISTS idx_events_business_id;
CREATE INDEX idx_events_account_id ON events(account_id);

-- Update foreign key constraint if it exists
ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_business;
ALTER TABLE events 
    ADD CONSTRAINT fk_events_account 
    FOREIGN KEY (account_id) 
    REFERENCES accounts(id) 
    ON DELETE CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Rollback
ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_account;
DROP INDEX IF EXISTS idx_events_account_id;

ALTER TABLE events RENAME COLUMN account_id TO business_id;

CREATE INDEX idx_events_business_id ON events(business_id);

ALTER TABLE events 
    ADD CONSTRAINT fk_events_business 
    FOREIGN KEY (business_id) 
    REFERENCES businesses(id) 
    ON DELETE CASCADE;

-- +goose StatementEnd