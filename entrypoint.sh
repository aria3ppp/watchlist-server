#!/bin/bash -e

echo "[`date`] Running entrypoint script..."

DSN="postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB?sslmode=disable"

echo "[`date`] Running DB migrations..."
migrate -database "${DSN}" -path ./migrations up

echo "[`date`] Starting server..."
./server