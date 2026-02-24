# Dark Storage - Complete Revamp Plan
## Production-Ready S3 Competitor

**Date**: 2026-02-24
**Goal**: Launch-ready competitor to AWS S3, Backblaze B2, Cloudflare R2
**Timeline**: ASAP - Full production release

---

## Competitive Landscape

### Direct Competitors:
- **AWS S3** - Industry standard, expensive, complex pricing
- **Backblaze B2** - Cost-effective, simple pricing
- **Cloudflare R2** - Zero egress fees, S3-compatible
- **Wasabi** - Flat pricing, no egress fees
- **DigitalOcean Spaces** - Developer-friendly

### Our Differentiators:
1. ✅ **End-to-end encryption** (client-side + optional HSM)
2. ✅ **3+1 key rotation** (security without complexity)
3. ✅ **Desktop sync daemon** (built-in, not third-party)
4. ✅ **Beautiful GUI** (Fyne-based, native feel)
5. ✅ **Storage classes** (AWS-compatible tiers)
6. ✅ **Pre-signed URLs** (easy sharing)
7. ✅ **ZIP downloads** (entire directories)
8. ✅ **HSM tier** (compliance for enterprise)
9. ✅ **Simple pricing** (no hidden fees)
10. ✅ **Privacy-first** (client-side encryption by default)
11. ✅ **Web3 Integration** (Storj, IPFS - opt-in decentralized storage) 🌐

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     CLIENT APPLICATIONS                      │
├──────────────────┬──────────────────┬──────────────────────┤
│   CLI Tool       │   Desktop GUI    │   Sync Daemon        │
│   (darkstorage)  │   (Fyne)         │   (Background)       │
└────────┬─────────┴────────┬─────────┴──────────┬───────────┘
         │                  │                    │
         └──────────────────┼────────────────────┘
                            │
         ┌──────────────────┴────────────────────┐
         │      Internal Storage Client          │
         │  - MinIO SDK                          │
         │  - Encryption Layer (3+1 keys)        │
         │  - Auth Manager                       │
         │  - Progress Tracking                  │
         └──────────────────┬────────────────────┘
                            │
         ┌──────────────────┴────────────────────┐
         │      s3.darkstorage.io                │
         ├───────────────────────────────────────┤
         │  MinIO Cluster (S3-compatible)        │
         │  - Standard Tier: Basic encryption    │
         │  - Premium Tier: Oracle eHSM          │
         └──────────────────┬────────────────────┘
                            │
         ┌──────────────────┴────────────────────┐
         │      Storage Backend                  │
         │  - Object Storage (MinIO)             │
         │  - Metadata Database (PostgreSQL)     │
         │  - Oracle eHSM (Premium)              │
         │  - CDN (for fast downloads)           │
         └───────────────────────────────────────┘
```

---

## Feature Set (Complete)

### Tier 1: Core Features (MVP - Required for Launch)

**Storage Operations:**
- [x] Create/delete buckets
- [x] Upload files (single & multipart)
- [x] Download files (with resume)
- [x] List files/folders
- [x] Delete files
- [x] Copy files
- [x] Move/rename files
- [x] Get file metadata
- [x] Set file metadata

**Authentication & Security:**
- [x] User login/logout
- [x] API key management
- [x] Token refresh
- [x] Client-side encryption (AES-256-GCM)
- [x] 3+1 key rotation system
- [x] Secure key storage (OS keychain)

**CLI Commands:**
```bash
darkstorage login
darkstorage logout
darkstorage whoami
darkstorage ls [bucket/path]
darkstorage put <local> <remote>
darkstorage get <remote> [local]
darkstorage rm <path>
darkstorage cp <source> <dest>
darkstorage mv <source> <dest>
darkstorage cat <path>
```

### Tier 2: Advanced Features (Launch+)

**Storage Classes:**
- [x] STANDARD (hot storage, instant access)
- [x] STANDARD_IA (infrequent access, 30-day min)
- [x] INTELLIGENT_TIERING (auto-optimize)
- [x] GLACIER (archival, minutes to retrieve)
- [x] DEEP_ARCHIVE (cold storage, hours to retrieve)

**Sharing & Collaboration:**
- [x] Pre-signed URLs (time-limited links)
- [x] Password-protected shares
- [x] Download limits
- [x] Share analytics (who downloaded, when)

**Advanced Operations:**
- [x] ZIP directory downloads
- [x] Batch operations
- [x] Recursive uploads/downloads
- [x] Search by name/hash/metadata
- [x] File versioning
- [x] Lifecycle policies

**CLI Commands:**
```bash
darkstorage share <path> --expires 7d --password secret
darkstorage get <folder> --zip ./archive.zip
darkstorage storage-class set <path> GLACIER
darkstorage search "*.jpg"
darkstorage hashsearch <sha256>
darkstorage versions list <path>
darkstorage lifecycle set <bucket> --transition-to GLACIER --days 90
```

### Tier 3: Premium Features (Revenue Drivers)

**HSM Tier (Oracle eHSM):**
- [x] FIPS 140-2 Level 3 compliance
- [x] Hardware-backed encryption
- [x] Tamper-proof key operations
- [x] Compliance reporting (HIPAA, PCI-DSS, SOC 2)
- [x] Audit logging

**Enterprise Features:**
- [x] User groups & permissions
- [x] Role-based access control (RBAC)
- [x] Organization management
- [x] SSO integration (SAML, OIDC)
- [x] Audit logs & compliance reports
- [x] SLA guarantees

**Desktop Sync:**
- [x] Automatic folder sync
- [x] Conflict resolution
- [x] Selective sync
- [x] Bandwidth throttling
- [x] Pause/resume sync
- [x] Offline mode

**CLI Commands:**
```bash
darkstorage groups create <name>
darkstorage groups add-user <group> <user>
darkstorage perms set <path> --group <name> --access read
darkstorage audit logs --start 2024-01-01 --end 2024-12-31
darkstorage compliance report --type HIPAA
darkstorage sync add <local-folder> <remote-path>
darkstorage sync status
```

---

## Pricing Strategy

### Free Tier (Teaser)
- 10 GB storage
- 10 GB bandwidth/month
- Client-side encryption
- Basic CLI/GUI
- **Price**: FREE

### Standard Tier (Individual/Small Business)
- 100 GB - 10 TB storage
- Unlimited bandwidth (fair use)
- Client-side encryption
- All storage classes
- Pre-signed URLs
- Desktop sync
- **Price**: $5-50/month (based on storage)

### Premium Tier (Enterprise)
- Unlimited storage
- Unlimited bandwidth
- Client-side + HSM encryption
- FIPS 140-2 compliance
- HIPAA/PCI-DSS/SOC 2
- Priority support
- SLA guarantees
- Audit logging
- **Price**: $99-999/month (based on features + storage)

---

## Technology Stack

### Client Side (This Repo)
```
darkstorage-cli/
├── cmd/
│   ├── cli/                    # CLI application
│   │   ├── main.go
│   │   ├── login.go
│   │   ├── storage.go         # ls, put, get, rm, cp, mv
│   │   ├── share.go           # Pre-signed URLs
│   │   ├── storage_class.go   # Storage tier management
│   │   ├── encryption.go      # Key management
│   │   ├── groups.go          # User groups
│   │   ├── perms.go           # Permissions
│   │   ├── audit.go           # Audit logs
│   │   └── sync.go            # Sync management
│   │
│   ├── gui/                   # Desktop GUI (Fyne)
│   │   ├── main.go
│   │   ├── dashboard.go
│   │   ├── browser.go
│   │   ├── sync.go
│   │   ├── settings.go
│   │   └── compliance.go
│   │
│   └── daemon/                # Background sync daemon
│       ├── main.go
│       ├── watcher.go
│       ├── syncer.go
│       └── ipc.go
│
├── internal/
│   ├── storage/               # ⭐ NEW: Core storage client
│   │   ├── client.go          # MinIO wrapper
│   │   ├── auth.go            # Authentication
│   │   ├── buckets.go         # Bucket operations
│   │   ├── objects.go         # Object operations
│   │   ├── multipart.go       # Large file handling
│   │   ├── shares.go          # Pre-signed URLs
│   │   ├── storage_class.go   # Tier management
│   │   └── versioning.go      # File versions
│   │
│   ├── crypto/                # ⭐ NEW: Encryption layer
│   │   ├── keys.go            # Key management (3+1)
│   │   ├── keystore.go        # OS keychain integration
│   │   ├── rotation.go        # Key rotation
│   │   ├── encrypt.go         # AES-256-GCM encryption
│   │   ├── decrypt.go         # Decryption
│   │   └── metadata.go        # Encryption metadata
│   │
│   ├── api/                   # API client (keep, but refactor)
│   │   ├── client.go          # HTTP client wrapper
│   │   └── auth.go            # Token management
│   │
│   ├── config/                # Configuration (keep)
│   │   ├── config.go
│   │   └── daemon.go
│   │
│   ├── db/                    # SQLite (keep for sync state)
│   │   ├── db.go
│   │   ├── migrations.go
│   │   ├── folders.go
│   │   ├── files.go
│   │   └── queue.go
│   │
│   ├── sync/                  # Sync engine (keep, update)
│   │   ├── engine.go
│   │   ├── watcher.go
│   │   ├── hasher.go
│   │   └── conflict.go
│   │
│   └── ipc/                   # IPC (keep)
│       ├── server.go
│       ├── client.go
│       └── protocol.go
│
├── pkg/
│   ├── progress/              # ⭐ NEW: Progress bars
│   ├── compression/           # ⭐ NEW: ZIP handling
│   └── utils/                 # Utilities
│
└── go.mod
```

### Backend (Separate Repo - Not This Project)
```
darkstorage-backend/
├── MinIO cluster (S3-compatible)
├── PostgreSQL (metadata, users, billing)
├── Redis (caching, sessions)
├── Oracle eHSM (premium tier)
├── API Gateway (authentication, routing)
└── CDN (CloudFlare/Fastly)
```

---

## Implementation Phases

### Phase 1: Core Foundation (Week 1-2)
**Goal**: Working CLI with real MinIO backend

- [ ] Add dependencies (MinIO SDK, crypto libs)
- [ ] Implement `internal/storage/client.go` with MinIO SDK
- [ ] Implement authentication (login/logout/whoami)
- [ ] Implement basic operations (ls, put, get, rm)
- [ ] Add configuration management
- [ ] Test: `login → ls → put → get → rm`

**Deliverable**: CLI that can upload/download files to real MinIO

### Phase 2: Encryption Layer (Week 2-3)
**Goal**: Transparent client-side encryption

- [ ] Implement `internal/crypto/` package
- [ ] 3+1 key management system
- [ ] OS keychain integration
- [ ] Encrypt on upload, decrypt on download
- [ ] Key rotation logic
- [ ] Test: Encrypted file upload/download

**Deliverable**: All operations encrypted by default

### Phase 3: Advanced Features (Week 3-4)
**Goal**: Feature parity with competitors

- [ ] Storage classes (STANDARD, GLACIER, etc.)
- [ ] Pre-signed URLs (sharing)
- [ ] ZIP directory downloads
- [ ] File versioning
- [ ] Metadata management
- [ ] Search (by name, hash, metadata)
- [ ] Lifecycle policies

**Deliverable**: Full-featured CLI

### Phase 4: GUI Integration (Week 4-5)
**Goal**: Beautiful desktop app

- [ ] Update GUI to use new storage client
- [ ] Dashboard with storage usage
- [ ] File browser with drag-drop
- [ ] Sync settings UI
- [ ] Encryption status indicators
- [ ] Share management UI

**Deliverable**: Production-ready GUI

### Phase 5: Sync Daemon (Week 5-6)
**Goal**: Automatic background sync

- [ ] Update daemon to use new storage client
- [ ] Add encryption to sync operations
- [ ] Conflict resolution UI
- [ ] Bandwidth throttling
- [ ] Selective sync
- [ ] Offline mode

**Deliverable**: Dropbox-like sync experience

### Phase 6: Enterprise Features (Week 6-7)
**Goal**: HSM and compliance

- [ ] Oracle eHSM integration
- [ ] Groups & permissions (RBAC)
- [ ] Audit logging
- [ ] Compliance reporting
- [ ] SSO integration
- [ ] Organization management

**Deliverable**: Enterprise-ready platform

### Phase 7: Production Hardening (Week 7-8)
**Goal**: Bulletproof reliability

- [ ] Security audit
- [ ] Performance optimization
- [ ] Error handling & recovery
- [ ] Comprehensive testing
- [ ] Documentation
- [ ] Monitoring & logging

**Deliverable**: Production-ready, secure, fast

### Phase 8: Web3 Integration (Week 8-9) 🌐
**Goal**: Decentralized storage option

- [ ] Storj integration (decentralized S3-compatible)
- [ ] IPFS integration (content-addressed storage)
- [ ] Web3 backend selector in config
- [ ] Hybrid mode (Web2 + Web3 mirroring)
- [ ] IPFS gateway for sharing
- [ ] Storj DCS SDK integration

**Deliverable**: Web3-enabled storage (opt-in)

### Phase 9: Polish & Launch (Week 9-10)
**Goal**: Go to market

- [ ] Marketing materials (highlight Web3!)
- [ ] Pricing page
- [ ] Documentation site
- [ ] Video tutorials
- [ ] Beta testing
- [ ] Launch! 🚀

---

## Dependencies to Add

```bash
# Core storage
go get github.com/minio/minio-go/v7

# Encryption
go get golang.org/x/crypto
go get github.com/99designs/keyring

# Compression
go get github.com/klauspost/compress

# Progress bars
go get github.com/schollz/progressbar/v3

# Testing
go get github.com/stretchr/testify

# Already have:
# - fyne.io/fyne/v2 (GUI)
# - github.com/mattn/go-sqlite3 (DB)
# - github.com/spf13/cobra (CLI)
# - github.com/spf13/viper (Config)
# - github.com/fsnotify/fsnotify (File watcher)
```

---

## Configuration Schema (Final)

```yaml
# ~/.darkstorage/config.yaml

# Storage backend
storage:
  endpoint: s3.darkstorage.io
  use_ssl: true
  region: us-east-1

# User credentials (from login)
auth:
  access_key: ${DARKSTORAGE_ACCESS_KEY}
  secret_key: ${DARKSTORAGE_SECRET_KEY}
  token: ${DARKSTORAGE_TOKEN}
  token_expiry: 2024-12-31T23:59:59Z

# Account tier (detected from backend)
account:
  tier: premium  # free, standard, premium
  hsm_enabled: true
  storage_quota: 1TB
  bandwidth_quota: unlimited

# Encryption
encryption:
  enabled: true
  algorithm: aes-256-gcm
  keystore: keychain  # keychain, vault, file
  active_key_id: key-2024-02-001
  backup_keys:
    - key-2024-01-001  # 90 days old
    - key-2023-12-001  # 60 days old
    - key-2023-11-001  # 30 days old
  rotation_interval: 90d

# Default settings
defaults:
  storage_class: STANDARD
  encryption: true
  multipart_threshold: 64MB
  multipart_chunk_size: 5MB

# Sync daemon
sync:
  enabled: true
  folders:
    - id: 1
      local: ~/Documents
      remote: my-bucket/Documents
      direction: bidirectional
      enabled: true

# Performance
performance:
  concurrent_uploads: 4
  concurrent_downloads: 4
  bandwidth_limit_up: 0    # 0 = unlimited
  bandwidth_limit_down: 0

# UI preferences
ui:
  theme: dark
  notifications: true
  show_hidden_files: false
```

---

## Success Metrics

### Technical Metrics:
- [ ] Upload speed: Max out network bandwidth
- [ ] Download speed: Max out network bandwidth
- [ ] Encryption overhead: <5% performance impact
- [ ] GUI responsiveness: 60 FPS always
- [ ] Sync latency: <5 seconds to detect changes
- [ ] Memory usage: <200 MB for daemon

### Feature Completeness:
- [ ] 100% S3 API compatibility
- [ ] All AWS storage classes supported
- [ ] Pre-signed URLs working
- [ ] ZIP downloads working
- [ ] Encryption working (3+1 keys)
- [ ] HSM tier working (Oracle eHSM)
- [ ] Desktop sync working (bidirectional)

### Quality Metrics:
- [ ] Unit test coverage: >80%
- [ ] Integration tests: All critical paths
- [ ] Security audit: No critical issues
- [ ] Performance tests: Pass all benchmarks
- [ ] Documentation: Complete user guide

---

## Competitive Pricing

### Our Pricing (Proposed):
```
Free Tier:    10 GB     - FREE
Standard:     100 GB    - $5/month
Standard:     1 TB      - $10/month
Standard:     10 TB     - $50/month
Premium:      Unlimited - $99/month (includes HSM)
```

### Competitors (Reference):
```
AWS S3:       1 TB      - $23/month + bandwidth
Backblaze B2: 1 TB      - $6/month + bandwidth
Cloudflare R2:1 TB      - $15/month (zero egress)
Wasabi:       1 TB      - $7/month (no egress)
```

**Our Advantage**: Encryption included, better UI, sync daemon built-in

---

## Launch Checklist

### Pre-Launch (Before Public Release):
- [ ] Feature complete (all tiers)
- [ ] Security audit passed
- [ ] Performance benchmarks met
- [ ] Documentation complete
- [ ] Marketing site live
- [ ] Pricing page ready
- [ ] Payment processing (Stripe)
- [ ] Beta testing complete (50+ users)

### Launch Day:
- [ ] Press release
- [ ] Social media announcement
- [ ] Product Hunt launch
- [ ] HackerNews post
- [ ] Reddit announcement
- [ ] Email existing waitlist

### Post-Launch:
- [ ] Monitor errors/crashes
- [ ] Support tickets
- [ ] User feedback
- [ ] Performance monitoring
- [ ] Iterate based on feedback

---

## Next Steps (Immediate)

1. **Review this plan** - Make sure we agree on scope
2. **Set up backend** - Deploy MinIO cluster (or use existing?)
3. **Start Phase 1** - Core foundation implementation
4. **Daily standups** - Track progress
5. **Aggressive timeline** - 9 weeks to launch

---

## Questions to Answer Before Starting

1. **Do we have a MinIO backend deployed?**
   - If no: Need to deploy MinIO cluster first
   - If yes: What's the endpoint? Credentials?

2. **Authentication backend?**
   - How do users register/login?
   - Is there an API for user management?
   - Token format (JWT?)

3. **Billing system?**
   - How do we track storage usage?
   - How do we charge users?
   - Integration with Stripe/payment processor?

4. **HSM access?**
   - Do we have Oracle Cloud account?
   - eHSM cluster deployed?
   - Cost structure understood?

5. **Domain & Infrastructure?**
   - Is `s3.darkstorage.io` ready?
   - SSL certificates?
   - CDN configured?

---

## Web3 Integration (Advanced Feature) 🌐

### Overview
Web3 storage is an **opt-in feature** for users who want:
- Decentralized storage (no single point of failure)
- Content-addressed data (IPFS)
- Distributed cloud storage (Storj DCS)
- Censorship resistance
- Crypto-native workflows

**Key Point**: This is OPTIONAL - most users will use traditional S3 backend, but Web3 enthusiasts get decentralized options.

### Storage Backend Options

When Web3 is enabled, users can choose:

```
┌─────────────────────────────────────────────┐
│ Storage Backend Selector                    │
├─────────────────────────────────────────────┤
│ ○ Traditional (s3.darkstorage.io)          │
│   • Fast, reliable, centralized             │
│   • Storage classes (GLACIER, etc.)         │
│   • HSM encryption available                │
│                                             │
│ ○ Storj DCS (Decentralized Cloud Storage)  │
│   • Distributed across nodes globally       │
│   • Geo-redundant by default                │
│   • Pay with STORJ tokens or credit card    │
│                                             │
│ ○ IPFS (InterPlanetary File System)        │
│   • Content-addressed storage               │
│   • Public/private networks                 │
│   • Pinning services (Pinata, Web3.Storage) │
│                                             │
│ ○ Hybrid (Mirror to multiple backends)     │
│   • Upload to both Web2 and Web3            │
│   • Maximum redundancy                      │
│   • Auto-failover                           │
└─────────────────────────────────────────────┘
```

### Architecture with Web3

```
Client (CLI/GUI)
    │
    ├─── Storage Abstraction Layer
    │       │
    │       ├─── Traditional Backend
    │       │       └─── MinIO (s3.darkstorage.io)
    │       │
    │       ├─── Storj DCS Backend
    │       │       └─── Storj Network (distributed nodes)
    │       │
    │       ├─── IPFS Backend
    │       │       └─── IPFS Network (content-addressed)
    │       │
    │       └─── Hybrid Backend
    │               └─── Upload to multiple backends simultaneously
    │
    └─── Encryption Layer (same for all backends)
            └─── 3+1 key rotation works everywhere
```

### Implementation Details

#### 1. Storj DCS Integration

**Storj Features:**
- S3-compatible API (easy to integrate!)
- Distributed across ~10,000 nodes globally
- Automatic geo-redundancy
- Enterprise-grade encryption by default
- Pay only for what you use

**Configuration:**
```yaml
web3:
  enabled: true
  backend: storj

storj:
  access_grant: ${STORJ_ACCESS_GRANT}
  satellite: us1.storj.io:7777
  encryption_passphrase: ${STORJ_PASSPHRASE}

  # Optional: S3 compatibility mode
  s3_gateway: gateway.storjshare.io
  s3_credentials:
    access_key: ${STORJ_S3_ACCESS_KEY}
    secret_key: ${STORJ_S3_SECRET_KEY}
```

**CLI Usage:**
```bash
# Enable Web3
darkstorage web3 enable

# Configure Storj
darkstorage web3 backend storj --access-grant "1abc..."

# Upload to Storj (same commands!)
darkstorage put file.txt my-bucket/
→ [Uploading to Storj DCS...]
→ ✓ Uploaded to decentralized network (12 nodes)

# Show backend info
darkstorage whoami
→ Storage Backend: Storj DCS (Decentralized)
→ Nodes storing data: 12
→ Geographic distribution: US-East, US-West, EU, Asia
```

#### 2. IPFS Integration

**IPFS Features:**
- Content-addressed (files identified by hash)
- Immutable by default
- Peer-to-peer network
- Great for public datasets, NFT metadata, static sites

**Configuration:**
```yaml
web3:
  enabled: true
  backend: ipfs

ipfs:
  # Local IPFS node
  api_endpoint: http://localhost:5001
  gateway: http://localhost:8080

  # Or use pinning service
  pinning_service: pinata  # pinata, web3.storage, nft.storage
  pinning_api_key: ${PINATA_API_KEY}

  # Settings
  pin_files: true  # Keep files pinned (don't let them disappear)
  public: false    # Use private IPFS network
```

**CLI Usage:**
```bash
# Configure IPFS
darkstorage web3 backend ipfs --pinning-service pinata

# Upload to IPFS
darkstorage put image.jpg ipfs://
→ [Uploading to IPFS...]
→ ✓ Uploaded: ipfs://QmX7Y8Z9... (CIDv1)
→ ✓ Pinned on Pinata

# Share IPFS file
darkstorage share ipfs://QmX7Y8Z9...
→ Public Gateway: https://gateway.pinata.cloud/ipfs/QmX7Y8Z9...
→ IPFS Native: ipfs://QmX7Y8Z9...

# Download from IPFS (by content hash)
darkstorage get ipfs://QmX7Y8Z9... ./image.jpg
```

#### 3. Hybrid Mode (Best of Both Worlds)

**Hybrid Features:**
- Upload to BOTH traditional and Web3 backends
- Automatic redundancy
- Failover if one backend is down
- Choose download source (fastest or cheapest)

**Configuration:**
```yaml
web3:
  enabled: true
  backend: hybrid

hybrid:
  backends:
    - type: traditional
      endpoint: s3.darkstorage.io
      priority: 1  # Try this first for downloads

    - type: storj
      satellite: us1.storj.io:7777
      priority: 2  # Fallback

    - type: ipfs
      pinning_service: pinata
      priority: 3  # Last resort

  upload_strategy: all  # Upload to all backends
  download_strategy: fastest  # Use fastest for downloads
```

**CLI Usage:**
```bash
# Enable hybrid mode
darkstorage web3 backend hybrid

# Upload (goes to ALL backends)
darkstorage put file.txt my-bucket/
→ [Uploading to 3 backends...]
→ ✓ s3.darkstorage.io (120ms)
→ ✓ Storj DCS (340ms, 12 nodes)
→ ✓ IPFS (580ms, pinned on Pinata)
→ File available on: Traditional, Storj, IPFS

# Download (automatically picks fastest)
darkstorage get my-bucket/file.txt
→ [Checking availability...]
→ ✓ Downloading from s3.darkstorage.io (fastest)

# Force download from specific backend
darkstorage get my-bucket/file.txt --backend storj
```

### File Structure Changes

```
darkstorage-cli/
├── internal/
│   ├── storage/
│   │   ├── client.go              # Storage abstraction interface
│   │   ├── traditional.go         # MinIO/S3 implementation
│   │   ├── storj.go              # ⭐ NEW: Storj DCS implementation
│   │   ├── ipfs.go               # ⭐ NEW: IPFS implementation
│   │   └── hybrid.go             # ⭐ NEW: Multi-backend orchestrator
│   │
│   └── web3/                      # ⭐ NEW PACKAGE
│       ├── config.go              # Web3 configuration
│       ├── storj_client.go        # Storj SDK wrapper
│       ├── ipfs_client.go         # IPFS SDK wrapper
│       └── pinning.go             # Pinning service integrations
│
└── cmd/
    └── web3.go                    # ⭐ NEW: Web3 management commands
```

### Storage Abstraction Interface

```go
// internal/storage/client.go

type StorageBackend interface {
    // Standard operations (work for ALL backends)
    Upload(ctx context.Context, src io.Reader, dest string, opts *UploadOptions) error
    Download(ctx context.Context, src string, dest io.Writer, opts *DownloadOptions) error
    Delete(ctx context.Context, path string) error
    List(ctx context.Context, prefix string, opts *ListOptions) ([]FileInfo, error)
    Stat(ctx context.Context, path string) (*FileInfo, error)

    // Backend-specific info
    BackendType() string  // "traditional", "storj", "ipfs", "hybrid"
    BackendInfo() map[string]interface{}
}

// Implementations:
type TraditionalBackend struct { ... }  // MinIO/S3
type StorjBackend struct { ... }        // Storj DCS
type IPFSBackend struct { ... }         // IPFS
type HybridBackend struct { ... }       // Multi-backend
```

### Dependencies for Web3

```bash
# Storj DCS
go get storj.io/uplink

# IPFS
go get github.com/ipfs/go-ipfs-api
go get github.com/ipfs/kubo  # Local IPFS node (optional)

# Pinning services
go get github.com/ipfs/go-pinning-service-http-client
```

### Pricing with Web3

**Traditional Tier:**
- Dark Storage hosted: $5-50/month
- You control the backend

**Storj Tier (Web3):**
- Pay-as-you-go to Storj network
- ~$4/TB/month storage
- ~$7/TB egress
- We don't charge extra (just pass through costs)

**IPFS Tier (Web3):**
- Free for public data
- Pinning services: $1-20/month (Pinata, Web3.Storage)
- We don't charge extra

**Hybrid Tier:**
- Traditional + Web3
- Best redundancy
- Premium feature: $99/month (includes both)

### Web3 CLI Commands

```bash
# Enable/disable Web3
darkstorage web3 enable
darkstorage web3 disable
darkstorage web3 status

# Configure backend
darkstorage web3 backend storj --access-grant "..."
darkstorage web3 backend ipfs --pinning-service pinata
darkstorage web3 backend hybrid

# Web3-specific operations
darkstorage web3 storj info
→ Satellite: us1.storj.io
→ Nodes storing data: 12
→ Total stored: 2.3 GB
→ Bandwidth used: 450 MB

darkstorage web3 ipfs pin ls
→ QmX7Y8Z9... (1.2 MB) - image.jpg
→ QmA1B2C3... (340 KB) - document.pdf

darkstorage web3 ipfs gateway
→ Local: http://localhost:8080/ipfs/
→ Public: https://gateway.pinata.cloud/ipfs/

# Show where files are stored
darkstorage ls my-bucket/ --show-backends
→ file1.txt    STANDARD    [Traditional, Storj]
→ file2.jpg    GLACIER     [Traditional]
→ file3.pdf    WEB3        [Storj, IPFS]
```

### Marketing Angle

**Traditional Cloud Users:**
- "Just works" - default s3.darkstorage.io
- No crypto, no complexity
- AWS S3 compatibility

**Web3 Enthusiasts:**
- "Enable Web3" toggle
- Decentralized storage (Storj)
- Content-addressed (IPFS)
- Censorship-resistant
- Crypto-friendly

**Enterprise (Hybrid):**
- "Best of both worlds"
- Traditional reliability + Web3 redundancy
- Geographic distribution
- Disaster recovery

### Implementation Priority

**Phase 1-7**: Traditional backend (get to market fast)
**Phase 8**: Add Web3 as opt-in feature
**Phase 9**: Market as unique differentiator

This gives us:
1. **Fast launch** with traditional backend
2. **Unique feature** no competitor has (Web3 + Traditional hybrid)
3. **Attract crypto community** (huge TAM)
4. **Hedge against centralization** (Web3 is the future)

---

Let's build something amazing! 🚀🐱
