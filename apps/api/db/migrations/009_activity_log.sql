-- Activity log for the recent activity feed on a user's profile
CREATE TABLE activity_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    kind VARCHAR(50) NOT NULL,
    text TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_activity_log_user_time ON activity_log(user_id, created_at DESC);

-- Also add index on sessions and heartbeats as requested in the plan
CREATE INDEX idx_sessions_user_time ON sessions(user_id, start_time);
CREATE INDEX idx_heartbeats_user_time ON heartbeats(user_id, time);
