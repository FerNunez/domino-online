-- name: CreateUser :one
INSERT INTO users (id, email, hashed_password, display_name)
VALUES ( $1, $2, $3, $4)
RETURNING *;


-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;
