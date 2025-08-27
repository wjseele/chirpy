-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET updated_at = now(), revoked_at = now()
WHERE token = $1;
