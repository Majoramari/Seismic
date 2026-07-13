-- Public-facing contact email, separate from the private account
-- email used for login. Nullable — if not set, no email shows on
-- the public profile at all.
ALTER TABLE users
    ADD COLUMN contact_email VARCHAR(255);