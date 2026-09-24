-- +goose Up
ALTER TABLE payments ADD COLUMN redirect_url TEXT;

-- +goose Down
ALTER TABLE payments DROP COLUMN redirect_url;