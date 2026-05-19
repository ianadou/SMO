-- +goose Up
-- +goose StatementBegin

-- Replaces the one-shot used_at semantics with a mutable yes/no
-- response. The used_at column is intentionally KEPT (legacy, no longer
-- written by the application) so this migration is non-destructive and
-- the rollback is trivial.
--
-- response defaults to 'pending' so existing rows that were never
-- "used" land in the correct initial state. A CHECK constraint pins
-- the domain enum at the database boundary (defense in depth: the
-- entity validates it too).
ALTER TABLE invitations
    ADD COLUMN response TEXT NOT NULL DEFAULT 'pending'
        CHECK (response IN ('pending', 'yes', 'no')),
    ADD COLUMN responded_at TIMESTAMPTZ;

-- Backfill: a previously "used" invitation was a confirmed attendance,
-- so it maps to response='yes' with responded_at carrying the original
-- used_at timestamp. Pending (used_at IS NULL) rows keep the default.
UPDATE invitations
SET response = 'yes',
    responded_at = used_at
WHERE used_at IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE invitations DROP COLUMN responded_at;
ALTER TABLE invitations DROP COLUMN response;

-- +goose StatementEnd
