-- +goose Up
-- +goose StatementBegin
ALTER TABLE registrations DROP CONSTRAINT chk_registrations_identity;
ALTER TABLE registrations ADD CONSTRAINT chk_registrations_identity 
CHECK ((user_id IS NOT NULL) OR (guest_email IS NOT NULL));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE registrations DROP CONSTRAINT chk_registrations_identity;
ALTER TABLE registrations ADD CONSTRAINT chk_registrations_identity 
CHECK ((((user_id IS NOT NULL) AND (guest_email IS NULL)) OR ((user_id IS NULL) AND (guest_email IS NOT NULL))));
-- +goose StatementEnd