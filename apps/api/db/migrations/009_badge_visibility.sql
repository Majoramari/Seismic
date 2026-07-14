CREATE TABLE hidden_badges
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID REFERENCES users (id) ON DELETE CASCADE,
    badge_type TEXT NOT NULL,
    created_at TIMESTAMPTZ      DEFAULT now(),
    UNIQUE (user_id, badge_type)
);