-- +goose Up

-- Backfill: for each event, renumber sessions 1..N ordered by start_date,
-- tie-broken by start_time then created_at. This resolves any existing
-- duplicates (e.g. two rows with session_number = 5).
WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY event_id
               ORDER BY start_date ASC NULLS LAST,
                        start_time ASC NULLS LAST,
                        created_at ASC
           ) AS rn
    FROM event_schedules
    WHERE deleted_at IS NULL
)
UPDATE event_schedules es
SET session_number = r.rn
FROM ranked r
WHERE es.id = r.id;

-- Enforce uniqueness going forward.
CREATE UNIQUE INDEX IF NOT EXISTS uniq_event_schedules_event_session_number
    ON event_schedules (event_id, session_number)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS uniq_event_schedules_event_session_number;