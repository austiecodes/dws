#!/bin/bash
# Apply incremental migrations to existing database
# Usage: ./migrate.sh

set -e

# Configuration (from app.toml)
CONTAINER_NAME="my-postgres"
DB_USER="admin"
DB_NAME="postgres"
PGPASSWORD="123456"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== DWS Database Migration ===${NC}"
echo "Container: $CONTAINER_NAME"
echo "Database: $DB_NAME"
echo ""

# Check if container is running
if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo -e "${RED}Error: Container '$CONTAINER_NAME' is not running${NC}"
    exit 1
fi

# Test connection
echo -e "${YELLOW}Testing database connection...${NC}"
if ! docker exec -i "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "${RED}Error: Cannot connect to database${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Connected${NC}"

# Execute migration files in order
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATIONS_DIR="$SCRIPT_DIR/../migrations"

echo ""
echo -e "${YELLOW}Available migrations:${NC}"
ls -1 "$MIGRATIONS_DIR"/*.sql 2>/dev/null || {
    echo "No migration files found."
    exit 0
}

echo ""
read -p "Apply all migrations? (y/n): " confirm
if [ "$confirm" != "y" ]; then
    echo "Aborted."
    exit 0
fi

for sql_file in "$MIGRATIONS_DIR"/*.sql; do
    filename=$(basename "$sql_file")
    echo -e "${YELLOW}Applying $filename...${NC}"
    docker exec -i "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME" < "$sql_file"
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ $filename completed${NC}"
    else
        echo -e "${RED}✗ $filename failed${NC}"
        echo "Migration stopped. Fix the error and retry."
        exit 1
    fi
done

echo ""
echo -e "${GREEN}=== All migrations applied successfully! ===${NC}"

