#!/bin/sh
set -e

echo "🚀 Starting Resume Backend Service initialization..."
echo ""

# ============================================
# 1. Wait for Database
# ============================================
echo "⏳ Waiting for PostgreSQL..."

until PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; do
    echo "   PostgreSQL is unavailable - sleeping"
    sleep 2
done

echo "✅ PostgreSQL is ready"
echo ""

# ============================================
# 2. Run Migrations
# ============================================
echo "📊 Running database migrations..."

# Check if migrations table exists
MIGRATIONS_EXIST=$(PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -tAc \
    "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name='schema_migrations');")

if [ "$MIGRATIONS_EXIST" = "f" ]; then
    echo "   Creating migrations tracking table..."
    PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" <<-EOSQL
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT NOW()
        );
EOSQL
fi

# Run each migration if not already applied
for migration in /app/migrations/*.sql; do
    if [ -f "$migration" ]; then
        filename=$(basename "$migration")
        version="${filename%.sql}"

        # Check if migration already applied
        APPLIED=$(PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -tAc \
            "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version='$version');")

        if [ "$APPLIED" = "f" ]; then
            echo "   Applying migration: $filename"
            PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -f "$migration"

            # Mark as applied
            PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c \
                "INSERT INTO schema_migrations (version) VALUES ('$version');"

            echo "   ✅ Migration $filename applied"
        else
            echo "   ⏭️  Migration $filename already applied"
        fi
    fi
done

echo "✅ All migrations completed"
echo ""

# ============================================
# 3. Start Application
# ============================================
echo "🚀 Starting application..."
echo ""

exec "$@"
