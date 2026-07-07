-- A richer achievement system linking users to specific badges and their metadata
CREATE TABLE achievement_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(50) UNIQUE NOT NULL,
    title VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    badge_class VARCHAR(50) NOT NULL
);

CREATE TABLE user_achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    achievement_type_id UUID REFERENCES achievement_types(id) ON DELETE CASCADE,
    earned_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (user_id, achievement_type_id)
);

CREATE INDEX idx_user_achievements_user_id ON user_achievements(user_id);
