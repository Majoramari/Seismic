ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS editor TEXT DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS os TEXT DEFAULT 'unknown';

CREATE INDEX IF NOT EXISTS idx_heartbeats_user_time ON heartbeats (user_id, time);

ALTER TABLE sessions
    ALTER COLUMN editor SET DEFAULT 'unknown',
    ALTER COLUMN os SET DEFAULT 'unknown';

UPDATE sessions SET editor = 'unknown' WHERE editor IS NULL OR editor = '';
UPDATE sessions SET os = 'unknown' WHERE os IS NULL OR os = '';

ALTER TABLE sessions
    ALTER COLUMN editor SET NOT NULL,
    ALTER COLUMN os SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_sessions_user_start_time ON sessions (user_id, start_time);
CREATE INDEX IF NOT EXISTS idx_sessions_user_end_time ON sessions (user_id, end_time);
