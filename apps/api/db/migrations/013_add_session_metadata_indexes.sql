ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS editor TEXT,
    ADD COLUMN IF NOT EXISTS os TEXT;

CREATE INDEX IF NOT EXISTS idx_heartbeats_user_time ON heartbeats (user_id, time);

UPDATE sessions s
SET editor = COALESCE((
        SELECT h.editor
        FROM heartbeats h
        WHERE h.user_id = s.user_id
            AND to_timestamp(h.time / 1000.0) >= s.start_time - INTERVAL '1 second'
            AND to_timestamp(h.time / 1000.0) < s.end_time
            AND h.editor IS NOT NULL
            AND h.editor <> ''
        GROUP BY h.editor
        ORDER BY COUNT(*) DESC
        LIMIT 1
    ), 'unknown'),
    os = COALESCE((
        SELECT h.os
        FROM heartbeats h
        WHERE h.user_id = s.user_id
            AND to_timestamp(h.time / 1000.0) >= s.start_time - INTERVAL '1 second'
            AND to_timestamp(h.time / 1000.0) < s.end_time
            AND h.os IS NOT NULL
            AND h.os <> ''
        GROUP BY h.os
        ORDER BY COUNT(*) DESC
        LIMIT 1
    ), 'unknown');

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
