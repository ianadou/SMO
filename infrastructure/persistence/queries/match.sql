-- name: CreateMatch :one
-- Inserts a new match row. All fields are required; the status is
-- typically 'draft' at creation time and evolves through the
-- Match.Open(), MarkTeamsReady(), etc. state machine methods.
INSERT INTO matches (
    id, group_id, title, venue, scheduled_at, status, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetMatchByID :one
SELECT * FROM matches WHERE id = $1;

-- name: ListMatchesByGroupID :many
-- Returns all matches for a group, ordered by scheduled date descending
-- so upcoming and recent matches appear first.
SELECT * FROM matches
WHERE group_id = $1
ORDER BY scheduled_at DESC;

-- name: UpdateMatchStatus :one
-- Persists the status and the final score. The state machine on the
-- Match entity controls which transitions are valid; this query trusts
-- the caller. Score is nil for every transition except complete (and
-- finalize uses its own query, so a recorded score is never clobbered
-- here).
UPDATE matches
SET status  = $2,
    score_a = $3,
    score_b = $4
WHERE id = $1
RETURNING *;

-- name: GetLatestDecidedMatchByGroup :one
-- Returns the group's most recent decided match (completed or closed,
-- with a non-draw score), excluding a given match id, ordered by
-- scheduled date. Feeds the "top player joins the previous winner" rule.
SELECT * FROM matches
WHERE group_id = $1
  AND id <> $2
  AND status IN ('completed', 'closed')
  AND score_a IS NOT NULL
  AND score_b IS NOT NULL
  AND score_a <> score_b
ORDER BY scheduled_at DESC
LIMIT 1;

-- name: FinalizeMatch :one
-- Atomic finalize: sets the MVP and the new status (typically 'closed')
-- in a single statement. Used by FinalizeMatchUseCase to avoid a window
-- where MVP is set but status hasn't transitioned yet, or vice versa.
UPDATE matches
SET mvp_player_id = $2,
    status        = $3
WHERE id = $1
RETURNING *;

-- name: DeleteMatch :exec
DELETE FROM matches WHERE id = $1;

-- name: CountClosedMatchesTogether :many
-- Read model for the vote page "matchs joués ensemble" meta: for each
-- other player, how many closed matches of the group both they and the
-- reference player attended (confirmed invitations on both sides).
SELECT i2.player_id, COUNT(*) AS shared_matches
FROM matches m
JOIN invitations i1
  ON i1.match_id = m.id AND i1.player_id = sqlc.arg(player_id) AND i1.response = 'yes'
JOIN invitations i2
  ON i2.match_id = m.id AND i2.player_id = ANY(sqlc.arg(other_player_ids)::text[]) AND i2.response = 'yes'
WHERE m.group_id = sqlc.arg(group_id) AND m.status = 'closed'
GROUP BY i2.player_id;
