#!/usr/bin/env bash

set -a
source .env
set +a

migrate -path db/migrations -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSLMODE" up