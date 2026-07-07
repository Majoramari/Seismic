-- Extends the users table with profile-specific fields not yet present
ALTER TABLE users
    ADD COLUMN first_name VARCHAR(50),
    ADD COLUMN role VARCHAR(50),
    ADD COLUMN location VARCHAR(100),
    ADD COLUMN university VARCHAR(100),
    ADD COLUMN time_zone VARCHAR(50),
    ADD COLUMN last_active_at TIMESTAMPTZ,
    ADD COLUMN gender VARCHAR(50),
    ADD COLUMN languages TEXT[]; -- Simple text array for display purposes on the profile
