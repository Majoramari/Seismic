-- Enable UUID generation function
CREATE
EXTENSION IF NOT EXISTS "pgcrypto";

-- Users who sign up for Seismic.
CREATE TABLE users
(
    id               UUID PRIMARY KEY             DEFAULT gen_random_uuid(),
    username         VARCHAR(32) UNIQUE  NOT NULL,
    email            VARCHAR(255) UNIQUE NOT NULL,
    api_key          UUID UNIQUE         NOT NULL DEFAULT gen_random_uuid(),
    country          VARCHAR(2),
    bio              TEXT,
    website          TEXT,
    avatar_url       TEXT,
    avatar_public_id TEXT,
    created_at       TIMESTAMPTZ                  DEFAULT now(),
    deleted_at       TIMESTAMPTZ                  DEFAULT NULL
);

-- Temporary login tokens for magic link authentication.
-- When someone requests to log in, we create one of these,
-- email them a link containing the token, and delete it
-- once it's used (or once it expires).
CREATE TABLE magic_links
(
    id         UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    email      VARCHAR(255) NOT NULL,
    token      UUID         NOT NULL DEFAULT gen_random_uuid(),
    expires_at TIMESTAMPTZ  NOT NULL,
    used       BOOLEAN               DEFAULT false,
    created_at TIMESTAMPTZ           DEFAULT now()
);

-- This is the raw, unprocessed data.
-- A background job later groups these into "sessions".
CREATE TABLE heartbeats
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users (id) ON DELETE CASCADE,
    file        TEXT   NOT NULL,
    project     TEXT   NOT NULL,
    language    TEXT   NOT NULL,
    editor      TEXT   NOT NULL,
    branch      TEXT,
    os          TEXT,
    machine     TEXT,
    lines       INTEGER,
    cursor_line INTEGER,
    timezone    TEXT,
    time        BIGINT NOT NULL,
    received_at TIMESTAMPTZ      DEFAULT now()
);

-- Calculated coding sessions, built from heartbeats by a
-- background job that runs every 5 minutes.
CREATE TABLE sessions
(
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID REFERENCES users (id) ON DELETE CASCADE,
    project          TEXT        NOT NULL,
    language         TEXT        NOT NULL,
    start_time       TIMESTAMPTZ NOT NULL,
    end_time         TIMESTAMPTZ NOT NULL,
    duration_seconds INTEGER     NOT NULL
);

-- Achievements a user has earned. Awarded automatically by
-- a background job that checks conditions after each session
-- processing cycle (e.g. "coded past 2am on 5 different days").
CREATE TABLE badges
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID REFERENCES users (id) ON DELETE CASCADE,
    badge_type TEXT NOT NULL,
    earned_at  TIMESTAMPTZ      DEFAULT now(),
    UNIQUE (user_id, badge_type)
);

-- User-defined coding goals, like "code Go for 5 hours this
-- week" or "code 2 hours per day overall". Checked periodically
-- so we can remind the user by email if they're falling behind.
CREATE TABLE goals
(
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID REFERENCES users (id) ON DELETE CASCADE,
    scope             TEXT    NOT NULL CHECK (scope IN ('overall', 'language', 'project')),
    scope_value       TEXT,
    period TEXT NOT NULL CHECK (period IN ('daily', 'weekly')),
    target_seconds    INTEGER NOT NULL,
    reminders_enabled BOOLEAN          DEFAULT true,
    last_reminded_at  TIMESTAMPTZ,
    created_at        TIMESTAMPTZ      DEFAULT now(),
    active            BOOLEAN          DEFAULT true
);

CREATE TABLE posts
(
    id                    VARCHAR(11) PRIMARY KEY,
    user_id               UUID REFERENCES users (id) ON DELETE CASCADE,
    type                  TEXT NOT NULL CHECK (type IN ('post', 'article')),
    title                 TEXT,
    content               TEXT NOT NULL,
    cover_image_url       TEXT,
    cover_image_public_id TEXT,
    published             BOOLEAN     DEFAULT true,
    created_at            TIMESTAMPTZ DEFAULT now(),
    updated_at            TIMESTAMPTZ DEFAULT now()
);

-- Comments users leave on profiles or posts. Supports nested
-- replies (like Reddit threads) via parent_id.
CREATE TABLE comments
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users (id) ON DELETE CASCADE,
    target_type TEXT NOT NULL CHECK (target_type IN ('profile', 'post')),
    target_id   TEXT NOT NULL,
    parent_id   UUID REFERENCES comments (id) ON DELETE CASCADE,
    content     TEXT NOT NULL,
    likes_count INTEGER          DEFAULT 0,
    created_at  TIMESTAMPTZ      DEFAULT now(),
    deleted_at  TIMESTAMPTZ      DEFAULT NULL
);

-- Tracks who liked which comment, so we can prevent
-- someone liking the same comment twice and let them unlike.
CREATE TABLE comment_likes
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id UUID REFERENCES comments (id) ON DELETE CASCADE,
    user_id    UUID REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ      DEFAULT now(),
    UNIQUE (comment_id, user_id)
);

-- Follow relationships between developers. If A follows B,
-- A sees B's posts and achievements in their feed.
CREATE TABLE follows
(
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    follower_id  UUID REFERENCES users (id) ON DELETE CASCADE,
    following_id UUID REFERENCES users (id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ      DEFAULT now(),
    UNIQUE (follower_id, following_id),
    CHECK (follower_id != following_id
)
    );

-- Per-user privacy controls. One row per user, created
-- automatically when they sign up (with sensible defaults).
CREATE TABLE privacy_settings
(
    user_id          UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    hide_projects    BOOLEAN     DEFAULT false,
    hide_time        BOOLEAN     DEFAULT false,
    hide_languages   BOOLEAN     DEFAULT false,
    hide_leaderboard BOOLEAN     DEFAULT false,
    profile_public   BOOLEAN     DEFAULT true,
    updated_at       TIMESTAMPTZ DEFAULT now()
);

-- Lets a user hide specific individual projects from their
-- public profile and stats, rather than hiding all projects
-- at once via privacy_settings.hide_projects.
CREATE TABLE hidden_projects
(
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID REFERENCES users (id) ON DELETE CASCADE,
    project_name TEXT NOT NULL,
    created_at   TIMESTAMPTZ      DEFAULT now(),
    UNIQUE (user_id, project_name)
);