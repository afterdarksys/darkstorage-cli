# DarkStorage Audit & Compliance Logging Specification

## Overview
Every action in the DarkStorage platform MUST be logged with complete forensic detail for compliance, security monitoring, and incident response.

---

## Audit Event Categories

### 1. **File Operations**
All file access and modifications are logged with full metadata.

#### Events to Log:
```
FILE_UPLOAD          - File uploaded to storage
FILE_DOWNLOAD        - File downloaded/accessed
FILE_VIEW            - File viewed (without download)
FILE_DELETE          - File deleted (moved to trash)
FILE_PERMANENT_DELETE - File permanently removed
FILE_RESTORE         - File restored from trash
FILE_RENAME          - File renamed or moved
FILE_COPY            - File copied
FILE_METADATA_UPDATE - File metadata/tags changed
FILE_VERSION_CREATE  - New file version created
FILE_VERSION_RESTORE - Previous version restored
```

#### Required Fields:
```json
{
  "event_id": "evt_2024_abc123xyz",
  "timestamp": "2026-03-01T14:23:45.123Z",
  "event_type": "FILE_DOWNLOAD",
  "resource_type": "file",
  "resource_id": "file_xyz789",
  "resource_path": "my-bucket/documents/contract.pdf",
  "resource_size": 2458624,
  "resource_hash": "sha256:a1b2c3d4...",
  "actor": {
    "user_id": "user_123",
    "email": "alice@example.com",
    "name": "Alice Johnson",
    "role": "admin",
    "organization": "ACME Corp"
  },
  "session": {
    "session_id": "sess_abc123",
    "ip_address": "203.0.113.45",
    "geo_location": {
      "country": "US",
      "region": "California",
      "city": "San Francisco"
    },
    "user_agent": "DarkStorage CLI/1.0.0",
    "device_fingerprint": "fp_xyz789"
  },
  "context": {
    "method": "api", // api, web, cli, mobile
    "endpoint": "/v1/files/download",
    "compartment": "classified",
    "previous_value": null,
    "new_value": null,
    "reason": "Quarterly review",
    "ticket_number": "TICKET-1234"
  },
  "result": {
    "success": true,
    "status_code": 200,
    "error_message": null,
    "bytes_transferred": 2458624,
    "duration_ms": 1245
  },
  "compliance": {
    "retention_days": 2555, // 7 years for financial records
    "classification": "CONFIDENTIAL",
    "requires_mfa": true,
    "mfa_verified": true
  }
}
```

---

### 2. **Permission Changes**
Critical for security - who changed access to what.

#### Events to Log:
```
PERMISSION_GRANT     - Permission granted to user/group
PERMISSION_REVOKE    - Permission revoked
PERMISSION_MODIFY    - Permission level changed
SHARE_LINK_CREATE    - Public/private share link created
SHARE_LINK_ACCESS    - Someone accessed via share link
SHARE_LINK_REVOKE    - Share link invalidated
ACL_UPDATE           - Access Control List modified
COMPARTMENT_ASSIGN   - File assigned to security compartment
COMPARTMENT_REMOVE   - File removed from compartment
```

#### Required Fields (in addition to base fields):
```json
{
  "permission_change": {
    "subject_type": "user", // user, group, api_key
    "subject_id": "user_456",
    "subject_email": "bob@example.com",
    "permission_before": "read",
    "permission_after": "write",
    "granted_by": "user_123",
    "expiration": "2027-01-01T00:00:00Z",
    "conditions": {
      "ip_whitelist": ["203.0.113.0/24"],
      "time_restriction": "business_hours"
    }
  }
}
```

---

### 3. **Authentication & Access**
Who logged in, from where, and any suspicious activity.

#### Events to Log:
```
LOGIN_SUCCESS        - User logged in successfully
LOGIN_FAILURE        - Failed login attempt
LOGIN_MFA_CHALLENGE  - MFA challenge issued
LOGIN_MFA_SUCCESS    - MFA completed successfully
LOGIN_MFA_FAILURE    - MFA failed
LOGOUT               - User logged out
SESSION_EXPIRED      - Session timed out
PASSWORD_CHANGE      - Password changed
PASSWORD_RESET       - Password reset requested
API_KEY_CREATED      - API key generated
API_KEY_ROTATED      - API key rotated
API_KEY_REVOKED      - API key revoked
OAUTH_AUTHORIZE      - OAuth authorization granted
OAUTH_TOKEN_REFRESH  - OAuth token refreshed
```

#### Required Fields:
```json
{
  "auth_event": {
    "method": "oauth", // password, oauth, api_key, saml
    "provider": "authentik",
    "mfa_method": "totp", // totp, sms, hardware_token
    "login_device": {
      "device_type": "desktop",
      "os": "macOS 14.3",
      "browser": "Chrome 122"
    },
    "risk_score": 15, // 0-100, higher is riskier
    "risk_factors": [
      "new_device",
      "unusual_location"
    ],
    "consecutive_failures": 0,
    "account_locked": false
  }
}
```

---

### 4. **Data Security Events**
Encryption, integrity, and compliance.

#### Events to Log:
```
FILE_ENCRYPTED       - File encrypted
FILE_DECRYPTED       - File decrypted (access)
KEY_ROTATION         - Encryption key rotated
INTEGRITY_VERIFIED   - Hash verification passed
INTEGRITY_FAILED     - Hash mismatch detected
MALWARE_DETECTED     - Virus/malware found
MALWARE_QUARANTINED  - File quarantined
DATA_EXPORT          - Bulk data export
DATA_IMPORT          - Bulk data import
BACKUP_CREATED       - Backup snapshot created
BACKUP_RESTORED      - Data restored from backup
```

#### Required Fields:
```json
{
  "security_event": {
    "encryption_algorithm": "AES-256-GCM",
    "key_id": "key_production_2026",
    "hash_algorithm": "SHA-256",
    "expected_hash": "a1b2c3...",
    "actual_hash": "a1b2c3...",
    "threat_type": "ransomware", // virus, malware, ransomware, pii_leak
    "threat_signature": "EICAR-Test-File",
    "quarantine_reason": "Potential malware",
    "export_destination": "s3://backups/",
    "export_record_count": 15234
  }
}
```

---

### 5. **Administrative Actions**
System configuration and user management.

#### Events to Log:
```
USER_CREATED         - New user account created
USER_DELETED         - User account deleted
USER_SUSPENDED       - User account suspended
USER_REACTIVATED     - User account reactivated
ROLE_ASSIGNED        - Role/group assigned to user
ROLE_REMOVED         - Role/group removed from user
BUCKET_CREATED       - New bucket created
BUCKET_DELETED       - Bucket deleted
BUCKET_POLICY_UPDATE - Bucket policy modified
LIFECYCLE_RULE_ADD   - Lifecycle rule added
LIFECYCLE_RULE_DELETE - Lifecycle rule removed
QUOTA_SET            - Storage quota set
QUOTA_EXCEEDED       - Quota limit exceeded
COMPLIANCE_POLICY_UPDATE - Compliance policy changed
```

#### Required Fields:
```json
{
  "admin_event": {
    "admin_user_id": "user_789",
    "admin_email": "admin@example.com",
    "target_user_id": "user_456",
    "target_email": "bob@example.com",
    "policy_before": {...},
    "policy_after": {...},
    "approval_required": true,
    "approval_ticket": "TICKET-5678",
    "approved_by": "user_999"
  }
}
```

---

### 6. **Compartment & Classification**
Security compartment activity.

#### Events to Log:
```
COMPARTMENT_CREATED  - New compartment created
COMPARTMENT_DELETED  - Compartment deleted
COMPARTMENT_POLICY_UPDATE - Compartment policy changed
FILE_COMPARTMENT_ASSIGN - File moved to compartment
FILE_COMPARTMENT_REMOVE - File removed from compartment
COMPARTMENT_ACCESS_GRANT - User granted compartment access
COMPARTMENT_ACCESS_REVOKE - User access revoked
CLASSIFICATION_CHANGE - File classification changed
```

---

### 7. **API & Integration**
API usage and webhook activity.

#### Events to Log:
```
API_REQUEST          - API endpoint called
API_RATE_LIMIT       - Rate limit hit
API_KEY_USED         - API key authentication
WEBHOOK_TRIGGERED    - Webhook fired
WEBHOOK_DELIVERY_SUCCESS - Webhook delivered
WEBHOOK_DELIVERY_FAILURE - Webhook failed
INTEGRATION_ENABLED  - Third-party integration enabled
INTEGRATION_DISABLED - Third-party integration disabled
```

---

## Audit Query Interface

### CLI Commands

#### Basic Queries
```bash
# List recent events
darkstorage audit list --limit 100

# Filter by event type
darkstorage audit list --type FILE_DOWNLOAD

# Filter by user
darkstorage audit list --user alice@example.com

# Filter by file/resource
darkstorage audit list --resource my-bucket/contract.pdf

# Time range
darkstorage audit list --from 2026-02-01 --to 2026-02-28

# Multiple filters
darkstorage audit list \
  --type FILE_DOWNLOAD \
  --user alice@example.com \
  --from 2026-03-01 \
  --limit 50
```

#### Advanced Queries
```bash
# Failed authentication attempts
darkstorage audit list --type LOGIN_FAILURE --limit 100

# Permission changes
darkstorage audit list --category permission

# Security events
darkstorage audit list --category security

# Events from specific IP
darkstorage audit list --ip 203.0.113.45

# Events from specific country
darkstorage audit list --country RU,CN,KP

# High-risk events
darkstorage audit list --risk-score ">75"

# MFA failures
darkstorage audit list --type LOGIN_MFA_FAILURE

# Compartment activity
darkstorage audit list --compartment classified

# Compliance violations
darkstorage audit violations
```

#### File Activity Timeline
```bash
# Complete history for a file
darkstorage audit file my-bucket/contract.pdf

# Who accessed a file
darkstorage audit file my-bucket/contract.pdf --accessed-by

# Permission changes for a file
darkstorage audit file my-bucket/contract.pdf --permission-changes

# Downloads of a file
darkstorage audit file my-bucket/contract.pdf --downloads
```

#### User Activity Reports
```bash
# All activity by a user
darkstorage audit user alice@example.com

# Files accessed by user
darkstorage audit user alice@example.com --files

# Permission grants by user
darkstorage audit user alice@example.com --permissions-granted

# User's login history
darkstorage audit user alice@example.com --logins
```

#### Export & Reports
```bash
# Export to CSV
darkstorage audit export --format csv --output audit-2026-02.csv \
  --from 2026-02-01 --to 2026-02-28

# Export to JSON
darkstorage audit export --format json --output audit.json

# Compliance report
darkstorage audit report --type compliance --month 2026-02

# Security incident report
darkstorage audit report --type security --severity high

# Access report for specific files
darkstorage audit report --type access \
  --files my-bucket/sensitive/ \
  --from 2026-01-01
```

#### Real-time Monitoring
```bash
# Stream audit events (live tail)
darkstorage audit stream

# Stream specific event types
darkstorage audit stream --type FILE_DOWNLOAD,FILE_DELETE

# Stream security events
darkstorage audit stream --category security

# Alert on suspicious activity
darkstorage audit alert \
  --on LOGIN_FAILURE \
  --threshold 5 \
  --window 5m \
  --webhook https://alerts.example.com
```

---

## Retention & Compliance

### Retention Policies
```
Default Retention: 90 days
HIPAA Compliance: 6 years (2,190 days)
Financial Records: 7 years (2,555 days)
Security Events: 1 year (365 days)
Critical Events: Indefinite
```

### Immutability
- Audit logs MUST be write-once, read-many (WORM)
- No deletions or modifications allowed
- Blockchain anchoring for critical events
- Cryptographic signing of all entries

### Access Control
- Only admins and compliance officers can access
- All audit log access is also audited
- Export requires approval for sensitive data
- Real-time alerts for suspicious queries

---

## Example Use Cases

### 1. Compliance Audit (HIPAA)
```bash
# Get all access to patient records
darkstorage audit list \
  --resource-pattern "patients/*" \
  --from 2026-01-01 \
  --to 2026-12-31 \
  --export compliance-2026.csv

# Verify all access had proper authorization
darkstorage audit verify --compartment pii
```

### 2. Security Incident Investigation
```bash
# Find all activity from compromised account
darkstorage audit user bob@example.com --from 2026-02-15

# Find files accessed before password reset
darkstorage audit list \
  --user bob@example.com \
  --type FILE_DOWNLOAD \
  --before 2026-02-15T14:30:00Z

# Check for unusual download patterns
darkstorage audit anomalies --user bob@example.com
```

### 3. Permission Review
```bash
# Who has access to sensitive files
darkstorage audit file my-bucket/classified/secrets.pdf \
  --permission-changes

# Find all permissions granted by former employee
darkstorage audit user exemployee@example.com \
  --permissions-granted \
  --not-revoked
```

### 4. Data Breach Response
```bash
# Find all files downloaded in last 24h
darkstorage audit list \
  --type FILE_DOWNLOAD \
  --since 24h \
  --export breach-investigation.json

# Check for bulk exports
darkstorage audit list \
  --type DATA_EXPORT \
  --user suspicious@example.com
```

---

## Alert & Notification Rules

### Automatic Alerts
```yaml
alerts:
  - name: "Failed Login Attempts"
    condition: "event_type = LOGIN_FAILURE AND count > 5 within 5m"
    severity: high
    notify:
      - email: security@example.com
      - webhook: https://slack.com/hooks/...
      - sms: +1-555-0123

  - name: "Unusual Download Volume"
    condition: "event_type = FILE_DOWNLOAD AND bytes_transferred > 10GB within 1h"
    severity: medium
    notify:
      - email: security@example.com

  - name: "Permission Changes to Sensitive Files"
    condition: "compartment = classified AND category = permission"
    severity: critical
    notify:
      - email: security@example.com
      - email: compliance@example.com
      - pagerduty: incident

  - name: "Integrity Failures"
    condition: "event_type = INTEGRITY_FAILED"
    severity: critical
    notify:
      - email: security@example.com
      - webhook: https://alerts.example.com

  - name: "Access from Blocked Countries"
    condition: "geo_location.country IN (CN, RU, KP, IR)"
    severity: high
    action: block
    notify:
      - email: security@example.com
```

---

## Summary

**YES - ALL actions are logged with:**
- ✅ Who performed the action (user, email, session)
- ✅ What was done (view, edit, delete, permission change)
- ✅ When it happened (precise timestamp)
- ✅ Where it came from (IP, geolocation, device)
- ✅ Why (context, reason, ticket number)
- ✅ Result (success/failure, bytes transferred)
- ✅ Security context (compartment, classification, MFA)
- ✅ Compliance metadata (retention, classification)

**No action goes unrecorded.**
