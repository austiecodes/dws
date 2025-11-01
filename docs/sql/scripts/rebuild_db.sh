#!/bin/bash
# Complete database rebuild script for distributed architecture
# WARNING: This will DROP and RECREATE all tables!

set -e

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-dws}
DB_NAME=${DB_NAME:-dws}

echo "=== DWS Database Rebuild (Distributed Architecture) ==="
echo
echo "⚠️  WARNING: This will DELETE all existing data!"
echo "    Database: $DB_NAME"
echo "    Host: $DB_HOST:$DB_PORT"
echo
read -p "Are you sure? Type 'yes' to continue: " confirmation

if [ "$confirmation" != "yes" ]; then
    echo "Aborted."
    exit 1
fi

# Check if PostgreSQL is running
if ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER > /dev/null 2>&1; then
  echo "❌ PostgreSQL is not running at $DB_HOST:$DB_PORT"
  echo "   Please start it first: ./scripts/docker/pg.sh"
  exit 1
fi

echo
echo "✓ PostgreSQL is running"
echo

# Run rebuild
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << 'EOF'
\echo 'Dropping existing tables...'
DROP TABLE IF EXISTS tasks CASCADE;
DROP TABLE IF EXISTS containers CASCADE;
DROP TABLE IF EXISTS workers CASCADE;
DROP TABLE IF EXISTS users CASCADE;

\echo ''
\echo 'Dropping existing functions...'
DROP FUNCTION IF EXISTS set_timestamp() CASCADE;
DROP FUNCTION IF EXISTS set_containers_timestamp() CASCADE;
DROP FUNCTION IF EXISTS set_tasks_timestamp() CASCADE;
DROP FUNCTION IF EXISTS set_workers_timestamp() CASCADE;

\echo ''
\echo 'Creating tables in correct order...'
\i docs/sql/schema/01_users.sql
\i docs/sql/schema/04_workers.sql
\i docs/sql/schema/02_containers.sql
\i docs/sql/schema/03_tasks.sql

\echo ''
\echo '✓ Database rebuild complete!'
\echo ''
\echo 'Summary:'
SELECT 
    schemaname, 
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY tablename;
EOF

echo
echo "=== Rebuild Complete ==="
echo
echo "Next steps:"
echo "1. Start the platform: go run ./cmd/platform/main.go"
echo "2. Register a user and set is_admin=true in the database"
echo "3. Create worker nodes via the API or web UI"
echo
echo "Optional: Seed sample data"
echo "  INSERT INTO users (email, display_name, password_hash, is_admin)"
echo "  VALUES ('admin@example.com', 'Admin', '\$2a\$10\$...', true);"

