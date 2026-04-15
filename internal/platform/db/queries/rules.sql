-- name: GetRule :one
SELECT * FROM rules WHERE id = $1;

-- name: ListRules :many
SELECT * FROM rules ORDER BY priority ASC, created_at DESC LIMIT $1 OFFSET $2;

-- name: ListRulesByTemplate :many
SELECT * FROM rules WHERE template_id = $1 ORDER BY priority ASC, created_at DESC LIMIT $2 OFFSET $3;

-- name: ListActiveRulesByTemplate :many
SELECT * FROM rules WHERE template_id = $1 AND status = 'active' ORDER BY priority ASC;

-- name: CreateRule :one
INSERT INTO rules (template_id, name, expression, severity, message, priority, status, mode)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateRule :one
UPDATE rules
SET name = $2, expression = $3, severity = $4, message = $5,
    priority = $6, status = $7, mode = $8, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteRule :exec
DELETE FROM rules WHERE id = $1;
