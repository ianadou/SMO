-- name: CreateVote :one
INSERT INTO votes (id, match_id, voter_id, voted_id, score, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetVoteByID :one
SELECT * FROM votes WHERE id = $1;

-- name: ListVotesByMatchID :many
SELECT * FROM votes
WHERE match_id = $1
ORDER BY created_at ASC;

-- name: ListVotesByVoterID :many
SELECT * FROM votes
WHERE voter_id = $1
ORDER BY created_at DESC;

-- name: DeleteVote :exec
DELETE FROM votes WHERE id = $1;
