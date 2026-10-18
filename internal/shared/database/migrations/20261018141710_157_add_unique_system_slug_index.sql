-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX idx_system_templates_unique_slug
  ON certificate_templates (system_slug)
  WHERE is_system = true AND system_slug != '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_system_templates_unique_slug;
-- +goose StatementEnd