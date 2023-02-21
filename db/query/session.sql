-- name: CreateSession :one
INSERT INTO sessions (
  user_id, token
) VALUES (
  $1, $2
)
RETURNING *;
