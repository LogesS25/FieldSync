-- name: UpsertPushToken :one
INSERT INTO push_tokens (user_id, token)
VALUES ($1, $2)
ON CONFLICT (token) DO UPDATE SET user_id = EXCLUDED.user_id
RETURNING *;

-- name: ListPushTokensForUser :many
SELECT * FROM push_tokens WHERE user_id = $1;

-- name: DeletePushToken :exec
DELETE FROM push_tokens WHERE token = $1;
