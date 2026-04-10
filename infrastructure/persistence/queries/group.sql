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
-- ============================================================================

-- name: CreateGroup :one
INSERT INTO groups (id, organizer_id, name, created_at)
VALUES ($1, $2, $3, $4)
RETURNING id, organizer_id, name, created_at;

-- name: GetGroupByID :one
SELECT id, organizer_id, name, created_at
FROM groups
WHERE id = $1;

-- name: ListGroupsByOrganizerID :many
SELECT id, organizer_id, name, created_at
FROM groups
WHERE organizer_id = $1
ORDER BY created_at DESC;

-- name: UpdateGroup :one
UPDATE groups
SET name = $2
WHERE id = $1
RETURNING id, organizer_id, name, created_at;

-- name: DeleteGroup :exec
DELETE FROM groups
WHERE id = $1;
