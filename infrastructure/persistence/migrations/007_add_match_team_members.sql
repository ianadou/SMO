-- +goose Up
-- +goose StatementBegin

-- One row per player assigned to a match team. team is pinned to the
-- domain enum at the DB boundary (defense in depth: the entity also
-- validates). slot preserves the on-field ordering chosen by the
-- organizer. The whole set for a match is replaced atomically by the
-- repository (delete-all + insert) inside a transaction.
CREATE TABLE match_team_members (
    match_id   TEXT NOT NULL REFERENCES matches (id) ON DELETE CASCADE,
    player_id  TEXT NOT NULL REFERENCES players (id) ON DELETE CASCADE,
    team       TEXT NOT NULL CHECK (team IN ('A', 'B')),
    slot       INTEGER NOT NULL,
    PRIMARY KEY (match_id, player_id)
);

CREATE INDEX idx_match_team_members_match ON match_team_members (match_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE match_team_members;
-- +goose StatementEnd
