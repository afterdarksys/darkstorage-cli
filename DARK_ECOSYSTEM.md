# The Dark Ecosystem - Complete Platform Architecture

**Status**: Active Development
**SSO**: Unified across all services
**Vision**: Complete developer platform from code to production

---

## The Complete Service Mesh

### 🗄️ Storage & Data Layer
- **DarkStorage.io** - S3-compatible encrypted storage
  - Traditional (MinIO/S3)
  - Web3 (Storj + IPFS)
  - 3+1 encryption
  - Storage classes (STANDARD → DEEP_ARCHIVE)
  - Instant DR backups
  - API key management

### 🚢 Distribution & Deployment
- **ShipShack.io** - Software distribution services
  - Package registry (npm, docker, binary releases)
  - Version management
  - Release channels (stable, beta, dev)
  - Download CDN
  - Checksums & signatures

- **DarkShip.io** - Infrastructure as Code + Deployment Orchestrator
  - Receives from ShipShack
  - Deployment targets:
    - Kubernetes clusters
    - Docker hosts
    - Dark Infrastructure (HostScience)
    - Customer clouds (AWS/GCP/Azure)
  - Deployment strategies:
    - Blue/green
    - Canary
    - Rolling updates
  - Logs → DarkStorage
  - Monitoring integration

### 🌐 DNS & Networking
- **DNScience.io** - DNS management platform
  - Authoritative DNS
  - DNSSEC support
  - Geo-routing
  - Health checks
  - Failover automation
  - Analytics & insights

- **OneDNS.io** - Emergency DNS management
  - Instant failover
  - Disaster recovery DNS
  - MX record takeover (Disaster Mail)
  - Emergency site hosting
  - Admin API for automated changes
  - Integration with DarkShip for auto-routing

### 🧠 AI & Compute
- **AIServe.farm** - AI workload platform
  - GPU clusters
  - DarkStorage mount directly to jobs
  - Model training
  - Inference serving
  - Job queuing
  - Cost optimization
  - Results → DarkStorage

- **ComputeAPI.io** - General compute API (needs fixing!)
  - Serverless functions
  - Batch processing
  - WebAssembly runtime
  - Container execution
  - API gateway

### 🏢 Hosting & Infrastructure
- **HostScience** - Application hosting platform
  - Managed Kubernetes
  - Docker hosting
  - Static site hosting
  - Database hosting
  - Redis/cache layers
  - Load balancers

- **APIFirewall** - API security layer
  - Rate limiting
  - DDoS protection
  - JWT validation
  - API key management
  - Request logging → DarkStorage
  - Threat detection

### 📧 Communication
- **msgs.global** - Email platform
  - Professional email hosting
  - Webmail interface
  - IMAP/SMTP/POP3
  - SSO integration
  - Email forwarding
  - Spam filtering
  - Disaster Mail failover

---

## The Developer Journey

### 1. Build & Distribute

```bash
# Developer writes code
git commit -m "New feature"

# Publish to ShipShack
shipshack publish myapp:v1.0.0 \
  --channel stable \
  --platforms linux,darwin,windows \
  --storage darkstorage://releases/myapp

# Output:
# ✓ Published: myapp:v1.0.0
# ✓ Stored: darkstorage://releases/myapp/v1.0.0/
# ✓ CDN: https://cdn.shipshack.io/myapp/v1.0.0/
# ✓ Checksums: SHA256 verified
```

### 2. Deploy with DarkShip

```yaml
# darkship.yaml
name: myapp
version: v1.0.0

source:
  registry: shipshack.io/myapp
  version: v1.0.0

deployment:
  target: kubernetes
  cluster: production

  storage:
    mount: darkstorage://app-data
    logs: darkstorage://logs/myapp
    backups: darkstorage://backups/myapp
    class: STANDARD_IA  # Auto-tier to IA after 30 days

  ai:
    enabled: true
    provider: aiserve.farm
    resources:
      gpu: 1
      memory: 16GB
    mount: darkstorage://ml-models

  networking:
    dns: dnscience.io
    domains:
      - myapp.com
      - www.myapp.com
    failover: onedns.io
    firewall: apifirewall

  email:
    domain: myapp.com
    provider: msgs.global
    disaster_mail: true

  hosting:
    platform: hostscience
    replicas: 3
    health_check: /health

  monitoring:
    logs: true
    metrics: true
    alerts: true
```

```bash
# Deploy
darkship deploy darkship.yaml

# Output:
# ✓ Fetched: shipshack.io/myapp:v1.0.0
# ✓ Storage mounted: darkstorage://app-data
# ✓ AI cluster ready: aiserve.farm/gpu-001
# ✓ DNS configured: myapp.com → 203.0.113.10
# ✓ Failover ready: onedns.io watching
# ✓ Email configured: mx.msgs.global
# ✓ Deployed: hostscience/k8s/myapp (3 replicas)
# ✓ Firewall: apifirewall protecting
#
# URLs:
#   App:      https://myapp.com
#   Admin:    https://admin.myapp.com
#   Logs:     https://console.darkstorage.io/logs/myapp
#   Metrics:  https://hostscience.io/metrics/myapp
```

### 3. App Runs with Full Integration

```javascript
// Your app gets automatic access to everything

// Storage - no setup needed
const storage = require('@darkstorage/sdk');
const file = await storage.get('darkstorage://app-data/users.db');

// AI - seamlessly integrated
const ai = require('@aiserve/sdk');
const result = await ai.infer('darkstorage://ml-models/mymodel.onnx', data);

// Email - built-in
const email = require('@msgs/sdk');
await email.send({
  from: 'noreply@myapp.com',
  to: user.email,
  subject: 'Welcome!',
  template: 'welcome'
});

// DNS - dynamic updates
const dns = require('@dnscience/sdk');
await dns.updateRecord('myapp.com', 'A', newIP);

// Everything authenticated with same SSO token
```

---

## Single Sign-On Flow

### Initial Login

```bash
# User logs in once
dark login
# Opens browser → auth.darkstorage.io
# Successfully authenticated

# JWT token saved to ~/.dark/token
```

### Token Works Everywhere

```javascript
// Token structure (JWT)
{
  "sub": "user_abc123",
  "email": "ryan@afterdark.com",
  "services": {
    "darkstorage": { "tier": "professional" },
    "darkship": { "tier": "professional" },
    "shipshack": { "tier": "professional" },
    "aiserve": { "credits": 1000 },
    "msgs": { "domain": "afterdark.com" },
    "dnscience": { "zones": 10 },
    "onedns": { "enabled": true },
    "hostscience": { "clusters": 3 },
    "apifirewall": { "enabled": true }
  },
  "permissions": ["admin:*"],
  "iat": 1708790000,
  "exp": 1708876400
}
```

### Service Access

```bash
# All services recognize the same token

# Storage
darkstorage ls --token $DARK_TOKEN

# Deployment
darkship deploy --token $DARK_TOKEN

# AI
aiserve submit --token $DARK_TOKEN

# Email
msgs send --token $DARK_TOKEN

# DNS
dnscience update --token $DARK_TOKEN

# Or use API
curl https://api.darkstorage.io/v1/files \
  -H "Authorization: Bearer $DARK_TOKEN"
```

---

## Service Integration Matrix

| Service | Storage | Deploy | AI | Email | DNS | Hosting |
|---------|---------|--------|-----|-------|-----|---------|
| **DarkStorage** | ● | Logs | Models | Attachments | - | Static |
| **DarkShip** | ✓ | ● | Config | Notifications | Records | Deploy |
| **ShipShack** | ✓ | Source | - | Notifications | - | CDN |
| **AIServe.farm** | ✓ | - | ● | Results | - | Jobs |
| **msgs.global** | ✓ | - | - | ● | MX | Webmail |
| **DNScience** | Logs | - | - | MX | ● | - |
| **OneDNS** | Logs | Failover | - | MX | ✓ | ● |
| **HostScience** | ✓ | Target | - | - | A/AAAA | ● |
| **APIFirewall** | Logs | - | - | - | - | Protect |

**Legend:**
- ● = Primary function
- ✓ = Direct integration
- Text = Integration type

---

## Disaster Recovery Flow

### Website Down (Instant DR)

```yaml
# DarkShip monitors your site
# If site goes down...

1. OneDNS takes over
   - Changes A records → onedns.io
   - Serves emergency site
   - Shows: "Site temporarily unavailable"

2. Notifications
   - Email via msgs.global
   - SMS alert
   - Dashboard update

3. Automatic Failover
   - If you have backup site → route there
   - If using HostScience → auto-scale
   - If total failure → static site from DarkStorage

4. Recovery
   - Site comes back online
   - OneDNS verifies health
   - Routes traffic back
   - Logs everything → DarkStorage
```

### Email Down (Disaster Mail)

```yaml
# msgs.global monitors your MX
# If your mail server goes down...

1. MX Failover
   - High-priority MX → msgs.global
   - Accepts all mail
   - Stores in DarkStorage

2. User Access
   - Users log in via SSO
   - Read mail on msgs.global webmail
   - Send/receive normally

3. Recovery
   - Your server comes back
   - Mail forwards to your server
   - Sync happens automatically
   - MX priority restored
```

---

## Pricing Integration

### Single Bill for Everything

```
Dark Ecosystem - Professional Plan: $99/month
═══════════════════════════════════════════════

Storage (DarkStorage.io)
  ✓ 1TB storage (STANDARD)
  ✓ Unlimited bandwidth
  ✓ Client-side encryption
  ✓ Instant DR

Distribution (ShipShack.io)
  ✓ Unlimited packages
  ✓ CDN delivery
  ✓ Download analytics

Deployment (DarkShip.io)
  ✓ 10 deployments/month
  ✓ 3 target platforms
  ✓ Deployment logs

AI (AIServe.farm)
  ✓ 100 GPU hours/month
  ✓ Storage mounting
  ✓ Model hosting

Email (msgs.global)
  ✓ 10 email accounts
  ✓ 5GB mailbox each
  ✓ Disaster Mail

DNS (DNScience.io + OneDNS.io)
  ✓ 10 zones
  ✓ Emergency failover
  ✓ DNSSEC

Hosting (HostScience)
  ✓ 3 small instances
  ✓ Load balancing
  ✓ Auto-scaling

Security (APIFirewall)
  ✓ DDoS protection
  ✓ Rate limiting
  ✓ Request analytics

═══════════════════════════════════════════════
Total Value: $500+/month (if purchased separately)
Your Price: $99/month
```

---

## API Integration Example

### Full Stack App Deployment

```bash
# 1. Push code to ShipShack
git push shipshack main

# 2. ShipShack builds and stores
# Build → Store in DarkStorage → Notify DarkShip

# 3. DarkShip deploys
# Pull from DarkStorage → Deploy to HostScience

# 4. Configure services
curl -X POST https://api.darkship.io/v1/deployments \
  -H "Authorization: Bearer $DARK_TOKEN" \
  -d '{
    "app": "myapp",
    "services": {
      "storage": {
        "provider": "darkstorage",
        "buckets": ["app-data", "user-uploads", "backups"]
      },
      "ai": {
        "provider": "aiserve",
        "models": ["sentiment-analysis", "image-recognition"]
      },
      "email": {
        "provider": "msgs",
        "domain": "myapp.com",
        "disaster_recovery": true
      },
      "dns": {
        "provider": "dnscience",
        "domains": ["myapp.com"],
        "failover": {
          "provider": "onedns",
          "health_check": "https://myapp.com/health"
        }
      },
      "hosting": {
        "provider": "hostscience",
        "type": "kubernetes",
        "replicas": 3
      },
      "security": {
        "provider": "apifirewall",
        "rate_limit": "100/minute"
      }
    }
  }'

# Result: Fully deployed, monitored, secured, backed up
```

---

## Developer SDKs

### Unified SDK

```javascript
// npm install @dark/sdk

const Dark = require('@dark/sdk');

// Initialize with SSO token
const dark = new Dark({ token: process.env.DARK_TOKEN });

// Storage
await dark.storage.upload('file.txt', 'bucket/path');
await dark.storage.download('bucket/path', 'local.txt');

// Deployment
await dark.ship.deploy('myapp', { target: 'kubernetes' });

// AI
const result = await dark.ai.infer('model-name', inputData);

// Email
await dark.email.send({
  to: 'user@example.com',
  subject: 'Hello',
  body: 'World'
});

// DNS
await dark.dns.update('myapp.com', 'A', '203.0.113.10');

// All services share authentication automatically
```

---

## Next Steps

1. **Complete DarkStorage** (current focus)
   - ✅ Core storage operations
   - ✅ API key system
   - ✅ Encryption (3+1)
   - ✅ Storage classes
   - ⏳ Web console
   - ⏳ Production deployment

2. **Build Integration Layer**
   - SSO token service
   - Service registry
   - API gateway
   - Unified billing

3. **Launch Services**
   - DarkStorage → production
   - DarkShip → beta
   - ShipShack → alpha
   - Others → planning

4. **Fix ComputeAPI.io**
   - Investigate outage
   - Restore service
   - Integrate with ecosystem

---

## The Vision

**One account. One bill. One ecosystem.**

No more:
- Managing multiple cloud providers
- Stitching together services
- Separate billing
- Different authentication systems
- Complex integrations

Just:
```bash
dark login
dark deploy myapp
```

**And everything just works.** 🚀

---

*Last Updated: 2026-02-24*
*Status: Active Development*
