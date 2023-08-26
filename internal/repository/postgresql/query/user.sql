-- name: CreateUser :one
INSERT INTO users (
    id,
    email,
    username,
    photo
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: UpdateUser :exec
UPDATE users SET
  username = sqlc.arg(username),
  photo = COALESCE(sqlc.narg(photo), photo)
WHERE id = sqlc.arg(id);

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;