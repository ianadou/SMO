-- +goose Up
-- +goose StatementBegin

-- Adds the final score of a match: goals scored by team A and team B.
--
-- Both columns are nullable: a match only has a score once it has been
-- completed (transition in_progress → completed records it). The winner
-- is NOT stored — it is derived from the score by Match.WinningSide(),
-- keeping a single source of truth and avoiding a denormalized,
-- drift-prone column.
ALTER TABLE matches
    ADD COLUMN score_a INTEGER NULL,
    ADD COLUMN score_b INTEGER NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE matches
    DROP COLUMN score_a,
    DROP COLUMN score_b;

-- +goose StatementEnd
