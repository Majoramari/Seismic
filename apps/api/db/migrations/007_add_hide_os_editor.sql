ALTER TABLE privacy_settings
    ADD COLUMN hide_os BOOLEAN DEFAULT false,
    ADD COLUMN hide_editor BOOLEAN DEFAULT false;