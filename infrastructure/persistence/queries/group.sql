-- ============================================================================
-- group.sql — Queries for the groups table.
--
-- Each query is annotated with `-- name: FunctionName :returnType` so sqlc
-- generates the corresponding Go function in
-- infrastructure/persistence/postgres/generated/group.sql.go
--
-- Return type annotations:
--   :one      → returns exactly one row (error if zero or more than one)
--   :many     → returns zero or more rows as a slice
--   :exec     → executes the query and returns only an error
--   :execrows → executes and returns the number of affected rows
--
-- Column order matches the table's physical order (id, organizer_id,
-- name, created_at, discord_webhook_url) so sqlc reuses the Groups
-- struct rather than generating per-query row types.
-- ============================================================================

-- name: CreateGroup :one
INSERT INTO groups (id, organizer_id, name, created_at, discord_webhook_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, organizer_id, name, created_at, discord_webhook_url;

-- name: GetGroupByID :one
SELECT id, organizer_id, name, created_at, discord_webhook_url
FROM groups
WHERE id = $1;

-- name: ListGroupsByOrganizerID :many
SELECT id, organizer_id, name, created_at, discord_webhook_url
FROM groups
WHERE organizer_id = $1
ORDER BY created_at DESC;

-- name: UpdateGroup :one
UPDATE groups
SET name                = $2,
    discord_webhook_url = $3
WHERE id = $1
RETURNING id, organizer_id, name, created_at, discord_webhook_url;

-- name: DeleteGroup :exec
DELETE FROM groups
WHERE id = $1;
