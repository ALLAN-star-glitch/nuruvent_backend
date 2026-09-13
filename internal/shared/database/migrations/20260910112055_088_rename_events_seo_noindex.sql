-- +goose Up
-- +goose StatementBegin

-- Rename seo_noindex -> seo_no_index for consistency with other SEO columns
-- (seo_canonical_url, seo_description, etc.) and GORM's snake_case convention.

ALTER TABLE events RENAME COLUMN seo_noindex TO seo_no_index;

-- Rename the index for consistency
ALTER INDEX IF EXISTS idx_events_seo_noindex RENAME TO idx_events_seo_no_index;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Rollback: rename the index first, then the column
ALTER INDEX IF EXISTS idx_events_seo_no_index RENAME TO idx_events_seo_noindex;

ALTER TABLE events RENAME COLUMN seo_no_index TO seo_noindex;

-- +goose StatementEnd