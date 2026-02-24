# SSO Integration - Dark Storage ↔ msgs.global

**Vision**: One login, access to both platforms seamlessly

---

## User Experience

### Login Flow (SSO)

```
User visits storage.darkstorage.io
    ↓
Clicks "Sign In"
    ↓
SSO Provider (OAuth 2.0 / SAML)
    ├─→ Google Workspace
    ├─→ Microsoft Azure AD
    ├─→ Okta
    └─→ Dark Storage native
    ↓
User authenticates once
    ↓
Receives JWT token (session valid across platforms)
    ↓
Can now access:
    ├─→ storage.darkstorage.io (files, DR dashboard, settings)
    └─→ msgs.global (email, queued disaster mail)
```

### Seamless Platform Switching

**In Dark Storage Dashboard:**
```
┌────────────────────────────────────────────────┐
│  Dark Storage                    [user@co.com ▼]│
├────────────────────────────────────────────────┤
│                                                │
│  Quick Links:                                  │
│  🗂️  Files & Storage                          │
│  🔄  Sync Settings                             │
│  🚨  Disaster Recovery                         │
│  📧  Email (msgs.global) →                    │
│      └─ Opens msgs.global with SSO            │
│                                                │
└────────────────────────────────────────────────┘
```

**User clicks "Email" link:**
- Browser navigates to `https://msgs.global?sso_token=...`
- msgs.global validates token
- User is logged in automatically
- No second login required!

---

## Technical Architecture

### SSO Token Flow

```
┌─────────────────┐
│  User Browser   │
└────────┬────────┘
         │
         ├──[1]── Login Request ──────────────────┐
         │                                         │
         │                                    ┌────▼─────┐
         │                                    │   SSO    │
         │                                    │ Provider │
         │                                    │ (OAuth)  │
         │                                    └────┬─────┘
         │                                         │
         │◄───[2]── JWT Token ─────────────────────┘
         │
         ├──[3]── Access storage.darkstorage.io ──┐
         │         (with JWT token)                │
         │                                    ┌────▼──────────┐
         │                                    │ Dark Storage  │
         │◄───[4]── Dashboard ────────────────│   Backend     │
         │                                    └───────────────┘
         │
         ├──[5]── Click "Email" link ─────────────┐
         │         (includes sso_token param)      │
         │                                    ┌────▼──────────┐
         │                                    │  msgs.global  │
         │                                    │    Backend    │
         │                                    └────┬──────────┘
         │                                         │
         │                                    [Validate token]
         │                                         │
         │◄───[6]── Email interface ───────────────┘
         │         (logged in automatically)
         │
```

### JWT Token Structure

```json
{
  "iss": "auth.darkstorage.io",
  "sub": "user-12345",
  "email": "user@client.com",
  "name": "Ryan Smith",
  "exp": 1708876800,
  "iat": 1708790400,
  "platforms": {
    "storage": {
      "access": true,
      "tier": "enterprise",
      "features": ["dr", "hsm", "web3"]
    },
    "email": {
      "access": true,
      "domains": ["client.com", "client.net"],
      "disaster_mail": true
    }
  },
  "sso_provider": "google",
  "organization": "Client Corp"
}
```

### Shared Authentication Backend

Both platforms validate tokens from the same auth service:

```
┌───────────────────────────────────────┐
│  auth.darkstorage.io                  │
│  (Shared SSO Service)                 │
├───────────────────────────────────────┤
│  - Issue JWT tokens                   │
│  - Validate tokens                    │
│  - Refresh tokens                     │
│  - User directory (LDAP/AD)           │
│  - OAuth providers (Google, MS, Okta) │
└──────────┬────────────────────┬───────┘
           │                    │
    ┌──────▼──────┐      ┌──────▼──────┐
    │   storage   │      │msgs.global  │
    │ darkstorage │      │   (Email)   │
    │     .io     │      │             │
    └─────────────┘      └─────────────┘
```

---

## Platform Integration Points

### 1. Disaster Mail Notifications

When Disaster Mail activates, notification shows in Dark Storage dashboard:

```
┌────────────────────────────────────────────────┐
│  🚨 Disaster Mail Active                       │
├────────────────────────────────────────────────┤
│  Your mail server (mail.client.com) is down   │
│  We're queuing incoming emails (47 queued)    │
│                                                │
│  [View Queued Mail on msgs.global] ───────►   │
│  [Check DR Status]                            │
└────────────────────────────────────────────────┘
```

Click "View Queued Mail" → opens msgs.global, already logged in via SSO

---

### 2. Email Status in Storage Dashboard

```
┌────────────────────────────────────────────────┐
│  Dark Storage Dashboard                        │
├────────────────────────────────────────────────┤
│  Storage: 2.3 TB / 5 TB used                  │
│  DR Status: ✅ Healthy                         │
│  Email Status: 📧 47 emails queued (msgs.global)│
│                ↑                                │
│                └─ Click to open msgs.global    │
└────────────────────────────────────────────────┘
```

---

### 3. Unified Billing

One invoice covers both platforms:

```
Dark Storage - Invoice #2024-02-001

Enterprise Plan                        $999.00
├─ Storage (5 TB)                      included
├─ Disaster Recovery                   included
├─ HSM Encryption                      included
└─ Email (msgs.global - 20 mailboxes)  included

Total:                                 $999.00
```

User manages billing in Dark Storage → applies to msgs.global too

---

### 4. Unified Admin Panel

**Organization Settings** (accessible from either platform):

```
┌────────────────────────────────────────────────┐
│  Organization: Client Corp                     │
├────────────────────────────────────────────────┤
│                                                │
│  Users (25)                                    │
│  ┌──────────────────────────────────────────┐ │
│  │ Name          Email         Platforms    │ │
│  │ Ryan Smith    ryan@co.com   Storage+Email│ │
│  │ Jane Doe      jane@co.com   Storage+Email│ │
│  │ Bob Johnson   bob@co.com    Email only   │ │
│  └──────────────────────────────────────────┘ │
│                                                │
│  [Add User] [SSO Settings] [Billing]          │
│                                                │
└────────────────────────────────────────────────┘
```

Manage users once → applies to both platforms

---

## msgs.global Features (Email Platform)

### Core Email Features
- 📧 **Webmail interface** (modern, fast)
- 📨 **IMAP/SMTP** access (use with Outlook, Apple Mail, etc.)
- 📂 **Unlimited folders**
- 🔍 **Full-text search**
- 📎 **Large attachments** (up to 100 MB)
- 🗑️ **Spam filtering** (AI-powered)
- 🔐 **Encryption** (TLS in transit, encrypted at rest)

### Disaster Mail Features
- 🚨 **Queued mail viewer** (during DR events)
- 📊 **Queue status** (how many emails waiting)
- ✉️ **Send during disaster** (SMTP relay)
- 🔄 **Auto-delivery** (when primary recovers)
- 📈 **DR analytics** (how long was primary down, how many emails handled)

### Integration Features
- 🔗 **SSO with Dark Storage** (seamless login)
- 💾 **Email attachments → Dark Storage** (optional: auto-save to object storage)
- 🔐 **Shared encryption keys** (use same 3+1 key system)
- 📊 **Unified dashboard** (email stats visible in Dark Storage)

---

## Implementation Details

### Dark Storage Side

**Add "Email" link to dashboard:**

```go
// cmd/gui/dashboard.go

type DashboardView struct {
    // ... existing fields
    EmailButton *widget.Button
}

func NewDashboardView(client *api.Client) *DashboardView {
    emailBtn := widget.NewButton("📧 Email (msgs.global)", func() {
        // Get SSO token
        token := client.GetSSOToken()

        // Open msgs.global with SSO token
        url := fmt.Sprintf("https://msgs.global?sso_token=%s", token)
        browser.OpenURL(url)
    })

    return &DashboardView{
        EmailButton: emailBtn,
        // ...
    }
}
```

**Add email status widget:**

```go
// internal/api/email.go

type EmailStatus struct {
    DisasterMailActive bool   `json:"disaster_mail_active"`
    QueuedEmails      int    `json:"queued_emails"`
    Domain            string `json:"domain"`
    PrimaryHealthy    bool   `json:"primary_healthy"`
}

func (c *Client) GetEmailStatus(ctx context.Context) (*EmailStatus, error) {
    // Call msgs.global API to get status
    resp, err := c.httpClient.Get("https://api.msgs.global/v1/status?domain=" + c.domain)
    // ...
}
```

### msgs.global Side

**SSO token validation endpoint:**

```go
// msgs.global backend

POST /api/v1/auth/sso
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "platform": "darkstorage"
}

Response:
{
  "success": true,
  "session_token": "session-abc123...",
  "user": {
    "email": "user@client.com",
    "name": "Ryan Smith",
    "domains": ["client.com"],
    "features": ["disaster_mail", "unlimited_storage"]
  }
}
```

**Disaster Mail queue API:**

```go
// msgs.global backend

GET /api/v1/disaster-mail/queue?domain=client.com

Response:
{
  "active": true,
  "queued_emails": 47,
  "emails": [
    {
      "id": "msg-001",
      "from": "customer@example.com",
      "to": "sales@client.com",
      "subject": "Question about your product",
      "received_at": "2024-02-24T10:30:00Z",
      "size": 12456
    },
    // ... more emails
  ]
}
```

---

## User Onboarding Flow

### Step 1: User signs up for Dark Storage Enterprise

```
https://darkstorage.io/signup

[Sign Up for Enterprise]
- Name: Ryan Smith
- Email: ryan@client.com
- Company: Client Corp
- Choose SSO provider: [Google Workspace ▼]
```

### Step 2: Configure SSO

```
Set up Google Workspace SSO:
1. Domain: client.com
2. OAuth Client ID: [provided by Google]
3. OAuth Client Secret: [provided by Google]
4. Authorized domains: storage.darkstorage.io, msgs.global
```

### Step 3: Enable Disaster Mail

```
Disaster Mail Setup:
- Domain: client.com
- Add MX record to DNS:
  @ IN MX 90 disaster-mail.darkstorage.io

[Test Configuration] [Activate Disaster Mail]
```

### Step 4: Access msgs.global

```
Welcome to Dark Storage!

Your account includes:
✅ 5 TB storage
✅ Disaster Recovery
✅ Email (msgs.global with 20 mailboxes)

[Open Dashboard] [Set Up Email →]
                  └─ Opens msgs.global with SSO
```

---

## Configuration

**Dark Storage config:**

```yaml
# ~/.darkstorage/config.yaml

account:
  email: ryan@client.com
  organization: Client Corp
  tier: enterprise

sso:
  enabled: true
  provider: google
  client_id: abc123.apps.googleusercontent.com
  domains:
    - storage.darkstorage.io
    - msgs.global

email:
  platform: msgs.global
  disaster_mail:
    enabled: true
    domains:
      - client.com
      - client.net
  sso_integration: true
```

**CLI command to open email:**

```bash
# Open msgs.global in browser (SSO auto-login)
darkstorage email open

# Check disaster mail status
darkstorage email status
→ Disaster Mail: Active
→ Domain: client.com
→ Queued emails: 47
→ Primary server: DOWN
→ View at: https://msgs.global

# Quick link to msgs.global
darkstorage email web
→ Opens https://msgs.global?sso_token=...
```

---

## Benefits

### For Users
- ✅ **One login** for everything (storage + email)
- ✅ **Seamless experience** (click link, already logged in)
- ✅ **Unified billing** (one invoice)
- ✅ **Single admin panel** (manage users once)

### For Dark Storage
- ✅ **Stickiness** (users locked into ecosystem)
- ✅ **Cross-sell** (storage users → email, email users → storage)
- ✅ **Higher revenue** (bundles worth more)
- ✅ **Better retention** (integrated platforms = less churn)

### For Enterprise Customers
- ✅ **Simplified IT** (one vendor, one contract)
- ✅ **Better security** (centralized auth, SSO)
- ✅ **Compliance** (unified audit logs)
- ✅ **Cost savings** (bundle cheaper than separate services)

---

## Competitive Advantage

**Nobody else offers this:**

| Feature | Dark Storage + msgs.global | AWS | Google Workspace | Microsoft 365 |
|---------|---------------------------|-----|-----------------|---------------|
| Storage + Email integrated | ✅ SSO seamless | ❌ Separate | ✅ Integrated | ✅ Integrated |
| Disaster Recovery (website) | ✅ Built-in | ⚠️ CloudEndure | ❌ No | ❌ No |
| Disaster Mail (MX backup) | ✅ Built-in | ❌ No | ❌ No | ❌ No |
| Client-side encryption | ✅ Default | ❌ Extra | ❌ Enterprise only | ❌ Enterprise only |
| Web3 support | ✅ Storj + IPFS | ❌ No | ❌ No | ❌ No |
| Price (Enterprise) | **$999/mo** | **~$2000/mo** | **~$1500/mo** | **~$1800/mo** |

We're the **only** platform with:
- Storage + Email + DR (all integrated)
- SSO across platforms
- Disaster Mail
- Affordable pricing

---

## Next Steps

1. **Validate msgs.global availability** (is domain available?)
2. **Build SSO auth service** (shared between platforms)
3. **Implement SSO flow** (Dark Storage → msgs.global)
4. **Build msgs.global email platform** (or integrate existing?)
5. **Test seamless switching**
6. **Launch as integrated bundle**

---

**This is the killer feature.** One login, complete business infrastructure.

🚀🐱
