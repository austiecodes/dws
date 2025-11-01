#!/bin/bash
# Quick connect to PostgreSQL via Docker
# Usage: ./connect.sh [optional_sql_command]

CONTAINER_NAME="my-postgres"
DB_USER="admin"
DB_NAME="postgres"

if [ -z "$1" ]; then
    # Interactive mode
    docker exec -it "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME"
else
    # Execute command mode
    docker exec -i "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME" -c "$1"
fi

