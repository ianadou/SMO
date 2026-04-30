-- +goose Up
-- +goose StatementBegin

-- Adds a NOT NULL player_id link to invitations. Each invitation is
-- now created FOR a specific player (the organizer picks who gets
-- the link). The accepted invitation (used_at IS NOT NULL) is the
-- source of truth for "who participates in this match".
--
-- Fail-loud guard: NOT NULL without a default cannot be applied to a
-- non-empty table. Rather than silently DELETE existing rows or
-- backfill with a placeholder player_id, the migration aborts with
-- a clear message so an operator can backfill manually before
-- retrying. SMO has zero invitations in any environment at the time
-- of this change; the guard is defense-in-depth for future
-- redeploys to environments we don't fully control.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM invitations LIMIT 1) THEN
        RAISE EXCEPTION 'Cannot add NOT NULL player_id: invitations table is non-empty. Manual backfill required before re-running this migration.';
    END IF;
END $$;

ALTER TABLE invitations
    ADD COLUMN player_id TEXT NOT NULL
        REFERENCES players(id) ON DELETE CASCADE;

-- Index supports two future hot paths: the per-player invitation
-- listing (e.g., "what matches has this player been invited to?")
-- and the ON DELETE CASCADE walk when a player is removed.
CREATE INDEX idx_invitations_player_id ON invitations(player_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_invitations_player_id;
ALTER TABLE invitations DROP COLUMN player_id;

-- +goose StatementEnd
