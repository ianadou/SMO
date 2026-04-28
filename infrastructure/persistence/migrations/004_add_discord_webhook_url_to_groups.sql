-- +goose Up
-- +goose StatementBegin

-- Adds an optional Discord webhook URL to groups. When the organizer
-- triggers the teams_ready transition, a notification is posted to
-- this URL (wiring lands in a follow-up PR).
--
-- The CHECK constraint is defense-in-depth: the entity layer enforces
-- the same maximum length, but a direct INSERT/UPDATE bypassing the
-- application would be caught at the database boundary as well.
ALTER TABLE groups
    ADD COLUMN discord_webhook_url TEXT NULL
        CHECK (discord_webhook_url IS NULL OR LENGTH(discord_webhook_url) <= 2048);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE groups DROP COLUMN discord_webhook_url;

-- +goose StatementEnd
