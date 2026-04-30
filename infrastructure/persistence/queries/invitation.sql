-- name: CreateInvitation :one
INSERT INTO invitations (id, match_id, player_id, token_hash, expires_at, used_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
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

-- name: CountConfirmedInvitationsByMatchID :one
SELECT COUNT(*) FROM invitations
WHERE match_id = $1 AND used_at IS NOT NULL;

-- name: ListConfirmedParticipantsByMatchID :many
SELECT
    invitations.id            AS invitation_id,
    invitations.player_id     AS player_id,
    players.name              AS player_name,
    invitations.used_at       AS used_at
FROM invitations
JOIN players ON players.id = invitations.player_id
WHERE invitations.match_id = $1
  AND invitations.used_at IS NOT NULL
ORDER BY invitations.used_at ASC;
