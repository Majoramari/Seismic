CREATE TABLE project_metadata
(
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID REFERENCES users (id) ON DELETE CASCADE,
    project_name     TEXT NOT NULL,
    repo_url         TEXT,
    website_url      TEXT,
    last_commit_hash TEXT,
    last_commit_at   TIMESTAMPTZ,
    last_synced_at   TIMESTAMPTZ      DEFAULT now(),
    UNIQUE (user_id, project_name)
);

CREATE TABLE project_commits
(
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID REFERENCES users (id) ON DELETE CASCADE,
    project_name TEXT NOT NULL,
    repo_url     TEXT,
    hash         TEXT NOT NULL,
    message      TEXT,
    author_name  TEXT,
    author_email TEXT,
    committed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ      DEFAULT now(),
    UNIQUE (user_id, project_name, hash)
);
