-- ===============================================================
-- Seismic Project: Dummy Data Seed Script
-- Populates 4 users with full profiles and all related tables.
-- All timestamps are logically coherent.
-- ===============================================================

DO $$
DECLARE
    u1_id UUID;  -- Sarah Chen
    u2_id UUID;  -- Marcus Johnson
    u3_id UUID;  -- Elena Rodriguez
    u4_id UUID;  -- Kenji Tanaka

    a_first_blood UUID;
    a_streak_7    UUID;
    a_night_owl   UUID;
    a_century     UUID;
BEGIN

    -- =============================================================
    -- 1. USERS
    -- =============================================================

    INSERT INTO users (id, username, email, api_key, display_name, first_name, role, location, university, country, bio, website, time_zone, gender, languages, last_active_at, created_at)
    VALUES (gen_random_uuid(), 'sarahchen',   'sarah.chen@example.com',   gen_random_uuid(), 'Sarah Chen',    'Sarah',   'Researcher',     'Cambridge, MA',  'MIT',                         'US', 'Computational seismology researcher. Open-source contributor.',          'https://sarahchen.dev',     'America/New_York',   'Female', ARRAY['Python', 'Julia', 'Rust', 'MATLAB'],       '2026-07-07 14:30:00+00', '2025-01-15 08:00:00+00')
    RETURNING id INTO u1_id;

    INSERT INTO users (id, username, email, api_key, display_name, first_name, role, location, university, country, bio, website, time_zone, gender, languages, last_active_at, created_at)
    VALUES (gen_random_uuid(), 'mjohnson',    'marcus.j@example.com',     gen_random_uuid(), 'Marcus Johnson', 'Marcus',  'Student',        'Austin, TX',     'University of Texas',        'US', 'CS undergrad exploring high-performance computing and visualization.', 'https://marcusj.io',        'America/Chicago',    'Male',   ARRAY['Go', 'TypeScript', 'C++', 'Python'],         '2026-07-06 22:15:00+00', '2025-03-22 10:30:00+00')
    RETURNING id INTO u2_id;

    INSERT INTO users (id, username, email, api_key, display_name, first_name, role, location, university, country, bio, website, time_zone, gender, languages, last_active_at, created_at)
    VALUES (gen_random_uuid(), 'elena_r',     'elena.r@example.com',      gen_random_uuid(), 'Elena Rodríguez', 'Elena',   'Full-Stack Dev', 'Barcelona',      'Universitat Politècnica de Catalunya', 'ES', 'Building developer tools and data visualization platforms.',          'https://elena.dev',         'Europe/Madrid',      'Female', ARRAY['TypeScript', 'Go', 'Python', 'SQL'],       '2026-07-07 09:45:00+00', '2024-11-05 12:00:00+00')
    RETURNING id INTO u3_id;

    INSERT INTO users (id, username, email, api_key, display_name, first_name, role, location, university, country, bio, website, time_zone, gender, languages, last_active_at, created_at)
    VALUES (gen_random_uuid(), 'ktanaka',     'kenji.t@example.com',      gen_random_uuid(), 'Kenji Tanaka',  'Kenji',   'Data Scientist',  'Tokyo',          'University of Tokyo',        'JP', 'ML engineering and seismic data analysis. Rust enthusiast.',          'https://kenji.tech',        'Asia/Tokyo',         'Male',   ARRAY['Rust', 'Python', 'R', 'Go'],              '2026-07-07 16:00:00+00', '2025-06-10 09:00:00+00')
    RETURNING id INTO u4_id;

    -- =============================================================
    -- 2. USER PROFILE STATS
    -- =============================================================

    INSERT INTO user_profile_stats (user_id, total_coding_seconds, total_active_days, current_streak, max_streak, updated_at)
    VALUES
        (u1_id, 720000, 180, 14, 28, '2026-07-07 14:30:00+00'),
        (u2_id, 450000, 120,  5, 18, '2026-07-06 22:15:00+00'),
        (u3_id, 960000, 240, 31, 45, '2026-07-07 09:45:00+00'),
        (u4_id, 310000,  85,  3, 10, '2026-07-07 16:00:00+00');

    -- =============================================================
    -- 3. REFRESH TOKENS (mix: active, expired, revoked)
    -- =============================================================

    -- Sarah: 1 active, 1 expired, 1 revoked
    INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked, created_at)
    VALUES
        (gen_random_uuid(), u1_id, 'a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b', '2026-08-06 12:00:00+00', false, '2026-07-07 12:00:00+00'),
        (gen_random_uuid(), u1_id, 'b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2', '2026-06-01 12:00:00+00', false, '2026-05-02 12:00:00+00'),
        (gen_random_uuid(), u1_id, 'c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c', '2026-07-01 12:00:00+00', true,  '2026-06-01 12:00:00+00');

    -- Marcus: 2 active, 1 revoked
    INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked, created_at)
    VALUES
        (gen_random_uuid(), u2_id, 'd4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d', '2026-08-01 10:00:00+00', false, '2026-07-02 10:00:00+00'),
        (gen_random_uuid(), u2_id, 'e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4', '2026-08-05 15:00:00+00', false, '2026-07-06 15:00:00+00'),
        (gen_random_uuid(), u2_id, 'f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e', '2026-05-15 10:00:00+00', true,  '2026-04-15 10:00:00+00');

    -- Elena: 1 active, 2 expired
    INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked, created_at)
    VALUES
        (gen_random_uuid(), u3_id, 'a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f', '2026-08-07 08:00:00+00', false, '2026-07-08 08:00:00+00'),
        (gen_random_uuid(), u3_id, 'b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6', '2026-06-10 08:00:00+00', false, '2026-05-11 08:00:00+00'),
        (gen_random_uuid(), u3_id, 'c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a', '2026-04-20 08:00:00+00', false, '2026-03-21 08:00:00+00');

    -- Kenji: 1 active, 1 expired, 1 revoked
    INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked, created_at)
    VALUES
        (gen_random_uuid(), u4_id, 'd0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8', '2026-08-04 16:00:00+00', false, '2026-07-05 16:00:00+00'),
        (gen_random_uuid(), u4_id, 'e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c', '2026-05-01 16:00:00+00', false, '2026-04-01 16:00:00+00'),
        (gen_random_uuid(), u4_id, 'f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d', '2026-06-15 16:00:00+00', true,  '2026-05-16 16:00:00+00');

    -- =============================================================
    -- 4. ACHIEVEMENT TYPES
    -- =============================================================

    INSERT INTO achievement_types (id, key, title, description, badge_class)
    VALUES
        (gen_random_uuid(), 'first_blood', 'First Blood',      'Logged your first coding session.',            'gold'),
        (gen_random_uuid(), 'streak_7',    '7-Day Streak',     'Maintained a 7-day coding streak.',            'silver'),
        (gen_random_uuid(), 'night_owl',   'Night Owl',        'Coded past 2 AM — dedication knows no hour.',  'bronze'),
        (gen_random_uuid(), 'century',     '100-Day Streak',   'Coded for 100 consecutive days. Unstoppable.',  'gold')
    ON CONFLICT (key) DO NOTHING;

    -- Re-fetch IDs (in case they already existed)
    SELECT id INTO a_first_blood FROM achievement_types WHERE key = 'first_blood';
    SELECT id INTO a_streak_7  FROM achievement_types WHERE key = 'streak_7';
    SELECT id INTO a_night_owl FROM achievement_types WHERE key = 'night_owl';
    SELECT id INTO a_century   FROM achievement_types WHERE key = 'century';

    -- =============================================================
    -- 5. USER ACHIEVEMENTS
    -- =============================================================

    -- Sarah: first_blood (Jan), night_owl (Feb), streak_7 (Mar)
    INSERT INTO user_achievements (user_id, achievement_type_id, earned_at)
    VALUES
        (u1_id, a_first_blood, '2025-01-15 10:00:00+00'),
        (u1_id, a_night_owl,   '2025-02-10 03:15:00+00'),
        (u1_id, a_streak_7,    '2025-03-01 18:00:00+00');

    -- Marcus: first_blood (Mar), streak_7 (Apr)
    INSERT INTO user_achievements (user_id, achievement_type_id, earned_at)
    VALUES
        (u2_id, a_first_blood, '2025-03-22 11:00:00+00'),
        (u2_id, a_streak_7,    '2025-04-05 20:00:00+00');

    -- Elena: first_blood (Nov), night_owl (Dec), century (Feb)
    INSERT INTO user_achievements (user_id, achievement_type_id, earned_at)
    VALUES
        (u3_id, a_first_blood, '2024-11-05 12:00:00+00'),
        (u3_id, a_night_owl,   '2024-12-20 02:45:00+00'),
        (u3_id, a_century,     '2025-02-13 09:00:00+00');

    -- Kenji: first_blood (Jun)
    INSERT INTO user_achievements (user_id, achievement_type_id, earned_at)
    VALUES
        (u4_id, a_first_blood, '2025-06-10 09:30:00+00');

    -- =============================================================
    -- 6. ACTIVITY LOG (3-5 per user, timestamps match achievements)
    -- =============================================================

    -- Sarah (5 entries)
    INSERT INTO activity_log (id, user_id, kind, text, metadata, created_at)
    VALUES
        (gen_random_uuid(), u1_id, 'session_completed', 'Completed a 2-hour research coding session on waveform analysis.',       '{"session_id": "00000000-0000-0000-0000-000000000101", "duration_seconds": 7200, "lines_added": 340}'::JSONB,  '2026-07-07 14:30:00+00'),
        (gen_random_uuid(), u1_id, 'achievement_earned', 'Earned the First Blood badge!',                                          '{"achievement_key": "first_blood", "badge_class": "gold"}'::JSONB,                                               '2025-01-15 10:00:00+00'),
        (gen_random_uuid(), u1_id, 'achievement_earned', 'Earned the Night Owl badge!',                                           '{"achievement_key": "night_owl", "badge_class": "bronze"}'::JSONB,                                              '2025-02-10 03:15:00+00'),
        (gen_random_uuid(), u1_id, 'achievement_earned', 'Earned the 7-Day Streak badge!',                                        '{"achievement_key": "streak_7", "badge_class": "silver"}'::JSONB,                                               '2025-03-01 18:00:00+00'),
        (gen_random_uuid(), u1_id, 'streak_update',      'Current streak is 14 days — your longest is 28!',                      '{"current_streak": 14, "max_streak": 28}'::JSONB,                                                               '2026-07-07 14:30:00+00');

    -- Marcus (4 entries)
    INSERT INTO activity_log (id, user_id, kind, text, metadata, created_at)
    VALUES
        (gen_random_uuid(), u2_id, 'session_completed', 'Finished debugging a distributed systems simulation in Go.',             '{"session_id": "00000000-0000-0000-0000-000000000201", "duration_seconds": 5400, "lines_added": 120}'::JSONB,  '2026-07-06 22:15:00+00'),
        (gen_random_uuid(), u2_id, 'achievement_earned', 'Earned the First Blood badge!',                                          '{"achievement_key": "first_blood", "badge_class": "gold"}'::JSONB,                                               '2025-03-22 11:00:00+00'),
        (gen_random_uuid(), u2_id, 'achievement_earned', 'Earned the 7-Day Streak badge!',                                        '{"achievement_key": "streak_7", "badge_class": "silver"}'::JSONB,                                               '2025-04-05 20:00:00+00'),
        (gen_random_uuid(), u2_id, 'session_completed', 'Contributed to an open-source visualization library.',                  '{"session_id": "00000000-0000-0000-0000-000000000202", "duration_seconds": 3600, "commits": 3}'::JSONB,         '2026-07-05 16:00:00+00');

    -- Elena (5 entries)
    INSERT INTO activity_log (id, user_id, kind, text, metadata, created_at)
    VALUES
        (gen_random_uuid(), u3_id, 'session_completed', 'Deployed a new API endpoint for the dashboard backend.',                '{"session_id": "00000000-0000-0000-0000-000000000301", "duration_seconds": 4800, "deployments": 1}'::JSONB,     '2026-07-07 09:45:00+00'),
        (gen_random_uuid(), u3_id, 'achievement_earned', 'Earned the First Blood badge!',                                          '{"achievement_key": "first_blood", "badge_class": "gold"}'::JSONB,                                               '2024-11-05 12:00:00+00'),
        (gen_random_uuid(), u3_id, 'achievement_earned', 'Earned the Night Owl badge!',                                           '{"achievement_key": "night_owl", "badge_class": "bronze"}'::JSONB,                                              '2024-12-20 02:45:00+00'),
        (gen_random_uuid(), u3_id, 'achievement_earned', 'Earned the 100-Day Streak badge!',                                      '{"achievement_key": "century", "badge_class": "gold"}'::JSONB,                                                  '2025-02-13 09:00:00+00'),
        (gen_random_uuid(), u3_id, 'streak_update',      'Current streak is 31 days — keep it going!',                           '{"current_streak": 31, "max_streak": 45}'::JSONB,                                                               '2026-07-07 09:45:00+00');

    -- Kenji (3 entries)
    INSERT INTO activity_log (id, user_id, kind, text, metadata, created_at)
    VALUES
        (gen_random_uuid(), u4_id, 'session_completed', 'Ran a batch seismic inference pipeline in Rust.',                       '{"session_id": "00000000-0000-0000-0000-000000000401", "duration_seconds": 9000, "models_run": 12}'::JSONB,     '2026-07-07 16:00:00+00'),
        (gen_random_uuid(), u4_id, 'achievement_earned', 'Earned the First Blood badge!',                                          '{"achievement_key": "first_blood", "badge_class": "gold"}'::JSONB,                                               '2025-06-10 09:30:00+00'),
        (gen_random_uuid(), u4_id, 'session_completed', 'Refactored data-loading module for 3x throughput improvement.',         '{"session_id": "00000000-0000-0000-0000-000000000402", "duration_seconds": 6600, "throughput_gain_pct": 200}'::JSONB, '2026-07-06 18:30:00+00');

    -- =============================================================
    -- 7. USER PROBLEM STATS (placeholder defaults)
    -- =============================================================

    INSERT INTO user_problem_stats (user_id, solved_count, total_problems, attempting_count, rating, contribution_points, updated_at)
    VALUES
        (u1_id, 0, 0, 0, 0, 0, '2026-07-07 14:30:00+00'),
        (u2_id, 0, 0, 0, 0, 0, '2026-07-06 22:15:00+00'),
        (u3_id, 0, 0, 0, 0, 0, '2026-07-07 09:45:00+00'),
        (u4_id, 0, 0, 0, 0, 0, '2026-07-07 16:00:00+00');

END $$;
