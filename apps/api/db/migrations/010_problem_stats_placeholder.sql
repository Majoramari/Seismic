-- Placeholder table for problem-solving stats
-- WARNING: There is currently no problem-solving or contest domain in Seismic. 
-- This schema is intentionally introduced only to satisfy the future Profile API 
-- and remains unused until such a feature exists. Ratings and contributions are 
-- strictly placeholders and not derived from any existing system.
CREATE TABLE user_problem_stats (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    solved_count INTEGER DEFAULT 0,
    total_problems INTEGER DEFAULT 0,
    attempting_count INTEGER DEFAULT 0,
    rating INTEGER DEFAULT 0,
    contribution_points INTEGER DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT now()
);
