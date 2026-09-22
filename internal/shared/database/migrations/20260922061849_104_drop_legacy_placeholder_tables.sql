-- +goose Up
-- +goose StatementBegin
-- Legacy placeholder tables from the initial schema. Their owning
-- modules have real designs now; these stand-ins are not used by any
-- code and should not exist.
DROP TABLE IF EXISTS payouts CASCADE;
DROP TABLE IF EXISTS certificates CASCADE;
-- Do NOT drop payments — it's already being dropped by 101.
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 1;
-- +goose StatementEnd