-- name: CreatePlayer :one
INSERT INTO players (id, group_id, name, ranking, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPlayerByID :one
SELECT * FROM players WHERE id = $1;

-- name: ListPlayersByGroupID :many
SELECT * FROM players
WHERE group_id = $1
ORDER BY ranking DESC, name ASC;

-- name: UpdatePlayerRanking :one
UPDATE players
SET ranking = $2
WHERE id = $1
RETURNING *;

-- name: DeletePlayer :exec
DELETE FROM players WHERE id = $1;
