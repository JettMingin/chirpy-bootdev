-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, pw_hash)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users *;

-- name: LookupUser :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users SET email = $1, pw_hash = $2, updated_at = NOW() WHERE id = $3
RETURNING  *;

-- name: UpgradeUserToRed :exec
UPDATE users SET is_chirpy_red = true, updated_at = NOW() WHERE id = $1;