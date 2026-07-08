-- Ensure we have a user to attach these to
DO $$ 
DECLARE
    test_user_id UUID;
    ach_id UUID;
BEGIN
    -- Get or create a test user
    SELECT id INTO test_user_id FROM users WHERE username = 'testuser' LIMIT 1;
    
    IF NOT FOUND THEN
        INSERT INTO users (username, email, display_name, first_name, role, location, university, bio, time_zone, gender, languages)
        VALUES ('testuser', 'testuser@example.com', 'Test User', 'Test', 'Developer', 'San Francisco, CA', 'Tech University', 'Just a test user building cool things.', 'America/Los_Angeles', 'Male', ARRAY['Go', 'TypeScript', 'SQL'])
        RETURNING id INTO test_user_id;
    END IF;

    -- Upsert Profile Stats
    INSERT INTO user_profile_stats (user_id, total_coding_seconds, total_active_days, current_streak, max_streak)
    VALUES (test_user_id, 360000, 45, 12, 15)
    ON CONFLICT (user_id) DO UPDATE SET
        total_coding_seconds = EXCLUDED.total_coding_seconds,
        total_active_days = EXCLUDED.total_active_days,
        current_streak = EXCLUDED.current_streak,
        max_streak = EXCLUDED.max_streak;

    -- Upsert Problem Stats (Placeholder)
    INSERT INTO user_problem_stats (user_id, solved_count, total_problems, attempting_count, rating, contribution_points)
    VALUES (test_user_id, 150, 500, 3, 1450, 200)
    ON CONFLICT (user_id) DO UPDATE SET
        solved_count = EXCLUDED.solved_count,
        total_problems = EXCLUDED.total_problems,
        attempting_count = EXCLUDED.attempting_count,
        rating = EXCLUDED.rating,
        contribution_points = EXCLUDED.contribution_points;

    -- Seed Achievement Types
    INSERT INTO achievement_types (key, title, description, badge_class)
    VALUES 
        ('first_blood', 'First Blood', 'Logged your first session.', 'gold'),
        ('night_owl', 'Night Owl', 'Coded past 2 AM.', 'silver'),
        ('streak_7', '7-Day Streak', 'Maintained a 7 day streak.', 'bronze')
    ON CONFLICT (key) DO NOTHING;

    -- Give the user an achievement
    SELECT id INTO ach_id FROM achievement_types WHERE key = 'first_blood' LIMIT 1;
    IF ach_id IS NOT NULL THEN
        INSERT INTO user_achievements (user_id, achievement_type_id)
        VALUES (test_user_id, ach_id)
        ON CONFLICT (user_id, achievement_type_id) DO NOTHING;
    END IF;

    -- Seed Activity Log
    INSERT INTO activity_log (user_id, kind, text)
    VALUES (test_user_id, 'achievement', 'Earned the First Blood badge!'),
           (test_user_id, 'streak', 'Reached a 12 day streak!');

END $$;
