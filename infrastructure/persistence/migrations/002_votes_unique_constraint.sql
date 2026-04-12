-- +goose Up
-- +goose StatementBegin

-- Ensure a voter can cast at most one vote per (voted player) per match.
-- This matches the "peer rating" semantics: each rater has one opinion
-- per teammate per match, not multiple overlapping opinions.
ALTER TABLE votes
ADD CONSTRAINT uq_votes_match_voter_voted
UNIQUE (match_id, voter_id, voted_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE votes
DROP CONSTRAINT IF EXISTS uq_votes_match_voter_voted;

-- +goose StatementEnd
