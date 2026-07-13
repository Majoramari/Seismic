-- Adds a keystroke count to each heartbeat, so editor plugins can
-- report how many characters were typed since the last heartbeat.
-- Nullable since existing rows and older plugin versions won't send it.
ALTER TABLE heartbeats
    ADD COLUMN keystrokes INTEGER;