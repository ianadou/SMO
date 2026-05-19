-- name: DeleteMatchTeamMembers :exec
DELETE FROM match_team_members WHERE match_id = $1;

-- name: InsertMatchTeamMember :exec
INSERT INTO match_team_members (match_id, player_id, team, slot)
VALUES ($1, $2, $3, $4);

-- name: ListMatchTeamMembers :many
-- Raw membership rows for a match, ordered for deterministic rehydration.
SELECT match_id, player_id, team, slot
FROM match_team_members
WHERE match_id = $1
ORDER BY team, slot;

-- name: ListMatchTeamMembersWithPlayers :many
-- Read model: team membership joined with player display data, used by
-- GetMatchTeamsUseCase to render the DS field/present list.
SELECT mtm.player_id, mtm.team, mtm.slot, p.name
FROM match_team_members mtm
JOIN players p ON p.id = mtm.player_id
WHERE mtm.match_id = $1
ORDER BY mtm.team, mtm.slot;
