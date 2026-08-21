-- name: CreateNotification :one
INSERT INTO notifications (recipient_id, message)
VALUES ($1, $2)
RETURNING *;

-- name: ListNotificationsForUser :many
SELECT * FROM notifications
WHERE recipient_id = $1
ORDER BY created_at DESC;

-- name: GetNotificationByID :one
SELECT * FROM notifications WHERE id = $1;

-- name: MarkNotificationRead :one
UPDATE notifications
SET read_at = now()
WHERE id = $1
RETURNING *;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications
SET read_at = now()
WHERE recipient_id = $1 AND read_at IS NULL;
