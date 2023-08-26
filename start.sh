#!/bin/sh

set -e

echo "run db migration"
/app/migrate -path /app/migration -database "postgresql://$POSTGRESQL_USER:$POSTGRESQL_PASSWORD@$POSTGRESQL_HOST:$POSTGRESQL_PORT/$POSTGRESQL_DB_NAME?sslmode=disable" -verbose up

echo "start the app"
exec "$@"