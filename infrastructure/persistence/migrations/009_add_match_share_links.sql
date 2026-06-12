-- +goose Up
-- +goose StatementBegin

-- One shareable link per match: the organizer drops a single URL in the
-- group chat instead of sending each player a personal invitation link.
-- Like invitations, only the token hash is stored; the plain token is
-- returned once at creation time.
--
-- revoked_at is nullable: NULL means the link was never revoked. A link
-- is active when revoked_at IS NULL and expires_at is in the future.
-- Regenerating a link revokes the previous one, so a match has at most
-- one active link (enforced by the use case, not by the schema, because
-- "active" depends on the current time).
CREATE TABLE match_share_links (
    id         TEXT        PRIMARY KEY,
    match_id   TEXT        NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_match_share_links_expires_after_creation CHECK (expires_at > created_at)
);

-- Index supports the "active link of this match" lookup used by both
-- the organizer panel and the regenerate-revokes-previous flow.
CREATE INDEX idx_match_share_links_match_id ON match_share_links(match_id);

-- claimed_at records when a player claimed their invitation through a
-- match share link (the claim rotates the personal token). NULL means
-- never claimed. Existing rows predate share links, so NULL is the
-- correct state for all of them and no backfill is needed.
ALTER TABLE invitations
    ADD COLUMN claimed_at TIMESTAMPTZ;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE invitations DROP COLUMN claimed_at;
DROP TABLE IF EXISTS match_share_links;

-- +goose StatementEnd
