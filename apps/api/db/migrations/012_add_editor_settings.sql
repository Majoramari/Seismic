CREATE TABLE editor_settings
(
    id                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                   UUID REFERENCES users (id) ON DELETE CASCADE,
    use_git_root_project_name BOOLEAN          DEFAULT true,
    updated_at                TIMESTAMPTZ      DEFAULT now(),
    UNIQUE (user_id)
);
