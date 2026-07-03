-- Tracks which heartbeats have already been turned into sessions
ALTER TABLE heartbeats
    ADD COLUMN processed BOOLEAN DEFAULT false;
CREATE INDEX idx_heartbeats_unprocessed ON heartbeats (user_id, processed) WHERE processed = false;
