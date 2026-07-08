-- name: CreateGuest :one
INSERT INTO users (id, display_name, type)
VALUES ( $1, $2, $3)
RETURNING *;


-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;
