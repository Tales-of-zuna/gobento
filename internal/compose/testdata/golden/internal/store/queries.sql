-- name: ListItems :many
SELECT id, name, created_at FROM items ORDER BY id;

-- name: CreateItem :one
INSERT INTO items (name) VALUES (sqlc.arg(name)) RETURNING id, name, created_at;
