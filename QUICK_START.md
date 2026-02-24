# Dark Storage CLI - Quick Start Guide

**Status**: ✅ WORKING! Real MinIO backend integrated!

---

## What's Working Right Now

✅ **Storage Backend**: Real MinIO S3-compatible storage
✅ **CLI Commands**: All basic operations working
✅ **Progress Bars**: Upload/download progress tracking
✅ **Error Handling**: Proper error messages
✅ **Local Testing**: Full environment ready

---

## Quick Start (5 Minutes)

### 1. Start MinIO

```bash
./scripts/start-minio.sh
```

**MinIO Console**: http://localhost:9001
**Credentials**: `darkstorage` / `darkstorage123`

### 2. Build the CLI

```bash
go build -o darkstorage main.go
```

### 3. Test It!

```bash
# List buckets
./darkstorage ls

# Create a bucket
./darkstorage mb my-bucket

# Upload a file
echo "Hello World" > test.txt
./darkstorage put test.txt my-bucket/

# List files
./darkstorage ls my-bucket/ -l

# Download a file
./darkstorage get my-bucket/test.txt ./downloaded.txt

# Copy a file
./darkstorage cp my-bucket/test.txt my-bucket/test-copy.txt

# View file contents
./darkstorage cat my-bucket/test.txt

# Delete a file
./darkstorage rm my-bucket/test-copy.txt

# Delete a bucket (must be empty)
./darkstorage rm my-bucket --force
```

---

## Available Commands

### Storage Operations
- `ls [bucket/path]` - List buckets or files
  - `-l` - Long format (with details)
  - `-r` - Recursive
- `mb <bucket>` - Make (create) bucket
- `put <local> <remote>` - Upload file
- `get <remote> [local]` - Download file
- `rm <path>` - Remove file or bucket
  - `--force` - Force delete bucket
- `cp <source> <dest>` - Copy file
- `mv <source> <dest>` - Move/rename file
- `cat <path>` - Display file contents

### Other Commands (Mocked - Not Yet Implemented)
- `login` - Authenticate
- `logout` - Remove credentials
- `whoami` - Show current user
- `groups` - Manage user groups
- `perms` - Manage permissions
- `shares` - Manage file shares
- `scan` - Scan files
- `trash` - Manage trash
- `audit` - View audit logs
- `config` - Manage configuration
- `version` - Show version

---

## Configuration

Default config (local MinIO):
- **Endpoint**: `localhost:9000`
- **Access Key**: `darkstorage`
- **Secret Key**: `darkstorage123`
- **Use SSL**: `false` (local testing)

### Environment Variables

```bash
export DARKSTORAGE_ENDPOINT="localhost:9000"
export DARKSTORAGE_ACCESS_KEY="darkstorage"
export DARKSTORAGE_SECRET_KEY="darkstorage123"
export DARKSTORAGE_USE_SSL="false"
```

### Production Config (Future)

```yaml
# ~/.darkstorage/config.yaml
storage:
  endpoint: storage.darkstorage.io
  access_key: your-access-key
  secret_key: your-secret-key
  use_ssl: true
  region: us-east-1
```

---

## What Works vs What's Coming

### ✅ Working Now (v0.1 - Local Testing)
- MinIO backend integration
- List buckets
- Create/delete buckets
- Upload files (with progress bar)
- Download files (with progress bar)
- Copy files
- Move files
- Delete files
- Cat (view) files
- Beautiful CLI output with colors

### 🚧 Coming Next (v0.2 - Core Features)
- Encryption (3+1 key system)
- Recursive upload/download
- Storage classes (STANDARD, GLACIER, etc.)
- Pre-signed URLs (sharing)
- File versioning
- Metadata management

### 🌟 Future (v0.3+ - Advanced Features)
- Web3 integration (Storj + IPFS)
- Desktop GUI (Fyne)
- Sync daemon
- Instant DR
- Disaster Mail
- HSM encryption
- Everything from the master vision!

---

## Testing Tips

### Test Upload Performance

```bash
# Create a large test file
dd if=/dev/zero of=large-file.bin bs=1M count=100

# Upload it (watch the progress bar!)
./darkstorage put large-file.bin test-bucket/
```

### Test Multiple Files

```bash
# Create test files
for i in {1..10}; do
  echo "File $i" > file$i.txt
  ./darkstorage put file$i.txt test-bucket/
done

# List them all
./darkstorage ls test-bucket/ -l
```

### Test AWS CLI Compatibility

```bash
# Dark Storage is S3-compatible!
aws s3 ls --endpoint-url http://localhost:9000
aws s3 cp test.txt s3://test-bucket/ --endpoint-url http://localhost:9000
```

---

## Troubleshooting

### MinIO not starting?

```bash
# Check if Docker is running
docker ps

# Check MinIO logs
docker logs darkstorage-minio

# Restart MinIO
docker compose down
./scripts/start-minio.sh
```

### Build errors?

```bash
# Clean and rebuild
go clean
go mod tidy
go build -o darkstorage main.go
```

### Connection refused?

```bash
# Make sure MinIO is running
curl http://localhost:9000/minio/health/live

# Check your config
cat ~/.darkstorage/config.yaml
```

---

## Development

### Project Structure

```
darkstorage-cli/
├── main.go                      # Entry point
├── cmd/
│   ├── root.go                  # Root command
│   ├── storage.go               # ✅ Storage commands (WORKING!)
│   ├── login.go                 # Authentication (mocked)
│   └── ...                      # Other commands (mocked)
├── internal/
│   ├── storage/
│   │   ├── types.go             # ✅ Storage interfaces
│   │   └── traditional.go       # ✅ MinIO backend (WORKING!)
│   ├── config/
│   │   └── storage.go           # ✅ Configuration (WORKING!)
│   └── ...
├── scripts/
│   ├── start-minio.sh           # ✅ Start MinIO
│   └── setup-test-env.sh        # ✅ Create test bucket
└── docker-compose.yml           # ✅ MinIO container
```

### Add a New Command

1. Add command in `cmd/your-command.go`
2. Register in `cmd/root.go` init function
3. Use `storageBackend` for storage operations
4. Test it!

---

## Next Steps

Ready to help build more! Pick what you want next:

1. **Encryption Layer** - Add the 3+1 key system
2. **Recursive Operations** - Upload/download folders
3. **Storage Classes** - GLACIER, DEEP_ARCHIVE support
4. **Pre-signed URLs** - Sharing functionality
5. **Web3 Integration** - Storj + IPFS backends
6. **GUI** - Desktop application
7. **Sync Daemon** - Background sync
8. **Deploy Production** - Real storage.darkstorage.io

---

## Success! 🚀

You now have a **working S3-compatible CLI** with:
- Real MinIO backend
- Beautiful progress bars
- Fast uploads/downloads
- S3 API compatibility
- Local development environment

**This is the foundation for the entire Dark Storage platform!**

Let's keep building! 🐱⚡

---

*Last Updated: 2026-02-24*
*Version: 0.1-alpha (Local Testing)*
