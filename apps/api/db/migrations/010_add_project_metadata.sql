ALTER TABLE heartbeats
    ADD COLUMN repo_url TEXT,
    ADD COLUMN website_url TEXT,
    ADD COLUMN last_commit_at TIMESTAMPTZ;

CREATE TABLE project_settings
(
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID REFERENCES users (id) ON DELETE CASCADE,
    project_name TEXT NOT NULL,
    archived     BOOLEAN          DEFAULT false,
    updated_at   TIMESTAMPTZ      DEFAULT now(),
    UNIQUE (user_id, project_name)
);
