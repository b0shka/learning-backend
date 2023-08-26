-- name: CreateVerifyEmail :one
INSERT INTO verify_emails (
    id,
    email,
    secret_code,
    expires_at
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetVerifyEmail :one
SELECT * FROM verify_emails
WHERE email = $1 AND secret_code = $2 LIMIT 1;

-- name: DeleteVerifyEmailById :exec
DELETE FROM verify_emails
WHERE id = $1;

-- name: DeleteVerifyEmailByEmail :exec
DELETE FROM verify_emails
WHERE email = $1;