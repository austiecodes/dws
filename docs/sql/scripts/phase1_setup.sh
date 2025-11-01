#!/bin/bash
# Quick start script for Phase 1: Worker Management

set -e

echo "=== Phase 1: Worker Management Setup ==="
echo

# Check if PostgreSQL is running
if ! pg_isready -h localhost -p 5432 -U dws > /dev/null 2>&1; then
  echo "❌ PostgreSQL is not running. Please start it first:"
  echo "   ./scripts/docker/pg.sh"
  exit 1
fi

echo "✓ PostgreSQL is running"

# Run migrations
echo
echo "Running migrations..."
psql -h localhost -p 5432 -U dws -d dws << EOF
\echo 'Creating workers table...'
\i docs/sql/schema/04_workers.sql

\echo 'Adding worker_id to containers...'
\i docs/sql/migrations/20251101_add_worker_id_to_containers.sql

\echo 'Done!'
EOF

echo
echo "=== Setup Complete ==="
echo
echo "You can now:"
echo "1. Start the platform: go run ./cmd/platform/main.go"
echo "2. Start the scheduler: go run ./cmd/scheduler/main.go"  
echo "3. Start the worker: go run ./cmd/worker/main.go"
echo "4. Access the web UI: cd web && yarn dev"
echo
echo "Worker management is available at: http://localhost:5173/workers"
echo "(Admin users can create/edit/delete workers, all users can view)"

