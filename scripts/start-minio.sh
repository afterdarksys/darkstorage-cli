#!/bin/bash
# Start local MinIO for development/testing

echo "🚀 Starting MinIO for Dark Storage development..."
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Start MinIO
docker compose up -d

echo ""
echo "✅ MinIO is starting..."
echo ""
echo "📊 MinIO Console: http://localhost:9001"
echo "🔌 MinIO API: http://localhost:9000"
echo ""
echo "Credentials:"
echo "  Username: darkstorage"
echo "  Password: darkstorage123"
echo ""
echo "Waiting for MinIO to be ready..."

# Wait for MinIO to be healthy
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if curl -sf http://localhost:9000/minio/health/live > /dev/null 2>&1; then
        echo ""
        echo "✅ MinIO is ready!"
        echo ""
        echo "Next steps:"
        echo "  1. Open console: http://localhost:9001"
        echo "  2. Login with credentials above"
        echo "  3. Create a bucket called 'test-bucket'"
        echo "  4. Run: go run main.go login"
        echo ""
        exit 0
    fi
    attempt=$((attempt + 1))
    sleep 1
done

echo ""
echo "⚠️  MinIO is taking longer than expected to start."
echo "Check logs with: docker compose logs -f"
exit 1
