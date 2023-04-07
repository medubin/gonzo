-- name: CreateSession :one
INSERT INTO sessions (
  user_id, token
) VALUES (
  $1, $2
)
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE token = $1 and user_id = $2;