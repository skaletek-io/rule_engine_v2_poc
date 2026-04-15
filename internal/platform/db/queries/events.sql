-- name: GetEvent :one
SELECT * FROM events WHERE id = $1;

-- name: ListEvents :many
SELECT * FROM events ORDER BY received_at DESC LIMIT $1 OFFSET $2;

-- name: ListEventsByTemplate :many
SELECT * FROM events WHERE template_id = $1 ORDER BY received_at DESC LIMIT $2 OFFSET $3;

-- name: CreateEvent :one
INSERT INTO events (template_id, external_ref, payload, occurred_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateEvent :one
UPDATE events
SET external_ref = $2, payload = $3, occurred_at = $4
WHERE id = $1
RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM events WHERE id = $1;
