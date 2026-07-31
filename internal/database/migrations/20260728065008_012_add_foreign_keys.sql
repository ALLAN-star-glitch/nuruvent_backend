-- +goose Up

-- Users to Businesses
ALTER TABLE users ADD CONSTRAINT fk_users_business_id 
    FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE SET NULL;

-- Businesses to Business Types
ALTER TABLE businesses ADD CONSTRAINT fk_businesses_business_type_id 
    FOREIGN KEY (business_type_id) REFERENCES business_types(id) ON DELETE SET NULL;

-- Businesses to Users (created_by)
ALTER TABLE businesses ADD CONSTRAINT fk_businesses_created_by 
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;

-- Businesses to Users (verified_by)
ALTER TABLE businesses ADD CONSTRAINT fk_businesses_verified_by 
    FOREIGN KEY (verified_by) REFERENCES users(id) ON DELETE SET NULL;

-- Business Members
ALTER TABLE business_members ADD CONSTRAINT fk_business_members_business_id 
    FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE;

ALTER TABLE business_members ADD CONSTRAINT fk_business_members_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE business_members ADD CONSTRAINT uk_business_members_business_user 
    UNIQUE (business_id, user_id);

-- Events
ALTER TABLE events ADD CONSTRAINT fk_events_business_id 
    FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE;

ALTER TABLE events ADD CONSTRAINT fk_events_created_by 
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;

-- Attendees
ALTER TABLE attendees ADD CONSTRAINT fk_attendees_event_id 
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE;

ALTER TABLE attendees ADD CONSTRAINT fk_attendees_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- Certificates
ALTER TABLE certificates ADD CONSTRAINT fk_certificates_event_id 
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE;

ALTER TABLE certificates ADD CONSTRAINT fk_certificates_attendee_id 
    FOREIGN KEY (attendee_id) REFERENCES attendees(id) ON DELETE CASCADE;

ALTER TABLE certificates ADD CONSTRAINT fk_certificates_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- Payments
ALTER TABLE payments ADD CONSTRAINT fk_payments_event_id 
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE;

ALTER TABLE payments ADD CONSTRAINT fk_payments_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Payouts
ALTER TABLE payouts ADD CONSTRAINT fk_payouts_business_id 
    FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE;

-- Refresh Tokens
ALTER TABLE refresh_tokens ADD CONSTRAINT fk_refresh_tokens_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down

-- Drop all foreign keys (in reverse order)
ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS fk_refresh_tokens_user_id;
ALTER TABLE payouts DROP CONSTRAINT IF EXISTS fk_payouts_business_id;
ALTER TABLE payments DROP CONSTRAINT IF EXISTS fk_payments_user_id;
ALTER TABLE payments DROP CONSTRAINT IF EXISTS fk_payments_event_id;
ALTER TABLE certificates DROP CONSTRAINT IF EXISTS fk_certificates_user_id;
ALTER TABLE certificates DROP CONSTRAINT IF EXISTS fk_certificates_attendee_id;
ALTER TABLE certificates DROP CONSTRAINT IF EXISTS fk_certificates_event_id;
ALTER TABLE attendees DROP CONSTRAINT IF EXISTS fk_attendees_user_id;
ALTER TABLE attendees DROP CONSTRAINT IF EXISTS fk_attendees_event_id;
ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_created_by;
ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_business_id;
ALTER TABLE business_members DROP CONSTRAINT IF EXISTS uk_business_members_business_user;
ALTER TABLE business_members DROP CONSTRAINT IF EXISTS fk_business_members_user_id;
ALTER TABLE business_members DROP CONSTRAINT IF EXISTS fk_business_members_business_id;
ALTER TABLE businesses DROP CONSTRAINT IF EXISTS fk_businesses_verified_by;
ALTER TABLE businesses DROP CONSTRAINT IF EXISTS fk_businesses_created_by;
ALTER TABLE businesses DROP CONSTRAINT IF EXISTS fk_businesses_business_type_id;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_business_id;