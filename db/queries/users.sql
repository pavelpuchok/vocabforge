-- name: CreateOrUpdateUser :one
INSERT INTO users ( telegram_id, telegram_chat_id, name, created_at, updated_at ) VALUES (?, ?, ?, ?, ?)
ON CONFLICT (telegram_id, telegram_chat_id) DO UPDATE SET name=excluded.name, updated_at=excluded.updated_at
RETURNING *;

-- name: GetUserByTelegramID :one
SELECT * FROM users WHERE telegram_id = ? AND telegram_chat_id = ?;

