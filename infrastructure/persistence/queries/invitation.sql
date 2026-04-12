-- name: CreateInvitation :one
INSERT INTO invitations (id, match_id, token_hash, expires_at, used_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetInvitationByID :one
SELECT * FROM invitations WHERE id = $1;

-- name: GetInvitationByTokenHash :one
SELECT * FROM invitations WHERE token_hash = $1;

-- name: ListInvitationsByMatchID :many
SELECT * FROM invitations
WHERE match_id = $1
ORDER BY created_at DESC;

-- name: MarkInvitationAsUsed :one
UPDATE invitations
SET used_at = $2
WHERE id = $1
RETURNING *;

-- name: DeleteInvitation :exec
DELETE FROM invitations WHERE id = $1;
