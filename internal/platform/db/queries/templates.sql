-- name: GetTemplate :one
SELECT * FROM templates WHERE id = $1;

-- name: GetTemplateBySlug :one
SELECT * FROM templates WHERE slug = $1;

-- name: ListTemplates :many
SELECT * FROM templates ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CreateTemplate :one
INSERT INTO templates (slug, name, description, schema)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateTemplate :one
UPDATE templates
SET slug = $2, name = $3, description = $4, schema = $5, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTemplate :exec
DELETE FROM templates WHERE id = $1;
