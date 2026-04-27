-- name: CreateOrganizer :one
-- Inserts a new organizer. The UNIQUE constraint on email rejects
-- duplicates; the repository translates that violation into
-- ErrEmailAlreadyExists.
INSERT INTO organizers (
    id, email, password_hash, display_name, created_at
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetOrganizerByID :one
SELECT * FROM organizers WHERE id = $1;

-- name: GetOrganizerByEmail :one
-- Email lookups go through this query exclusively. Emails are stored
-- lower-cased by the entity, so the WHERE clause does not need
-- LOWER() — but we still pass the lowered value at the application
-- layer for safety.
SELECT * FROM organizers WHERE email = $1;
