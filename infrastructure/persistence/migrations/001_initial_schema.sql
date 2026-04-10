-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- Initial schema for SMO (Sports Match Organizer).
--
-- All identifiers are stored as TEXT to match the string-based custom types
-- used in the domain layer (MatchID, PlayerID, etc.). Postgres treats TEXT
-- and VARCHAR identically internally; TEXT avoids arbitrary length limits.
--
-- All timestamps are TIMESTAMPTZ (timestamp with time zone). Postgres stores
-- them in UTC and converts to the session timezone on read, which prevents
-- a whole class of timezone bugs.
-- ============================================================================

-- ----------------------------------------------------------------------------
-- organizers
-- ----------------------------------------------------------------------------
-- The authenticated users of the system. Organizers own groups and create
-- matches. Players (the people who actually play) do NOT have an account;
-- they are referenced by an organizer or invited via a token.
CREATE TABLE organizers (
    id            TEXT        PRIMARY KEY,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    display_name  TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ----------------------------------------------------------------------------
-- groups
-- ----------------------------------------------------------------------------
-- A group is a recurring set of players that play matches together. Owned
-- by exactly one organizer.
CREATE TABLE groups (
    id           TEXT        PRIMARY KEY,
    organizer_id TEXT        NOT NULL REFERENCES organizers(id) ON DELETE CASCADE,
    name         TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_groups_organizer_id ON groups(organizer_id);

-- ----------------------------------------------------------------------------
-- players
-- ----------------------------------------------------------------------------
-- A player belongs to exactly one group. The ranking is updated after each
-- match by the ranking calculator (domain/ranking).
CREATE TABLE players (
    id         TEXT        PRIMARY KEY,
    group_id   TEXT        NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    name       TEXT        NOT NULL,
    ranking    INTEGER     NOT NULL DEFAULT 1000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_players_group_id ON players(group_id);

-- ----------------------------------------------------------------------------
-- matches
-- ----------------------------------------------------------------------------
-- A match belongs to exactly one group. The status column is constrained
-- to the values defined by the domain MatchStatus enum.
CREATE TABLE matches (
    id           TEXT        PRIMARY KEY,
    group_id     TEXT        NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    title        TEXT        NOT NULL,
    venue        TEXT        NOT NULL,
    scheduled_at TIMESTAMPTZ NOT NULL,
    status       TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_matches_status CHECK (
        status IN ('draft', 'open', 'teams_ready', 'in_progress', 'completed', 'closed')
    )
);

CREATE INDEX idx_matches_group_id ON matches(group_id);
CREATE INDEX idx_matches_status ON matches(status);

-- ----------------------------------------------------------------------------
-- teams
-- ----------------------------------------------------------------------------
-- A team belongs to exactly one match. There are typically two teams per
-- match but the schema does not enforce a maximum, leaving room for future
-- evolutions (3-team formats, tournament brackets).
CREATE TABLE teams (
    id       TEXT PRIMARY KEY,
    match_id TEXT NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    label    TEXT NOT NULL
);

CREATE INDEX idx_teams_match_id ON teams(match_id);

-- ----------------------------------------------------------------------------
-- team_players
-- ----------------------------------------------------------------------------
-- Junction table between teams and players. A player can be in only one team
-- per match (enforced by the use case, not by the schema, because the
-- per-match uniqueness requires joining through teams).
CREATE TABLE team_players (
    team_id   TEXT NOT NULL REFERENCES teams(id)   ON DELETE CASCADE,
    player_id TEXT NOT NULL REFERENCES players(id) ON DELETE CASCADE,

    PRIMARY KEY (team_id, player_id)
);

CREATE INDEX idx_team_players_player_id ON team_players(player_id);

-- ----------------------------------------------------------------------------
-- votes
-- ----------------------------------------------------------------------------
-- Post-match peer ratings. Each vote is cast by one player for another
-- player on the same team. Score is an integer in [1, 5].
--
-- A player cannot vote for themselves (enforced both by the domain and
-- by a CHECK constraint here for defense in depth).
CREATE TABLE votes (
    id         TEXT        PRIMARY KEY,
    match_id   TEXT        NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    voter_id   TEXT        NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    voted_id   TEXT        NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    score      INTEGER     NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_votes_score CHECK (score BETWEEN 1 AND 5),
    CONSTRAINT chk_votes_no_self_vote CHECK (voter_id <> voted_id)
);

CREATE INDEX idx_votes_match_id ON votes(match_id);
CREATE INDEX idx_votes_voted_id ON votes(voted_id);

-- ----------------------------------------------------------------------------
-- invitations
-- ----------------------------------------------------------------------------
-- Invitation tokens that allow non-authenticated people to join a specific
-- match. The token is stored as a hash; the plain token is only returned
-- once at creation time and shared by the organizer with the invitee.
CREATE TABLE invitations (
    id         TEXT        PRIMARY KEY,
    match_id   TEXT        NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_invitations_expires_after_creation CHECK (expires_at > created_at)
);

CREATE INDEX idx_invitations_match_id ON invitations(match_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS invitations;
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS team_players;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS matches;
DROP TABLE IF EXISTS players;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS organizers;

-- +goose StatementEnd
