#!/usr/bin/env sh
set -eu

# envsubst '${API_HOST} ${API_PORT} ${SERVER_NAME}' < /nginx.conf > /etc/nginx/nginx.conf

cp /nginx.conf /etc/nginx/conf.d/default.conf

exec "$@"