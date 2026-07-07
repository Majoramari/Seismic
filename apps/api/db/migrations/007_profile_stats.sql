-- Cached aggregate coding data to avoid heavy calculations on every profile view
-- This will be populated and updated by the background session-processing job
CREATE TABLE user_profile_stats (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    total_coding_seconds BIGINT DEFAULT 0,
    total_active_days INTEGER DEFAULT 0,
    current_streak INTEGER DEFAULT 0,
    max_streak INTEGER DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT now()
);
