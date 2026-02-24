#!/bin/bash
# Setup test environment with MinIO

echo "🔧 Setting up Dark Storage test environment..."
echo ""

# Wait for MinIO to be fully ready
echo "Waiting for MinIO API..."
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if curl -sf http://localhost:9000/minio/health/live > /dev/null 2>&1; then
        echo "✅ MinIO API is ready!"
        break
    fi
    attempt=$((attempt + 1))
    sleep 1
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ MinIO failed to start"
    exit 1
fi

echo ""
echo "📦 Creating test bucket..."

# Use MinIO client (mc) via Docker to create bucket
docker run --rm --network darkstorage-cli_default \
    --entrypoint=/bin/sh \
    minio/mc:latest \
    -c "
        mc alias set local http://darkstorage-minio:9000 darkstorage darkstorage123 && \
        mc mb local/test-bucket --ignore-existing && \
        echo '✅ Bucket created: test-bucket'
    "

echo ""
echo "✅ Test environment ready!"
echo ""
echo "MinIO Console: http://localhost:9001"
echo "Username: darkstorage"
echo "Password: darkstorage123"
echo ""
echo "Test bucket: test-bucket"
echo ""
echo "Next: go run main.go --help"
