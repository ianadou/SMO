-- name: CreateInvitation :one
INSERT INTO invitations (id, match_id, player_id, token_hash, expires_at, response, responded_at, created_at)
VALUES ($1, $2, $3, $4, $5, 'pending', NULL, $6)
RETURNING *;

-- name: GetInvitationByID :one
SELECT * FROM invitations WHERE id = $1;

-- name: GetInvitationByTokenHash :one
SELECT * FROM invitations WHERE token_hash = $1;

-- name: ListInvitationsByMatchID :many
SELECT * FROM invitations
WHERE match_id = $1
ORDER BY created_at DESC;

-- name: UpdateInvitationResponse :one
UPDATE invitations
SET response = $2, responded_at = $3
WHERE id = $1
RETURNING *;

-- name: DeleteInvitation :exec
DELETE FROM invitations WHERE id = $1;

-- name: CountConfirmedInvitationsByMatchID :one
SELECT COUNT(*) FROM invitations
WHERE match_id = $1 AND response = 'yes';

-- name: LockMatchRow :one
SELECT id FROM matches WHERE id = $1 FOR UPDATE;

-- name: ListConfirmedParticipantsByMatchID :many
SELECT
    invitations.id            AS invitation_id,
    invitations.player_id     AS player_id,
    players.name              AS player_name,
    invitations.responded_at  AS responded_at
FROM invitations
JOIN players ON players.id = invitations.player_id
WHERE invitations.match_id = $1
  AND invitations.response = 'yes'
ORDER BY invitations.responded_at ASC;
