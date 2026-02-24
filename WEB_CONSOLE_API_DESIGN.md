# Web Console API Design - API Key Management

**Hostname**: `console.darkstorage.io`
**Purpose**: Web-based management interface for Dark Storage accounts
**Status**: Design Document

---

## Architecture Overview

```
User → Web Console (React/Next.js) → API Server (Go) → PostgreSQL
                                   ↓
                            S3 Storage (MinIO)
```

### Components

1. **Web Frontend** - `console.darkstorage.io`
   - React/Next.js single-page application
   - User authentication (OAuth/SSO)
   - API key management UI
   - Storage browser
   - Settings & billing

2. **API Server** - `api.darkstorage.io`
   - RESTful API
   - JWT-based authentication
   - API key generation & validation
   - User management
   - Storage operations proxy

3. **Desktop GUI** - Fyne application (already exists)
   - Local sync management
   - Talks to local daemon

4. **CLI** - Command-line tool (already exists)
   - Uses API keys for authentication
   - Direct S3 operations

---

## API Key Management Flow

### 1. User Journey

```
1. User logs into console.darkstorage.io (OAuth/SSO)
2. Navigate to Settings → API Keys
3. Click "Generate New API Key"
4. Choose:
   - Key name/description
   - Permissions (read-only, read-write, admin)
   - Expiration (30 days, 90 days, 1 year, never)
   - IP restrictions (optional)
5. Click "Generate"
6. Key is displayed ONCE
7. User can:
   - Copy key to clipboard
   - Download config file (darkstorage-config.yaml)
   - Download CLI setup script
8. Key is saved in database (hashed)
```

### 2. Config File Format

**Filename**: `darkstorage-config.yaml`

```yaml
# Dark Storage CLI Configuration
# Generated: 2026-02-24 10:30:00 UTC
# Account: user@example.com
# Key ID: dk_live_abc123xyz

version: "1.0"

# Authentication
auth:
  api_key: "dk_live_abc123xyz789def456ghi012jkl345mno678pqr901stu234vwx567yza890bcd"
  endpoint: "https://storage.darkstorage.io"

# Storage Configuration
storage:
  endpoint: "storage.darkstorage.io"
  access_key: "AKIAIOSFODNN7EXAMPLE"  # Generated from API key
  secret_key: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"  # Generated from API key
  region: "us-east-1"
  use_ssl: true

# User Information (read-only, for reference)
user:
  email: "user@example.com"
  account_id: "acc_abc123"
  plan: "professional"

# Key Metadata (read-only, for reference)
key:
  id: "dk_live_abc123xyz"
  name: "Production Server - Web01"
  created_at: "2026-02-24T10:30:00Z"
  expires_at: "2027-02-24T10:30:00Z"
  permissions:
    - "storage:read"
    - "storage:write"
    - "storage:delete"
  ip_restrictions:
    - "203.0.113.0/24"
    - "198.51.100.42"

# Optional: Advanced Settings
advanced:
  max_retries: 3
  timeout: 30
  bandwidth_limit: 0  # 0 = unlimited (MB/s)
```

### 3. Installation Script

**Filename**: `install-darkstorage.sh`

```bash
#!/bin/bash
# Dark Storage CLI Installation Script
# Generated for: user@example.com
# Key ID: dk_live_abc123xyz

set -e

echo "Installing Dark Storage CLI..."

# Detect OS
OS="$(uname -s)"
case "$OS" in
    Linux*)     PLATFORM=linux;;
    Darwin*)    PLATFORM=darwin;;
    *)          echo "Unsupported OS: $OS"; exit 1;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64)     ARCH=amd64;;
    arm64)      ARCH=arm64;;
    aarch64)    ARCH=arm64;;
    *)          echo "Unsupported architecture: $ARCH"; exit 1;;
esac

# Download CLI
VERSION="latest"
BINARY="darkstorage-${PLATFORM}-${ARCH}"
URL="https://downloads.darkstorage.io/cli/${VERSION}/${BINARY}"

echo "Downloading Dark Storage CLI for ${PLATFORM}/${ARCH}..."
curl -fsSL "$URL" -o /tmp/darkstorage
chmod +x /tmp/darkstorage

# Install
if [ -w /usr/local/bin ]; then
    sudo mv /tmp/darkstorage /usr/local/bin/darkstorage
else
    mv /tmp/darkstorage "$HOME/.local/bin/darkstorage"
    echo "Installed to $HOME/.local/bin/darkstorage"
    echo "Add to PATH: export PATH=\"\$HOME/.local/bin:\$PATH\""
fi

# Create config directory
mkdir -p "$HOME/.darkstorage"

# Save embedded config
cat > "$HOME/.darkstorage/config.yaml" << 'EOF'
# THIS CONFIG IS EMBEDDED IN THE SCRIPT
# Download the actual config file from the web console
EOF

echo ""
echo "✓ Dark Storage CLI installed successfully!"
echo ""
echo "Next steps:"
echo "  1. Download your config file from the web console"
echo "  2. Save it to: $HOME/.darkstorage/config.yaml"
echo "  3. Run: darkstorage whoami"
echo ""
echo "Or use your API key directly:"
echo "  darkstorage login --key YOUR_API_KEY"
echo ""
```

---

## API Endpoints

### Base URL: `https://api.darkstorage.io/v1`

### Authentication

All requests require JWT token in header:
```
Authorization: Bearer <jwt_token>
```

### API Key Management Endpoints

#### 1. List API Keys

```http
GET /api-keys
```

**Response:**
```json
{
  "keys": [
    {
      "id": "key_abc123",
      "name": "Production Server",
      "key_prefix": "dk_live_abc123",
      "created_at": "2026-02-24T10:30:00Z",
      "last_used": "2026-02-24T15:45:00Z",
      "expires_at": "2027-02-24T10:30:00Z",
      "permissions": ["storage:read", "storage:write"],
      "status": "active"
    }
  ],
  "total": 3
}
```

#### 2. Generate New API Key

```http
POST /api-keys
Content-Type: application/json

{
  "name": "Production Server - Web01",
  "permissions": ["storage:read", "storage:write", "storage:delete"],
  "expires_in_days": 365,
  "ip_restrictions": ["203.0.113.0/24"],
  "metadata": {
    "environment": "production",
    "server": "web01"
  }
}
```

**Response:**
```json
{
  "key": {
    "id": "key_abc123",
    "name": "Production Server - Web01",
    "api_key": "dk_live_abc123xyz789def456ghi012jkl345mno678pqr901stu234vwx567yza890bcd",
    "s3_access_key": "AKIAIOSFODNN7EXAMPLE",
    "s3_secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
    "created_at": "2026-02-24T10:30:00Z",
    "expires_at": "2027-02-24T10:30:00Z",
    "permissions": ["storage:read", "storage:write", "storage:delete"]
  },
  "warning": "This is the only time the API key will be displayed. Save it securely."
}
```

#### 3. Download Config File

```http
GET /api-keys/{key_id}/config?format=yaml
```

**Response:**
```
Content-Type: application/x-yaml
Content-Disposition: attachment; filename="darkstorage-config.yaml"

[YAML config file content]
```

#### 4. Download Install Script

```http
GET /api-keys/{key_id}/install-script
```

**Response:**
```
Content-Type: text/x-shellscript
Content-Disposition: attachment; filename="install-darkstorage.sh"

[Bash script content]
```

#### 5. Revoke API Key

```http
DELETE /api-keys/{key_id}
```

**Response:**
```json
{
  "success": true,
  "message": "API key revoked successfully"
}
```

#### 6. Update API Key

```http
PATCH /api-keys/{key_id}
Content-Type: application/json

{
  "name": "Updated Name",
  "ip_restrictions": ["203.0.113.0/24", "198.51.100.42"]
}
```

#### 7. Get Key Usage Stats

```http
GET /api-keys/{key_id}/usage?period=30d
```

**Response:**
```json
{
  "key_id": "key_abc123",
  "period": "30d",
  "stats": {
    "total_requests": 15420,
    "bandwidth_used": 52428800,  // bytes
    "operations": {
      "upload": 500,
      "download": 1200,
      "delete": 50,
      "list": 13670
    },
    "last_used": "2026-02-24T15:45:00Z",
    "last_used_ip": "203.0.113.42"
  }
}
```

---

## Database Schema

### Table: `api_keys`

```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Key Information
    name VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(50) NOT NULL UNIQUE,  -- "dk_live_abc123" for display
    key_hash VARCHAR(255) NOT NULL,          -- bcrypt hash of full key

    -- S3 Credentials (encrypted at rest)
    s3_access_key VARCHAR(255) NOT NULL,
    s3_secret_key_encrypted BYTEA NOT NULL,  -- AES-256 encrypted

    -- Permissions
    permissions JSONB NOT NULL DEFAULT '[]',
    ip_restrictions JSONB DEFAULT '[]',

    -- Status & Lifecycle
    status VARCHAR(20) NOT NULL DEFAULT 'active',  -- active, revoked, expired
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    last_used_ip INET,

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Indexes
    INDEX idx_user_id (user_id),
    INDEX idx_key_prefix (key_prefix),
    INDEX idx_status (status),
    INDEX idx_expires_at (expires_at)
);
```

### Table: `api_key_usage`

```sql
CREATE TABLE api_key_usage (
    id BIGSERIAL PRIMARY KEY,
    key_id UUID NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,

    -- Request Info
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    operation VARCHAR(50) NOT NULL,  -- upload, download, delete, list, etc.
    path VARCHAR(1000),

    -- Network Info
    ip_address INET NOT NULL,
    user_agent VARCHAR(500),

    -- Stats
    bytes_transferred BIGINT DEFAULT 0,
    status_code INTEGER,
    error_message TEXT,

    -- Partitioning by month
    PARTITION BY RANGE (timestamp)
);

-- Create partitions for current and next 12 months
CREATE TABLE api_key_usage_2026_02 PARTITION OF api_key_usage
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
-- ... more partitions
```

---

## API Key Format

### Structure

```
dk_{environment}_{random_chars}
```

**Examples:**
- `dk_test_abc123xyz789def456ghi012jkl345mno678`  (Test key)
- `dk_live_abc123xyz789def456ghi012jkl345mno678`  (Production key)

**Components:**
- `dk` - Dark Storage prefix
- `test|live` - Environment
- `64 chars` - Cryptographically secure random string

### S3 Credentials Generation

When an API key is created, generate corresponding S3 credentials:

```go
// Pseudo-code
apiKey := generateSecureKey("dk_live_", 64)
s3AccessKey := "DKSA" + generateRandomString(16)  // Dark Storage Access
s3SecretKey := generateSecureKey("", 40)

// Store in database
keyHash := bcrypt.Hash(apiKey)
secretKeyEncrypted := aes256Encrypt(s3SecretKey, masterKey)

db.Insert({
    key_prefix: apiKey[:20],  // "dk_live_abc123xyz789"
    key_hash: keyHash,
    s3_access_key: s3AccessKey,
    s3_secret_key_encrypted: secretKeyEncrypted,
})
```

---

## Permissions System

### Permission Scopes

```json
[
  "storage:read",       // List, download, stat files
  "storage:write",      // Upload files
  "storage:delete",     // Delete files
  "bucket:create",      // Create buckets
  "bucket:delete",      // Delete buckets
  "share:create",       // Create pre-signed URLs
  "admin:*"            // Full access (account owner only)
]
```

### Permission Validation

```go
func (k *APIKey) HasPermission(required string) bool {
    // Admin has all permissions
    if k.HasPermission("admin:*") {
        return true
    }

    // Check exact match
    for _, perm := range k.Permissions {
        if perm == required {
            return true
        }

        // Check wildcard (e.g., "storage:*" grants "storage:read")
        if strings.HasSuffix(perm, ":*") {
            prefix := strings.TrimSuffix(perm, "*")
            if strings.HasPrefix(required, prefix) {
                return true
            }
        }
    }

    return false
}
```

---

## Web Console UI Mockup

### API Keys Page

```
┌─────────────────────────────────────────────────────────────┐
│ Dark Storage Console                      user@example.com ▼│
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Dashboard | Storage | API Keys | Settings | Billing        │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ API Keys                          [+ Generate New Key] │ │
│  │                                                         │ │
│  │ Manage API keys for CLI and programmatic access        │ │
│  ├─────────────────────────────────────────────────────────┤ │
│  │                                                         │ │
│  │ Production Server - Web01                     [Revoke] │ │
│  │ dk_live_abc123...                                      │ │
│  │ Created: Feb 24, 2026 | Last used: 2 hours ago        │ │
│  │ Permissions: Read, Write, Delete                       │ │
│  │ [View Usage] [Download Config]                         │ │
│  │                                                         │ │
│  │ Development Environment                       [Revoke] │ │
│  │ dk_test_def456...                                      │ │
│  │ Created: Feb 20, 2026 | Last used: yesterday           │ │
│  │ Permissions: Read, Write                               │ │
│  │ [View Usage] [Download Config]                         │ │
│  │                                                         │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Generate Key Dialog

```
┌───────────────────────────────────────────┐
│ Generate New API Key                      │
├───────────────────────────────────────────┤
│                                           │
│ Key Name *                                │
│ ┌───────────────────────────────────────┐ │
│ │ Production Server - Web01             │ │
│ └───────────────────────────────────────┘ │
│                                           │
│ Permissions *                             │
│ ☑ Storage Read                            │
│ ☑ Storage Write                           │
│ ☑ Storage Delete                          │
│ ☐ Bucket Create                           │
│ ☐ Bucket Delete                           │
│                                           │
│ Expiration                                │
│ ● 30 days  ○ 90 days  ○ 1 year  ○ Never  │
│                                           │
│ IP Restrictions (optional)                │
│ ┌───────────────────────────────────────┐ │
│ │ 203.0.113.0/24                        │ │
│ │ 198.51.100.42                         │ │
│ └───────────────────────────────────────┘ │
│                                           │
│         [Cancel]  [Generate Key]          │
└───────────────────────────────────────────┘
```

### Key Generated Success

```
┌───────────────────────────────────────────┐
│ ✓ API Key Generated Successfully          │
├───────────────────────────────────────────┤
│                                           │
│ ⚠️  Save this key now - it won't be      │
│    shown again!                           │
│                                           │
│ API Key:                                  │
│ ┌───────────────────────────────────────┐ │
│ │ dk_live_abc123xyz789def456ghi012...   │ │
│ └───────────────────────────────────────┘ │
│                              [Copy] 📋    │
│                                           │
│ [Download Config File (.yaml)]            │
│ [Download Install Script (.sh)]           │
│                                           │
│ Next Steps:                               │
│ 1. Save the API key securely              │
│ 2. Download the config file               │
│ 3. Place in ~/.darkstorage/config.yaml    │
│ 4. Run: darkstorage whoami                │
│                                           │
│                    [Done]                 │
└───────────────────────────────────────────┘
```

---

## Security Considerations

### 1. Key Storage
- Store only bcrypt hash of full API key
- Encrypt S3 secret keys at rest (AES-256)
- Use separate encryption keys per environment
- Rotate master encryption key annually

### 2. Key Transmission
- Always use HTTPS
- Display full key only once at creation
- Support key prefix for identification (dk_live_abc123...)

### 3. Rate Limiting
- 100 requests/minute per API key (configurable)
- Track per key, not per IP
- Return 429 Too Many Requests with Retry-After header

### 4. Audit Logging
- Log all key creation/revocation events
- Log usage stats (without sensitive data)
- Retain logs for 90 days minimum
- Alert on suspicious activity

### 5. IP Restrictions
- Optional but recommended
- Support CIDR notation
- Validate on every request
- Allow empty list = no restrictions

### 6. Key Rotation
- Recommend rotation every 90 days
- Send email reminders before expiration
- Grace period: 7 days after expiration
- Auto-revoke after grace period

---

## Implementation Roadmap

### Phase 1: Backend API (Week 1-2)
- [ ] Database schema and migrations
- [ ] API key generation logic
- [ ] JWT authentication middleware
- [ ] CRUD endpoints for API keys
- [ ] Permission validation system
- [ ] Usage tracking

### Phase 2: Web Console (Week 3-4)
- [ ] React/Next.js setup
- [ ] Authentication flow (OAuth/SSO)
- [ ] API Keys UI page
- [ ] Generate key dialog
- [ ] Config file download
- [ ] Install script generation

### Phase 3: CLI Integration (Week 5)
- [ ] Update `darkstorage login --key` to accept new format
- [ ] Add config file auto-loading
- [ ] Validate API key format
- [ ] Error handling for expired/revoked keys

### Phase 4: Testing & Polish (Week 6)
- [ ] E2E testing
- [ ] Security audit
- [ ] Documentation
- [ ] Production deployment

---

## Related Documents

- `SSO_INTEGRATION.md` - Single sign-on across platforms
- `HOSTNAME_ARCHITECTURE.md` - Subdomain structure
- `MASTER_VISION.md` - Overall platform vision
- `cmd/login.go` - Existing CLI login implementation

---

*Last Updated: 2026-02-24*
*Status: Design Document - Ready for Implementation*
