-- +goose Up
-- Step 1: drop text[] defaults so the columns can change type
ALTER TABLE events
    ALTER COLUMN tags                   DROP DEFAULT,
    ALTER COLUMN recurrence_days_of_week DROP DEFAULT,
    ALTER COLUMN invited_emails         DROP DEFAULT,
    ALTER COLUMN approval_required_for  DROP DEFAULT,
    ALTER COLUMN seo_keywords           DROP DEFAULT;

-- Step 2: change type text[] → jsonb
ALTER TABLE events
    ALTER COLUMN tags                   TYPE jsonb USING to_jsonb(tags),
    ALTER COLUMN recurrence_days_of_week TYPE jsonb USING to_jsonb(recurrence_days_of_week),
    ALTER COLUMN invited_emails         TYPE jsonb USING to_jsonb(invited_emails),
    ALTER COLUMN approval_required_for  TYPE jsonb USING to_jsonb(approval_required_for),
    ALTER COLUMN seo_keywords           TYPE jsonb USING to_jsonb(seo_keywords),
    ALTER COLUMN venue_coordinates      TYPE jsonb USING to_jsonb(venue_coordinates),
    ALTER COLUMN social_links           TYPE jsonb USING to_jsonb(social_links),
    ALTER COLUMN metadata               TYPE jsonb USING to_jsonb(metadata),
    ALTER COLUMN schema_org             TYPE jsonb USING to_jsonb(schema_org);

ALTER TABLE event_speakers
    ALTER COLUMN social_links TYPE jsonb USING to_jsonb(social_links);

-- Step 3: re-add defaults as jsonb
ALTER TABLE events
    ALTER COLUMN tags                   SET DEFAULT '[]'::jsonb,
    ALTER COLUMN recurrence_days_of_week SET DEFAULT '[]'::jsonb,
    ALTER COLUMN invited_emails         SET DEFAULT '[]'::jsonb,
    ALTER COLUMN approval_required_for  SET DEFAULT '[]'::jsonb,
    ALTER COLUMN seo_keywords           SET DEFAULT '[]'::jsonb,
    ALTER COLUMN venue_coordinates      SET DEFAULT '{}'::jsonb,
    ALTER COLUMN social_links           SET DEFAULT '{}'::jsonb,
    ALTER COLUMN metadata               SET DEFAULT '{}'::jsonb;
    -- schema_org had no default, leave it

ALTER TABLE event_speakers
    ALTER COLUMN social_links SET DEFAULT '{}'::jsonb;

-- +goose Down
ALTER TABLE events
    ALTER COLUMN tags                   DROP DEFAULT,
    ALTER COLUMN recurrence_days_of_week DROP DEFAULT,
    ALTER COLUMN invited_emails         DROP DEFAULT,
    ALTER COLUMN approval_required_for  DROP DEFAULT,
    ALTER COLUMN seo_keywords           DROP DEFAULT,
    ALTER COLUMN venue_coordinates      DROP DEFAULT,
    ALTER COLUMN social_links           DROP DEFAULT,
    ALTER COLUMN metadata               DROP DEFAULT;

ALTER TABLE events
    ALTER COLUMN tags                   TYPE text[] USING ARRAY(SELECT jsonb_array_elements_text(COALESCE(tags, '[]'::jsonb))),
    ALTER COLUMN recurrence_days_of_week TYPE text[] USING ARRAY(SELECT jsonb_array_elements_text(COALESCE(recurrence_days_of_week, '[]'::jsonb))),
    ALTER COLUMN invited_emails         TYPE text[] USING ARRAY(SELECT jsonb_array_elements_text(COALESCE(invited_emails, '[]'::jsonb))),
    ALTER COLUMN approval_required_for  TYPE text[] USING ARRAY(SELECT jsonb_array_elements_text(COALESCE(approval_required_for, '[]'::jsonb))),
    ALTER COLUMN seo_keywords           TYPE text[] USING ARRAY(SELECT jsonb_array_elements_text(COALESCE(seo_keywords, '[]'::jsonb))),
    ALTER COLUMN venue_coordinates      TYPE text[] USING ARRAY[venue_coordinates::text],
    ALTER COLUMN social_links           TYPE text[] USING ARRAY[social_links::text],
    ALTER COLUMN metadata               TYPE text[] USING ARRAY[metadata::text],
    ALTER COLUMN schema_org             TYPE text[] USING ARRAY[schema_org::text];

ALTER TABLE event_speakers
    ALTER COLUMN social_links TYPE text[] USING ARRAY[social_links::text];