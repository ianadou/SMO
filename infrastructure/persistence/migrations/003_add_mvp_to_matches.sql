-- +goose Up
-- +goose StatementBegin

-- Adds an MVP column on the matches table. The MVP of a match is the
-- player who received the highest average vote score from teammates and
-- is set when the match is finalized (transition completed → closed).
--
-- The column is nullable: a match that ends without any vote has no MVP,
-- and the FK uses ON DELETE SET NULL so deleting an MVP player does not
-- cascade-delete the match.
ALTER TABLE matches
    ADD COLUMN mvp_player_id TEXT NULL REFERENCES players(id) ON DELETE SET NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE matches DROP COLUMN mvp_player_id;

-- +goose StatementEnd
