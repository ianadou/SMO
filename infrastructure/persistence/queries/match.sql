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
-- Updates only the status column. The state machine on the Match entity
-- controls which status transitions are valid; this query trusts the
-- caller and just persists the new value.
UPDATE matches
SET status = $2
WHERE id = $1
RETURNING *;

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
