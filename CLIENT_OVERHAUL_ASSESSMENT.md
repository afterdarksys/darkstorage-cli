# Dark Storage CLI - Client Overhaul Assessment

**Date**: 2026-02-24
**Status**: Planning Phase

## Executive Summary

The Dark Storage CLI requires a major overhaul to integrate with:
1. **MinIO S3-Compatible Storage Backend**
2. **Encryption Key Management System** (3 backup + 1 active key)
3. **Real Backend API Integration** (currently all commands are mocked)

## Current State Analysis

### What's Implemented ✅

1. **Full GUI + Daemon Architecture**
   - Fyne-based desktop application (`cmd/gui/`)
   - Background sync daemon (`cmd/daemon/`)
   - SQLite database for sync state tracking
   - IPC communication (Unix sockets/Named pipes)
   - File system watcher (fsnotify)
   - Sync engine with conflict resolution
   - Location: `/internal/db/`, `/internal/sync/`, `/internal/ipc/`

2. **CLI Command Structure**
   - All commands defined: `ls, put, get, rm, cp, mv, search, hashsearch, metadata, attrs`
   - User management: `login, logout, whoami`
   - Advanced features: `groups, perms, shares, scan, trash, audit`
   - Location: `/cmd/*.go`

3. **Dependencies Installed**
   - Fyne v2.7.2 (GUI framework)
   - SQLite3 v1.14.34 (local database)
   - fsnotify v1.9.0 (file watching)
   - Cobra v1.8.0 (CLI framework)
   - Viper v1.18.2 (configuration)

### What's Missing/Broken ❌

1. **API Client Implementation** - CRITICAL GAP
   - File: `/internal/api/client.go` and `/internal/api/storage.go`
   - Current status: Stub implementations that only print to console
   - Missing:
     - Actual HTTP requests to backend
     - MinIO S3 SDK integration
     - Authentication header handling
     - Error handling and retries
     - Progress tracking
     - Bandwidth limiting

2. **All CLI Commands Are Mocked**
   - File: `/cmd/storage.go` (lines 31-313)
   - Every command (ls, put, get, rm, etc.) just prints fake data
   - No actual connection to storage backend
   - No authentication flow

3. **No MinIO Integration**
   - Missing: MinIO Go SDK (`github.com/minio/minio-go/v7`)
   - Need: S3-compatible client configuration
   - Need: Bucket policy management
   - Need: Object versioning support

4. **No Encryption System**
   - Missing: Key management infrastructure
   - Missing: Key rotation mechanism (3 backup + 1 active)
   - Missing: Encryption/decryption layer
   - Missing: Secure key storage (keychain/vault integration)

5. **Configuration Incomplete**
   - No MinIO endpoint configuration
   - No encryption key storage location
   - No key rotation policy settings

## New Features Required

### 1. MinIO Storage Backend Integration

**Requirements:**
- Connect to MinIO-compatible S3 storage
- Support multiple endpoints (production, staging, local)
- Bucket operations: create, list, delete
- Object operations: put, get, delete, copy, move
- Multipart upload for large files
- Streaming downloads with progress
- Content-type detection
- Metadata management
- Object versioning
- Server-side encryption support

**Implementation Needs:**
```go
// Add to go.mod
github.com/minio/minio-go/v7 v7.0.66

// New files needed
internal/storage/minio_client.go  // MinIO SDK wrapper
internal/storage/buckets.go       // Bucket operations
internal/storage/objects.go       // Object operations
internal/storage/multipart.go     // Multipart upload/download
```

**Configuration Schema:**
```yaml
storage:
  backend: minio  # or s3, cloudflare-r2, etc.

  # Primary S3-compatible endpoint (user-friendly hostname)
  endpoint: s3.darkstorage.io

  # HSM (Hardware Security Module) support
  # Users wanting HSM should use: s3.darkstorage.io
  # This hostname provides easy access to HSM-backed encryption
  hsm_enabled: false  # Set to true when using HSM backend

  # Credentials
  access_key: ${DARKSTORAGE_ACCESS_KEY}
  secret_key: ${DARKSTORAGE_SECRET_KEY}

  # Connection settings
  use_ssl: true
  region: us-east-1
  bucket_prefix: "user-"

  # Alternative endpoints (for advanced users)
  endpoints:
    production: s3.darkstorage.io
    staging: s3-staging.darkstorage.io
    local: localhost:9000
    hsm: s3.darkstorage.io  # HSM-backed storage endpoint

  # Performance
  multipart_threshold: 64MB  # Files larger than this use multipart
  multipart_chunk_size: 5MB
  concurrent_uploads: 4
  concurrent_downloads: 4
```

### 2. Encryption Key Management (3+1 System)

**Requirements:**
- Store 1 active encryption key + 3 backup keys
- Automatic key rotation policy
- Secure key storage (OS keychain integration)
- Key versioning and migration
- Per-file encryption with key reference
- Client-side encryption before upload
- Decryption on download

**Architecture:**
```
Key Storage Hierarchy:
├── Active Key (ID: current)
│   └── Used for all new encryptions
├── Backup Key 1 (ID: backup-1)
│   └── Previous active, kept for 90 days
├── Backup Key 2 (ID: backup-2)
│   └── Older backup, kept for 60 days
└── Backup Key 3 (ID: backup-3)
    └── Oldest backup, kept for 30 days
```

**Implementation Needs:**
```go
// Add to go.mod
github.com/99designs/keyring v1.2.2     // OS keychain integration
golang.org/x/crypto v0.17.0             // AES-GCM encryption

// New files needed
internal/crypto/keys.go          // Key management
internal/crypto/keystore.go      // Secure key storage
internal/crypto/rotation.go      // Key rotation logic
internal/crypto/encrypt.go       // Encryption operations
internal/crypto/decrypt.go       // Decryption operations
internal/crypto/metadata.go      // Track which key encrypted each file
```

**Key Metadata Schema (stored with each file):**
```json
{
  "encryption": {
    "enabled": true,
    "algorithm": "AES-256-GCM",
    "key_id": "active",
    "key_version": 3,
    "encrypted_at": "2026-02-24T10:30:00Z",
    "nonce": "base64-encoded-nonce",
    "tag": "base64-encoded-auth-tag"
  }
}
```

**Configuration Schema:**
```yaml
encryption:
  enabled: true
  algorithm: aes-256-gcm

  # Key storage
  keystore_backend: keychain  # keychain, vault, file
  keystore_path: ~/.darkstorage/keys/

  # Key rotation policy
  rotation_enabled: true
  rotation_interval: 90d  # Rotate active key every 90 days
  backup_retention:
    backup_1: 90d
    backup_2: 60d
    backup_3: 30d

  # Security
  require_passphrase: false  # Require password to access keys
  lock_timeout: 15m          # Auto-lock keys after inactivity
```

### 3. Backend API Integration

**Current Problem:**
All commands in `/cmd/storage.go` are hardcoded mocks. They need to actually call the API client.

**Required Changes:**

**File: `/internal/api/client.go`**
```go
// Replace current stub with real implementation
type Client struct {
    minioClient *minio.Client
    endpoint    string
    accessKey   string
    secretKey   string
    useSSL      bool
    region      string
    httpClient  *http.Client
    encryptor   *crypto.Encryptor  // NEW
}

func NewClient(config *Config) (*Client, error) {
    // Initialize MinIO client
    minioClient, err := minio.New(config.Endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
        Secure: config.UseSSL,
        Region: config.Region,
    })

    // Initialize encryption
    encryptor := crypto.NewEncryptor(config.EncryptionConfig)

    return &Client{
        minioClient: minioClient,
        encryptor:   encryptor,
        // ...
    }
}
```

**File: `/internal/api/storage.go`**
```go
// Replace stub implementations with real MinIO calls

func (c *Client) UploadFile(localPath, remotePath string, progress UploadProgress) error {
    // 1. Read local file
    // 2. Encrypt if encryption enabled
    // 3. Upload to MinIO with progress tracking
    // 4. Store encryption metadata

    bucket, object := parsePath(remotePath)

    // Encrypt
    if c.encryptor.Enabled() {
        encryptedData, metadata, err := c.encryptor.Encrypt(fileData)
        // Store metadata with object
    }

    // Upload with progress
    info, err := c.minioClient.PutObject(ctx, bucket, object, reader, size, minio.PutObjectOptions{
        Progress: progress,
        // Include encryption metadata
    })

    return err
}

func (c *Client) DownloadFile(remotePath, localPath string, progress DownloadProgress) error {
    // 1. Download from MinIO
    // 2. Decrypt if encrypted (check metadata)
    // 3. Write to local file

    bucket, object := parsePath(remotePath)

    // Download
    obj, err := c.minioClient.GetObject(ctx, bucket, object, minio.GetObjectOptions{})

    // Check encryption metadata
    metadata := obj.Metadata()
    if metadata["X-Amz-Meta-Encryption"] == "true" {
        keyID := metadata["X-Amz-Meta-Key-Id"]
        decryptedData, err := c.encryptor.Decrypt(encryptedData, keyID)
    }

    return err
}
```

## Implementation Phases

### Phase 1: MinIO Integration (Week 1-2)
**Priority: CRITICAL**

- [ ] Add MinIO SDK dependency
- [ ] Implement `/internal/storage/minio_client.go`
  - Connection management
  - Credential handling
  - Region/endpoint configuration
- [ ] Implement bucket operations
  - Create bucket
  - List buckets
  - Delete bucket
  - Bucket policies
- [ ] Implement object operations
  - PutObject with streaming
  - GetObject with streaming
  - DeleteObject
  - CopyObject
  - StatObject (metadata)
- [ ] Add progress tracking
- [ ] Add bandwidth limiting
- [ ] Update configuration schema

**Deliverable:** Working MinIO client that can upload/download files

### Phase 2: Replace Mocked Commands (Week 2-3)
**Priority: CRITICAL**

- [ ] Update `/cmd/storage.go` to use real API client
  - Replace `lsCmd` mock with actual MinIO list
  - Replace `putCmd` mock with actual upload
  - Replace `getCmd` mock with actual download
  - Replace `rmCmd` mock with actual delete
  - Replace `cpCmd` and `mvCmd` with actual operations
- [ ] Update `/cmd/login.go` for real authentication
- [ ] Update `/cmd/whoami.go` for real user info
- [ ] Add error handling and retries
- [ ] Add progress bars for uploads/downloads

**Deliverable:** All CLI commands work with real backend

### Phase 3: Encryption Layer (Week 3-5)
**Priority: HIGH**

- [ ] Design key management architecture
- [ ] Implement `/internal/crypto/` package
  - Key generation (AES-256)
  - Secure key storage (keychain integration)
  - Key rotation logic
  - Encryption (AES-GCM)
  - Decryption with key versioning
- [ ] Implement 3+1 key system
  - Active key tracking
  - 3 backup key slots
  - Rotation scheduler
  - Migration tool for re-encrypting old files
- [ ] Add encryption metadata to uploads
- [ ] Add decryption to downloads
- [ ] Create key management commands
  - `darkstorage keys list`
  - `darkstorage keys rotate`
  - `darkstorage keys backup`
  - `darkstorage keys import/export`

**Deliverable:** Transparent client-side encryption for all operations

### Phase 4: Integration with Daemon (Week 5-6)
**Priority: MEDIUM**

The daemon/sync engine needs the updated API client:

- [ ] Update `/internal/sync/engine.go` to use real MinIO client
- [ ] Add encryption support to sync operations
- [ ] Handle key rotation during sync
  - Detect files encrypted with old keys
  - Re-encrypt with active key during sync
- [ ] Update conflict resolution for encrypted files
- [ ] Add encryption status to GUI
  - Show which files are encrypted
  - Show active key info
  - Show key rotation schedule

**Deliverable:** Daemon can sync encrypted files to MinIO

### Phase 5: Advanced Features (Week 6-8)
**Priority: MEDIUM**

- [ ] Multipart upload for large files (>64MB)
- [ ] Resume interrupted uploads/downloads
- [ ] Object versioning support
- [ ] Server-side encryption option (in addition to client-side)
- [ ] Compression before encryption (optional)
- [ ] Deduplication using content hashing
- [ ] Implement remaining commands
  - `/cmd/groups.go` - Group management via API
  - `/cmd/perms.go` - Permissions via bucket policies
  - `/cmd/shares.go` - Presigned URLs for sharing
  - `/cmd/scan.go` - Malware scanning integration
  - `/cmd/trash.go` - Soft delete with versioning
  - `/cmd/audit.go` - Audit log retrieval

**Deliverable:** Full-featured client with all advanced capabilities

### Phase 6: Security Hardening (Week 8-9)
**Priority: HIGH**

- [ ] Security audit of encryption implementation
- [ ] Key zeroization (clear keys from memory)
- [ ] Secure credential storage (never plaintext)
- [ ] Input validation and sanitization
- [ ] Rate limiting and retry logic
- [ ] TLS certificate validation
- [ ] Add security documentation
- [ ] Penetration testing

**Deliverable:** Production-ready security posture

## Critical Decisions Needed

### 1. Encryption Scope
**Decision Required:** What gets encrypted?
- [ ] Option A: Everything (all files, all metadata)
- [ ] Option B: User choice per file/folder
- [ ] Option C: Hybrid (sensitive folders auto-encrypted)

**Recommendation:** Option B - User choice with smart defaults

### 2. Key Storage Backend
**Decision Required:** Where to store encryption keys?
- [ ] Option A: OS Keychain (macOS/Windows/Linux)
- [ ] Option B: HashiCorp Vault
- [ ] Option C: Encrypted file with master password
- [ ] Option D: HSM (Hardware Security Module) via s3.darkstorage.io
- [ ] Option E: Multiple backends (user choice)

**Recommendation:** Option E - Default to OS keychain, support Vault for enterprise, HSM for high-security users

**HSM Integration Notes:**
- Users wanting HSM should connect to `s3.darkstorage.io`
- This endpoint provides FIPS 140-2 compliant hardware-backed encryption
- Keys never leave the HSM device
- Transparent to client - just set `hsm_enabled: true` in config
- Requires HSM-enabled account tier

### 3. Key Rotation Trigger
**Decision Required:** How to trigger key rotation?
- [ ] Option A: Automatic (time-based, e.g., every 90 days)
- [ ] Option B: Manual (user command)
- [ ] Option C: Hybrid (auto + manual option)

**Recommendation:** Option C - Auto-rotation with manual override

### 4. Old File Migration
**Decision Required:** What to do with files encrypted with old keys?
- [ ] Option A: Re-encrypt all during rotation (slow, secure)
- [ ] Option B: Re-encrypt on access (lazy, faster)
- [ ] Option C: Keep old files as-is until manual migration
- [ ] Option D: Hybrid (re-encrypt accessed files, schedule batch jobs)

**Recommendation:** Option D - Lazy re-encryption with background jobs

### 5. Performance vs Security
**Decision Required:** Encryption performance impact
- [ ] Option A: Always encrypt (slower, more secure)
- [ ] Option B: Encrypt on upload, cache plaintext locally (faster, less secure)
- [ ] Option C: Stream encryption (balanced)

**Recommendation:** Option C - Stream encryption with configurable local cache

## File Structure After Overhaul

```
darkstorage-cli/
├── cmd/
│   ├── storage.go           # UPDATED: Real backend calls
│   ├── login.go             # UPDATED: Real auth
│   ├── keys.go              # NEW: Key management commands
│   └── ...
│
├── internal/
│   ├── api/
│   │   ├── client.go        # UPDATED: Real HTTP + MinIO
│   │   └── storage.go       # UPDATED: Real operations
│   │
│   ├── storage/             # NEW PACKAGE
│   │   ├── minio_client.go  # MinIO SDK wrapper
│   │   ├── buckets.go       # Bucket operations
│   │   ├── objects.go       # Object operations
│   │   └── multipart.go     # Multipart handling
│   │
│   ├── crypto/              # NEW PACKAGE
│   │   ├── keys.go          # Key management
│   │   ├── keystore.go      # Secure storage
│   │   ├── rotation.go      # Rotation logic
│   │   ├── encrypt.go       # Encryption
│   │   ├── decrypt.go       # Decryption
│   │   └── metadata.go      # Encryption metadata
│   │
│   ├── sync/
│   │   ├── engine.go        # UPDATED: Use real client + encryption
│   │   └── ...
│   │
│   └── ...
│
├── go.mod                   # UPDATED: Add MinIO, crypto deps
└── ...
```

## Dependencies to Add

```bash
# MinIO SDK
go get github.com/minio/minio-go/v7

# Encryption
go get golang.org/x/crypto

# Key storage
go get github.com/99designs/keyring

# Additional utilities
go get github.com/schollz/progressbar/v3  # Better progress bars
go get github.com/klauspost/compress      # Compression (optional)
```

## Testing Strategy

### Unit Tests
- [ ] MinIO client operations (with mock server)
- [ ] Encryption/decryption functions
- [ ] Key rotation logic
- [ ] Key storage operations

### Integration Tests
- [ ] End-to-end upload/download with encryption
- [ ] Key rotation with file migration
- [ ] Multi-key decryption (files with different keys)
- [ ] Daemon sync with encrypted files

### Performance Tests
- [ ] Large file uploads (multipart)
- [ ] Concurrent operations
- [ ] Encryption overhead measurement
- [ ] Memory usage with encryption

### Security Tests
- [ ] Key zeroization verification
- [ ] Credential leak detection
- [ ] Encryption strength validation
- [ ] Key rotation edge cases

## Migration Path for Existing Users

If there are existing users with files in the system:

1. **Backward Compatibility:**
   - Detect unencrypted files
   - Support mixed encrypted/unencrypted buckets
   - Gradual migration option

2. **Migration Tool:**
   ```bash
   darkstorage migrate encrypt --bucket my-bucket --recursive
   ```

3. **Migration Strategy:**
   - Default: Opt-in encryption (don't break existing workflows)
   - Provide clear migration guide
   - Support rollback if needed

## Success Metrics

- [ ] All CLI commands work with real MinIO backend
- [ ] Encryption adds <10% overhead to upload/download speed
- [ ] Key rotation completes in <1 minute for 1000 files
- [ ] Zero plaintext keys in logs or config files
- [ ] All unit tests pass with >80% coverage
- [ ] Integration tests pass on macOS, Linux, Windows
- [ ] Security audit finds no critical vulnerabilities

## Timeline Estimate

- **Phase 1:** 2 weeks (MinIO integration)
- **Phase 2:** 1 week (Replace mocks)
- **Phase 3:** 2 weeks (Encryption layer)
- **Phase 4:** 1 week (Daemon integration)
- **Phase 5:** 2 weeks (Advanced features)
- **Phase 6:** 1 week (Security hardening)

**Total: 9 weeks** (can be parallelized with multiple developers)

## Risk Assessment

### High Risk
- **Encryption bugs** → Data loss or inaccessible files
  - Mitigation: Extensive testing, keep backups of keys
- **Key rotation failures** → Files locked with lost keys
  - Mitigation: Always keep 3 backup keys, export mechanism

### Medium Risk
- **Performance degradation** → Slow uploads/downloads
  - Mitigation: Stream processing, compression, caching
- **MinIO compatibility** → Issues with different S3 providers
  - Mitigation: Test with multiple backends (MinIO, AWS S3, R2)

### Low Risk
- **GUI integration** → Daemon might need updates
  - Mitigation: Well-defined interfaces already in place

## Next Steps

1. **Review this document** with team/stakeholders
2. **Make critical decisions** (encryption scope, key storage, etc.)
3. **Set up MinIO development instance** for testing
4. **Create feature branch** for development
5. **Start Phase 1** (MinIO integration)

## HSM Premium Feature (Oracle eHSM Integration)

### Overview
HSM (Hardware Security Module) support is a **premium tier feature** for enterprise customers requiring:
- FIPS 140-2 Level 3+ compliance
- Hardware-backed key storage
- Enhanced regulatory compliance (HIPAA, PCI-DSS, SOC 2)
- Tamper-proof key operations

### Architecture

**Standard Tier (Free/Basic):**
```
Client → s3.darkstorage.io → MinIO → Standard Encryption
         (Client-side encryption using OS keychain)
```

**Premium Tier (HSM-backed):**
```
Client → s3.darkstorage.io → MinIO → Oracle eHSM
         (Server-side encryption with HSM-backed keys)
         (Client can still do client-side encryption too - double encryption)
```

### Implementation Requirements

**Backend (Server-side):**
- [ ] Oracle Cloud Infrastructure (OCI) account with eHSM access
- [ ] Configure MinIO to use Oracle eHSM for SSE (Server-Side Encryption)
- [ ] Set up KMS (Key Management Service) integration
- [ ] Configure HSM key rotation policies
- [ ] Implement access controls and audit logging

**Client-side (CLI):**
- [ ] Add HSM tier detection (via API endpoint)
- [ ] Update configuration to support HSM mode
- [ ] Add HSM status indicators in CLI/GUI
- [ ] Support both client-side + server-side encryption (layered)

### Configuration Schema for HSM

```yaml
# User-facing configuration
storage:
  endpoint: s3.darkstorage.io
  hsm_enabled: true  # Requires premium tier subscription

  # HSM settings (auto-configured from account tier)
  hsm:
    provider: oracle  # Fixed: using Oracle eHSM
    tier: premium     # Detected from account
    double_encryption: true  # Client-side + server-side (recommended)
    compliance_mode: fips140-2-level3

encryption:
  # Client-side encryption (always available)
  client_side:
    enabled: true
    algorithm: aes-256-gcm

  # Server-side encryption (HSM-backed, premium only)
  server_side:
    enabled: true  # Only works if hsm_enabled: true
    algorithm: aes-256-gcm
    key_management: hsm  # Keys managed by Oracle eHSM
```

### Client Behavior by Tier

**Standard Tier:**
```bash
$ darkstorage put file.txt my-bucket/
[Encrypting with client-side key...]
✓ Uploaded file.txt (encrypted with AES-256-GCM)
Encryption: Client-side only
```

**Premium Tier (HSM enabled):**
```bash
$ darkstorage put file.txt my-bucket/
[Encrypting with client-side key...]
[Server will apply HSM encryption...]
✓ Uploaded file.txt (double-encrypted)
Encryption: Client-side (AES-256) + Server-side HSM (Oracle eHSM)
Compliance: FIPS 140-2 Level 3
```

### Premium Feature Indicators

**CLI Status:**
```bash
$ darkstorage whoami
User: ryan@darkstorage.io
Tier: Premium (HSM-enabled)
Storage: s3.darkstorage.io
HSM Provider: Oracle eHSM
Compliance: FIPS 140-2 Level 3
Encryption: Double (Client + Server HSM)
```

**GUI Indicators:**
- Show "HSM Protected" badge on files
- Display compliance certifications
- Show encryption layers (client + server)
- Premium tier badge in top bar

### Pricing Tiers (Suggested)

**Standard Tier:**
- Client-side encryption only
- OS keychain key storage
- Good for: Personal use, small teams
- Cost: Free or low monthly fee

**Premium Tier (HSM):**
- Client-side + Server-side HSM encryption
- Oracle eHSM backed keys
- FIPS 140-2 Level 3 compliance
- Enhanced audit logging
- Priority support
- Good for: Enterprise, healthcare, finance, government
- Cost: Premium monthly fee (e.g., $99-$499/month)

### Backend Implementation (Oracle eHSM)

**Required Oracle Cloud Services:**
1. **OCI Vault** - Key management service
2. **HSM Cluster** - Hardware security module cluster
3. **KMS Integration** - Key Management Service API
4. **Audit Service** - Compliance logging

**MinIO Configuration with Oracle eHSM:**
```bash
# Set MinIO to use OCI KMS with HSM
mc admin kms key create myminio premium-hsm-key \
  --insecure-kms \
  --kms-type=oci \
  --kms-endpoint=https://kms.us-ashburn-1.oraclecloud.com \
  --hsm-enabled

# Configure auto-encryption for premium buckets
mc encrypt set sse-kms premium-hsm-key myminio/premium-bucket-*
```

**Environment Variables (Backend):**
```bash
# Oracle Cloud credentials
export OCI_TENANCY_OCID="ocid1.tenancy.oc1..."
export OCI_USER_OCID="ocid1.user.oc1..."
export OCI_FINGERPRINT="aa:bb:cc:..."
export OCI_PRIVATE_KEY_PATH="/path/to/oci-key.pem"

# HSM configuration
export MINIO_KMS_KES_ENDPOINT="https://kms.us-ashburn-1.oraclecloud.com"
export MINIO_KMS_KES_KEY_NAME="premium-hsm-master-key"
export MINIO_KMS_HSM_ENABLED="true"
```

### Client Detection Flow

```go
// internal/storage/minio_client.go

type AccountTier struct {
    Tier         string `json:"tier"`          // "standard" or "premium"
    HSMEnabled   bool   `json:"hsm_enabled"`
    HSMProvider  string `json:"hsm_provider"`  // "oracle"
    Compliance   string `json:"compliance"`    // "fips140-2-level3"
}

func (c *Client) DetectAccountTier() (*AccountTier, error) {
    // Call API endpoint to get account tier
    resp, err := c.httpClient.Get(c.endpoint + "/api/v1/account/tier")

    var tier AccountTier
    json.NewDecoder(resp.Body).Decode(&tier)

    return &tier, nil
}

func (c *Client) UploadFile(localPath, remotePath string, opts *UploadOptions) error {
    // Check account tier
    tier, _ := c.DetectAccountTier()

    // Always do client-side encryption
    encryptedData, _ := c.encryptor.Encrypt(fileData)

    // Set server-side encryption headers if HSM enabled
    headers := make(map[string]string)
    if tier.HSMEnabled {
        headers["X-Amz-Server-Side-Encryption"] = "aws:kms"
        headers["X-Amz-Server-Side-Encryption-Aws-Kms-Key-Id"] = "premium-hsm-key"
    }

    // Upload with appropriate encryption
    return c.minioClient.PutObject(ctx, bucket, object, encryptedData, headers)
}
```

### Migration Path: Standard → Premium

**When user upgrades to Premium tier:**

1. **Automatic detection:**
   ```bash
   $ darkstorage sync
   [Detected account upgrade to Premium tier]
   [HSM encryption now available]
   [Re-encrypting existing files with HSM? (y/n)]
   ```

2. **Gradual migration:**
   - New uploads: Automatic double encryption (client + HSM)
   - Existing files: Re-encrypt on next access (lazy)
   - Batch migration: `darkstorage migrate hsm --bucket my-bucket`

3. **Downgrade protection:**
   - Files encrypted with HSM remain accessible
   - Automatic fallback to client-side only on tier downgrade
   - Warning when premium features will be lost

### Compliance & Audit Features

**Premium tier includes:**
- [ ] Detailed encryption audit logs
- [ ] HSM key access logs (from Oracle)
- [ ] Compliance reports (HIPAA, PCI-DSS, SOC 2)
- [ ] Automated compliance checking
- [ ] Tamper detection and alerts

**CLI Command for Compliance:**
```bash
$ darkstorage compliance report --month 2026-02
Generating compliance report for February 2026...

Encryption Status:
  Total Files: 10,234
  HSM Encrypted: 10,234 (100%)
  Client Encrypted: 10,234 (100%)
  Double Encrypted: 10,234 (100%)

HSM Operations (Oracle eHSM):
  Key Rotations: 1
  Encryption Operations: 15,678
  Decryption Operations: 8,432
  Failed Operations: 0
  Tamper Attempts: 0

Compliance:
  FIPS 140-2 Level 3: ✓ Compliant
  HIPAA: ✓ Compliant
  PCI-DSS: ✓ Compliant
  SOC 2: ✓ Compliant

Report saved to: compliance-2026-02.pdf
```

### Development Phases for HSM Support

**Phase 1: Infrastructure Setup (Oracle Cloud)**
- [ ] Set up OCI account with eHSM access
- [ ] Configure HSM cluster and KMS
- [ ] Integrate MinIO with Oracle KMS
- [ ] Test server-side encryption with HSM

**Phase 2: Client Detection & Config**
- [ ] Add tier detection API endpoint
- [ ] Update client configuration schema
- [ ] Implement HSM status indicators
- [ ] Add double encryption support

**Phase 3: User Experience**
- [ ] Premium tier badges in GUI
- [ ] Compliance status displays
- [ ] Migration tools (standard → premium)
- [ ] Documentation for premium features

**Phase 4: Compliance & Reporting**
- [ ] Audit log integration
- [ ] Compliance report generation
- [ ] Automated compliance checks
- [ ] Security certifications

### Cost Considerations

**Oracle eHSM Pricing (Estimated):**
- HSM Cluster: ~$1,500-$3,000/month
- KMS Operations: ~$1 per 10,000 operations
- Storage: Standard S3 pricing

**Revenue Model:**
- Charge premium tier at $99-$499/month per user
- Break-even at ~20-50 premium users
- Higher margins with volume

### Questions for Discussion

1. Do we need server-side encryption in addition to client-side?
   - **Answer: YES for Premium/HSM tier** - Double encryption provides defense in depth
2. What's the expected file size range? (affects multipart strategy)
3. Are there compliance requirements (HIPAA, GDPR, etc.)?
   - **Answer: YES for Premium tier** - HSM provides FIPS 140-2, HIPAA, PCI-DSS compliance
4. Should we support multiple encryption algorithms or just AES-256-GCM?
5. What's the disaster recovery plan if all keys are lost?
   - **Answer: HSM keys in Oracle are backed up and recoverable** - Enhanced DR for premium tier
6. What tier pricing makes sense for HSM feature?
   - **Recommendation: $99-$499/month** - Depends on target market and Oracle costs
