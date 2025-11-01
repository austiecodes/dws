#!/bin/bash
# Initialize database from scratch
# Usage: ./init_db.sh

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

echo -e "${YELLOW}=== DWS Database Initialization ===${NC}"
echo "Container: $CONTAINER_NAME"
echo "Database: $DB_NAME"
echo ""

# Check if container is running
if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo -e "${RED}Error: Container '$CONTAINER_NAME' is not running${NC}"
    echo "Start it with: docker start $CONTAINER_NAME"
    exit 1
fi

# Test connection
echo -e "${YELLOW}Testing database connection...${NC}"
if ! docker exec -i "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "${RED}Error: Cannot connect to database${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Connected${NC}"

# Drop existing tables (WARNING: destructive!)
echo ""
echo -e "${RED}WARNING: This will DROP all existing tables!${NC}"
read -p "Are you sure? (type 'yes' to confirm): " confirm
if [ "$confirm" != "yes" ]; then
    echo "Aborted."
    exit 0
fi

echo -e "${YELLOW}Dropping existing tables...${NC}"
docker exec -i "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME" <<EOF
DROP TABLE IF EXISTS tasks CASCADE;
DROP TABLE IF EXISTS containers CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP FUNCTION IF EXISTS set_timestamp() CASCADE;
DROP FUNCTION IF EXISTS set_containers_timestamp() CASCADE;
EOF
echo -e "${GREEN}✓ Tables dropped${NC}"

# Execute schema files in order
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCHEMA_DIR="$SCRIPT_DIR/../schema"

for sql_file in "$SCHEMA_DIR"/*.sql; do
    filename=$(basename "$sql_file")
    echo -e "${YELLOW}Executing $filename...${NC}"
    docker exec -i "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME" < "$sql_file"
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ $filename completed${NC}"
    else
        echo -e "${RED}✗ $filename failed${NC}"
        exit 1
    fi
done

# Verify tables created
echo ""
echo -e "${YELLOW}Verifying tables...${NC}"
docker exec -i "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME" -c "\dt"

echo ""
echo -e "${GREEN}=== Database initialization complete! ===${NC}"

