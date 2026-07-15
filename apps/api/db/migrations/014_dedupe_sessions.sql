WITH ranked_sessions AS (
    SELECT
        id,
        row_number() OVER (
            PARTITION BY user_id, project, language, editor, os, start_time, end_time, duration_seconds
            ORDER BY id::text
        ) AS duplicate_rank
    FROM sessions
)
DELETE FROM sessions
WHERE id IN (
    SELECT id
    FROM ranked_sessions
    WHERE duplicate_rank > 1
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_no_exact_duplicates
ON sessions (user_id, project, language, editor, os, start_time, end_time, duration_seconds);
