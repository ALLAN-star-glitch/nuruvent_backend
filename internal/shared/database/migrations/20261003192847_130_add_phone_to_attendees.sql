-- +goose Up
ALTER TABLE attendees
  ADD COLUMN phone VARCHAR(50) NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE attendees
  DROP COLUMN phone;