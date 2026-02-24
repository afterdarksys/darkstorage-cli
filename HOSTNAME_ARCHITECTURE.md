# Dark Storage - Hostname Architecture

**Clear separation of concerns across subdomains**

---

## Production Hostnames

### Core Services

**s3.darkstorage.io**
- **Purpose**: Alias for storage.darkstorage.io (S3 API)
- **Used by**: Users who prefer `s3.*` naming convention
- **Protocol**: HTTPS (TLS required)
- **Ports**: 443 (HTTPS)
- **DNS**: CNAME → storage.darkstorage.io
- **Examples**:
  ```bash
  # CLI usage (both work identically)
  darkstorage put file.txt my-bucket/
  → Can use either storage.darkstorage.io or s3.darkstorage.io

  # AWS CLI compatibility
  aws s3 cp file.txt s3://my-bucket/ --endpoint-url https://s3.darkstorage.io
  # OR
  aws s3 cp file.txt s3://my-bucket/ --endpoint-url https://storage.darkstorage.io
  ```

---

**storage.darkstorage.io**
- **Purpose**: Primary storage backend + Management API & web interface
- **Used by**: CLI, SDK, web UI, account management
- **Dual Purpose**:
  - **S3 Storage API**: Full S3-compatible API (same as s3.darkstorage.io)
  - **Management UI**: User management, billing, account settings
- **Features**:
  - ✅ S3-compatible storage operations (PUT, GET, DELETE objects)
  - ✅ Bucket operations
  - ✅ User registration/login
  - ✅ Account management
  - ✅ Billing & invoices
  - ✅ API key generation
  - ✅ Organization settings
  - ✅ SSO configuration
- **Examples**:
  ```bash
  # Storage operations (S3 API)
  darkstorage put file.txt my-bucket/
  → Uploads to https://storage.darkstorage.io/my-bucket/file.txt

  # AWS CLI compatibility
  aws s3 cp file.txt s3://my-bucket/ --endpoint-url https://storage.darkstorage.io

  # Management UI
  https://storage.darkstorage.io/login
  https://storage.darkstorage.io/account/billing
  https://storage.darkstorage.io/api/v1/users
  ```

- **Note**: `s3.darkstorage.io` is an alias/CNAME to `storage.darkstorage.io` for users who prefer the `s3.*` convention

---

**console.darkstorage.io**
- **Purpose**: Web-based dashboard & file browser
- **Used by**: Users who want GUI access to their storage
- **Features**:
  - File browser (upload/download/delete)
  - Bucket management
  - User management
  - Storage analytics
  - DR dashboard
  - Email viewer (msgs.global integration)
  - Settings & preferences
- **Examples**:
  ```
  https://console.darkstorage.io/
  → Main dashboard

  https://console.darkstorage.io/buckets/my-bucket
  → Browse files in bucket

  https://console.darkstorage.io/dr
  → Disaster Recovery dashboard

  https://console.darkstorage.io/email
  → Email viewer (msgs.global embedded)
  ```

---

### Disaster Recovery Services

**disaster-mail.darkstorage.io**
- **Purpose**: Disaster Mail MX backup (email failover)
- **Used by**: SMTP servers when client's mail server is down
- **Ports**: 25 (SMTP), 587 (submission), 993 (IMAPS)
- **DNS Example**:
  ```dns
  # Client's DNS configuration
  client.com.  IN  MX  10  mail.client.com.          # Primary
  client.com.  IN  MX  90  disaster-mail.darkstorage.io.  # Backup
  ```

---

**dr-web.darkstorage.io**
- **Purpose**: Website/app failover (Instant DR mirrors)
- **Used by**: Hosting client websites/apps during disasters
- **Dynamic routing**: `{client-id}.dr-web.darkstorage.io`
- **Examples**:
  ```
  # Client's normal site
  https://www.client.com
  → Points to their infrastructure (normal operation)

  # During disaster (DNS switched by onedns.io)
  https://www.client.com
  → Points to client123.dr-web.darkstorage.io
  → Serves from DR mirror
  ```

---

### Authentication & Identity

**auth.darkstorage.io**
- **Purpose**: SSO and authentication service
- **Used by**: All Dark Storage platforms for unified login
- **Features**:
  - OAuth 2.0 provider
  - SAML 2.0 provider
  - JWT token issuance
  - User directory (LDAP integration)
  - Third-party SSO (Google, Microsoft, Okta)
- **Examples**:
  ```
  https://auth.darkstorage.io/oauth/authorize
  https://auth.darkstorage.io/saml/login
  https://auth.darkstorage.io/api/v1/token/validate
  ```

---

### API Endpoints

**api.darkstorage.io**
- **Purpose**: RESTful API for programmatic access
- **Used by**: CLI, SDKs, integrations
- **Features**:
  - User management
  - Bucket operations
  - File metadata
  - DR operations
  - Email operations (msgs.global proxy)
  - Billing/usage
- **Examples**:
  ```
  GET  https://api.darkstorage.io/v1/user
  POST https://api.darkstorage.io/v1/buckets
  GET  https://api.darkstorage.io/v1/dr/status
  GET  https://api.darkstorage.io/v1/email/queue
  ```

---

### Partner Services

**msgs.global**
- **Purpose**: Email platform (integrated with Dark Storage)
- **Used by**: Email management, webmail, disaster mail
- **Features**:
  - Full webmail interface
  - IMAP/SMTP access
  - Disaster Mail queue viewer
  - Email analytics
- **SSO**: Integrated with auth.darkstorage.io
- **Examples**:
  ```
  https://msgs.global/
  → Full email client

  https://msgs.global/disaster-mail/client.com
  → View queued emails during DR

  https://api.msgs.global/v1/mailboxes
  → API for email operations
  ```

---

**onedns.io**
- **Purpose**: DNS management with emergency admin access
- **Used by**: Dark Storage DR for automatic DNS failover
- **Features**:
  - DNS record management
  - Emergency DNS updates (admin key)
  - Health-based DNS routing
  - Automatic failover
- **Examples**:
  ```
  https://onedns.io/zones/client.com
  → Manage DNS records

  POST https://api.onedns.io/v1/emergency/failover
  Authorization: Bearer {ADMIN_KEY}
  → Trigger emergency DNS change
  ```

---

## Development/Staging Hostnames

**s3-staging.darkstorage.io**
- Staging S3 API for testing

**console-staging.darkstorage.io**
- Staging dashboard

**api-staging.darkstorage.io**
- Staging API

---

## Internal Hostnames (Not Public)

**minio-01.internal.darkstorage.io**
- Internal MinIO cluster node 1

**minio-02.internal.darkstorage.io**
- Internal MinIO cluster node 2

**postgres-01.internal.darkstorage.io**
- Internal PostgreSQL database

**redis-01.internal.darkstorage.io**
- Internal Redis cache

---

## Complete User Journey (Hostnames in Action)

### 1. User Signs Up
```
https://storage.darkstorage.io/signup
→ Creates account
→ Redirects to auth.darkstorage.io for OAuth setup
→ Returns to console.darkstorage.io/welcome
```

### 2. User Uploads Files (CLI)
```bash
darkstorage login
→ Authenticates with auth.darkstorage.io
→ Stores token locally

darkstorage put file.txt my-bucket/
→ Uploads to s3.darkstorage.io/my-bucket/file.txt
```

### 3. User Browses Files (Web)
```
https://console.darkstorage.io/buckets/my-bucket
→ Loads dashboard from console.darkstorage.io
→ Fetches file list from api.darkstorage.io/v1/buckets/my-bucket/objects
→ Downloads files from s3.darkstorage.io/my-bucket/file.txt
```

### 4. Disaster Strikes - Website Down
```
onedns.io detects client.com is down
→ Calls api.darkstorage.io/v1/dr/activate
→ Dark Storage spins up DR mirror at client123.dr-web.darkstorage.io
→ Dark Storage calls onedns.io API with admin key
→ onedns.io updates DNS: www.client.com → client123.dr-web.darkstorage.io
→ User sees notification in console.darkstorage.io/dr
```

### 5. Disaster Strikes - Email Down
```
SMTP server tries to deliver to mail.client.com (MX 10) → fails
→ Tries disaster-mail.darkstorage.io (MX 90) → succeeds
→ Email queued in msgs.global
→ User sees "47 emails queued" in console.darkstorage.io/email
→ Clicks "View Emails" → opens msgs.global (SSO auto-login)
```

### 6. User Manages Everything
```
console.darkstorage.io/
├── Files (s3.darkstorage.io backend)
├── DR Status (api.darkstorage.io/v1/dr/status)
├── Email (msgs.global API embedded)
└── DNS (onedns.io API for status)

All in one dashboard!
```

---

## CLI Configuration

```yaml
# ~/.darkstorage/config.yaml

endpoints:
  # Storage API (S3-compatible) - PRIMARY ENDPOINT
  # Both storage.darkstorage.io and s3.darkstorage.io work
  storage: https://storage.darkstorage.io

  # Management API
  api: https://api.darkstorage.io

  # Authentication
  auth: https://auth.darkstorage.io

  # Console (for opening in browser)
  console: https://console.darkstorage.io

  # Email platform
  email: https://msgs.global

  # DNS management
  dns: https://onedns.io

account:
  access_key: ${DARKSTORAGE_ACCESS_KEY}
  secret_key: ${DARKSTORAGE_SECRET_KEY}
```

---

## DNS Configuration Template

For clients setting up Dark Storage services:

```dns
; Dark Storage Configuration for client.com

; Storage (optional - if using custom domain)
storage.client.com.     IN  CNAME  s3.darkstorage.io.

; Website DR (configured during failover)
www.client.com.         IN  A      192.168.1.100  ; Normal
; During DR, onedns.io changes to:
; www.client.com.       IN  CNAME  client123.dr-web.darkstorage.io.

; Email DR (always configured as backup MX)
client.com.             IN  MX  10  mail.client.com.
client.com.             IN  MX  90  disaster-mail.darkstorage.io.

; Verification (for account validation)
_darkstorage.client.com. IN  TXT  "darkstorage-verification=abc123..."
```

---

## SSL/TLS Certificates

All public hostnames use Let's Encrypt with automatic renewal:

- ✅ s3.darkstorage.io
- ✅ storage.darkstorage.io
- ✅ console.darkstorage.io
- ✅ api.darkstorage.io
- ✅ auth.darkstorage.io
- ✅ disaster-mail.darkstorage.io
- ✅ dr-web.darkstorage.io
- ✅ *.dr-web.darkstorage.io (wildcard for client mirrors)
- ✅ msgs.global
- ✅ api.msgs.global
- ✅ onedns.io
- ✅ api.onedns.io

---

## Load Balancing & HA

Each service runs behind load balancers:

```
               [Cloudflare CDN]
                      │
         ┌────────────┼────────────┐
         │            │            │
    [US-East]    [US-West]    [EU-West]
         │            │            │
    [Load Balancer] [Load Balancer] [Load Balancer]
         │            │            │
    [App Servers] [App Servers] [App Servers]
```

**Geographic routing:**
- US users → us-east or us-west
- EU users → eu-west
- Asia users → asia-pacific (future)

---

## Summary

| Hostname | Purpose | Users | Protocol |
|----------|---------|-------|----------|
| **s3.darkstorage.io** | S3 API | CLI, SDK, Tools | HTTPS |
| **storage.darkstorage.io** | Management | Web UI | HTTPS |
| **console.darkstorage.io** | Dashboard | Web UI | HTTPS |
| **api.darkstorage.io** | REST API | CLI, SDK | HTTPS |
| **auth.darkstorage.io** | SSO/Auth | All platforms | HTTPS |
| **disaster-mail.darkstorage.io** | Email DR | SMTP servers | SMTP/IMAP |
| **dr-web.darkstorage.io** | Website DR | Web traffic | HTTPS |
| **msgs.global** | Email platform | Users | HTTPS/SMTP/IMAP |
| **onedns.io** | DNS management | Dark Storage DR | HTTPS |

Clean, organized, scalable! 🚀🐱
