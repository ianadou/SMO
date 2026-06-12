-- name: CreateMatchShareLink :one
INSERT INTO match_share_links (id, match_id, token_hash, expires_at, revoked_at, created_at)
VALUES ($1, $2, $3, $4, NULL, $5)
RETURNING *;

-- name: GetMatchShareLinkByTokenHash :one
SELECT * FROM match_share_links WHERE token_hash = $1;

-- name: GetActiveMatchShareLinkByMatchID :one
SELECT * FROM match_share_links
WHERE match_id = $1
  AND revoked_at IS NULL
  AND expires_at > NOW();

-- name: UpdateMatchShareLink :one
UPDATE match_share_links
SET revoked_at = $2
WHERE id = $1
RETURNING *;
