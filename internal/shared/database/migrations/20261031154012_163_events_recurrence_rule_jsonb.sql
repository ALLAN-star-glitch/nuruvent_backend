-- +goose Up
-- +goose StatementBegin

-- 1. Add the new JSONB column to store the RecurrenceRule value object
ALTER TABLE events ADD COLUMN recurrence_rule JSONB;

-- 2. Backfill existing recurring events (best-effort migration from old flat fields)
UPDATE events
SET recurrence_rule = jsonb_build_object(
  'frequency',
    CASE
      WHEN recurrence_pattern_id IS NOT NULL THEN 'weekly'
      ELSE 'daily'
    END,
  'interval',    COALESCE(recurrence_interval, 1),
  'days_of_week', COALESCE(recurrence_days_of_week, '[]'::jsonb),
  'end_date',    recurrence_ends_on,
  'count',       recurrence_occurrences
)
WHERE is_recurring = true;

-- 3. Drop the old columns (safe now that data is moved)
ALTER TABLE events DROP COLUMN IF EXISTS recurrence_pattern_id;
ALTER TABLE events DROP COLUMN IF EXISTS recurrence_interval;
ALTER TABLE events DROP COLUMN IF EXISTS recurrence_ends_on;
ALTER TABLE events DROP COLUMN IF EXISTS recurrence_occurrences;
ALTER TABLE events DROP COLUMN IF EXISTS recurrence_days_of_week;
ALTER TABLE events DROP COLUMN IF EXISTS recurrence_day_of_month;
ALTER TABLE events DROP COLUMN IF EXISTS recurrence_week_of_month;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

-- 1. Restore the old columns
ALTER TABLE events ADD COLUMN recurrence_pattern_id UUID;
ALTER TABLE events ADD COLUMN recurrence_interval INT;
ALTER TABLE events ADD COLUMN recurrence_ends_on TIMESTAMPTZ;
ALTER TABLE events ADD COLUMN recurrence_occurrences INT;
ALTER TABLE events ADD COLUMN recurrence_days_of_week JSONB;
ALTER TABLE events ADD COLUMN recurrence_day_of_month INT;
ALTER TABLE events ADD COLUMN recurrence_week_of_month VARCHAR(20);

-- 2. Backfill old columns from the JSONB column (best-effort reverse)
UPDATE events
SET 
  recurrence_interval = COALESCE((recurrence_rule->>'interval')::int, 1),
  recurrence_ends_on = (recurrence_rule->>'end_date')::timestamptz,
  recurrence_occurrences = (recurrence_rule->>'count')::int,
  recurrence_days_of_week = CASE 
    WHEN recurrence_rule->>'days_of_week' IS NOT NULL THEN recurrence_rule->'days_of_week'
    ELSE '[]'::jsonb
  END
WHERE recurrence_rule IS NOT NULL;

-- 3. Drop the new column
ALTER TABLE events DROP COLUMN IF EXISTS recurrence_rule;

-- +goose StatementEnd