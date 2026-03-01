# DarkStorage Authentication Architecture

## Overview

**Question:** How does OAuth2 (via Authentik) authentication work with MinIO S3-compatible storage?

**Answer:** DarkStorage uses a **backend API gateway** that sits between the CLI/Console and MinIO. The gateway handles OAuth2 authentication and translates it into MinIO credentials using **temporary STS (Security Token Service) credentials** or **pre-signed URLs**.

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           USER AUTHENTICATION                            │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    │               │               │
                    ▼               ▼               ▼
           ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
           │     CLI      │  │ Web Console │  │   Mobile    │
           │  (Go binary) │  │  (Next.js)  │  │    App      │
           └─────────────┘  └─────────────┘  └─────────────┘
                    │               │               │
                    └───────────────┼───────────────┘
                                    ▼
                        ┌─────────────────────┐
                        │   1. OAuth2 Login   │
                        │    via Authentik    │
                        └─────────────────────┘
                                    │
                                    ▼
                        ┌─────────────────────┐
                        │  2. Get JWT Token   │
                        │  (access_token)     │
                        └─────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         DARKSTORAGE API GATEWAY                          │
│                           (Backend Middleware)                           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │  3. Receive Request with JWT Bearer Token                      │    │
│  │     Authorization: Bearer eyJhbGciOiJSUzI1NiIs...              │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                              │                                          │
│                              ▼                                          │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │  4. Validate JWT with Authentik                                │    │
│  │     - Verify signature                                          │    │
│  │     - Check expiration                                          │    │
│  │     - Extract user info (email, groups, roles)                 │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                              │                                          │
│                              ▼                                          │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │  5. Load User Permissions from Database                        │    │
│  │     - File permissions                                          │    │
│  │     - Bucket policies                                           │    │
│  │     - Compartment access                                        │    │
│  │     - Group memberships                                         │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                              │                                          │
│                              ▼                                          │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │  6. Generate MinIO Credentials                                 │    │
│  │     OPTION A: Use STS (AssumeRole)                             │    │
│  │     OPTION B: Generate Pre-signed URLs                         │    │
│  │     OPTION C: Use service account with bucket policies         │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                              │                                          │
└──────────────────────────────┼──────────────────────────────────────────┘
                               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                            MINIO S3 STORAGE                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │  7. Execute S3 Operation                                        │    │
│  │     - PUT /bucket/file.txt                                      │    │
│  │     - GET /bucket/file.txt                                      │    │
│  │     - DELETE /bucket/file.txt                                   │    │
│  │     - LIST /bucket/                                             │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                              │                                          │
│                              ▼                                          │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │  8. Verify MinIO Credentials                                   │    │
│  │     - Check STS token validity                                  │    │
│  │     - Verify bucket policies                                    │    │
│  │     - Check IAM permissions                                     │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                              │                                          │
│                              ▼                                          │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │  9. Return File Data or Status                                 │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │  10. Return to User │
                    └─────────────────────┘
```

---

## Three Options for MinIO Authentication

### **Option 1: MinIO STS (Security Token Service)** ⭐ RECOMMENDED

This is the most secure and scalable approach.

#### How It Works:

1. **User logs in via Authentik OAuth2**
   - CLI/Console gets JWT access token
   - Token contains user identity (email, groups, roles)

2. **Backend API Gateway receives request**
   ```
   GET /files/my-bucket/file.txt
   Authorization: Bearer eyJhbGciOiJSUzI1NiIs...
   ```

3. **Gateway validates JWT with Authentik**
   - Verify signature using Authentik's public key
   - Check token hasn't expired
   - Extract user claims (email, groups)

4. **Gateway requests STS credentials from MinIO**
   ```go
   // Backend code (Go)
   import "github.com/minio/minio-go/v7/pkg/credentials"

   // Use AssumeRoleWithWebIdentity
   stsCredentials, err := credentials.NewSTSAssumeRole(
       minioEndpoint,
       credentials.STSAssumeRoleOptions{
           AccessKey:       minioRootUser,     // Service account
           SecretKey:       minioRootPassword,
           SessionToken:    jwtToken,          // Authentik JWT
           DurationSeconds: 3600,              // 1 hour
           Policy:          userPolicy,        // Dynamic policy based on permissions
       },
   )
   ```

5. **MinIO returns temporary credentials**
   ```json
   {
     "accessKeyId": "temp-access-key-abc123",
     "secretAccessKey": "temp-secret-key-xyz789",
     "sessionToken": "session-token-...",
     "expiration": "2026-03-01T15:30:00Z"
   }
   ```

6. **Gateway uses temp credentials to access MinIO**
   ```go
   minioClient, err := minio.New(minioEndpoint, &minio.Options{
       Creds:  stsCredentials,
       Secure: true,
   })

   // Now perform S3 operations with user-specific permissions
   object, err := minioClient.GetObject(ctx, "my-bucket", "file.txt", minio.GetObjectOptions{})
   ```

7. **Cache STS credentials** (until expiration)
   - Store in Redis with user session key
   - Refresh automatically when expired
   - Invalidate on logout

#### Benefits:
- ✅ Temporary credentials (auto-expire)
- ✅ No static access keys
- ✅ Per-user access control
- ✅ Audit trail (who accessed what)
- ✅ Can revoke access instantly

---

### **Option 2: Pre-signed URLs**

Generate temporary URLs for direct client access.

#### How It Works:

1. **User requests file**
   ```
   GET /api/files/my-bucket/file.txt
   Authorization: Bearer <jwt>
   ```

2. **Gateway validates JWT and permissions**

3. **Gateway generates pre-signed URL from MinIO**
   ```go
   // Backend code
   presignedURL, err := minioClient.PresignedGetObject(ctx,
       "my-bucket",
       "file.txt",
       time.Hour,  // Valid for 1 hour
       url.Values{},
   )

   // Returns: https://minio.example.com/my-bucket/file.txt?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...
   ```

4. **Client downloads directly from MinIO**
   ```bash
   # CLI gets pre-signed URL and downloads
   curl -o file.txt "https://minio.example.com/my-bucket/file.txt?X-Amz-Algorithm=..."
   ```

#### Benefits:
- ✅ Direct download (no proxy overhead)
- ✅ Faster for large files
- ✅ CDN-friendly
- ❌ Can't enforce real-time permission changes
- ❌ URLs can be shared (until expiry)

---

### **Option 3: Backend Proxy** (Simplest but less scalable)

Gateway acts as full proxy to MinIO.

#### How It Works:

1. **User requests file**
   ```
   GET /api/files/my-bucket/file.txt
   Authorization: Bearer <jwt>
   ```

2. **Gateway validates JWT**

3. **Gateway uses service account to access MinIO**
   ```go
   // Backend uses single service account
   minioClient, err := minio.New(minioEndpoint, &minio.Options{
       Creds:  credentials.NewStaticV4(serviceAccessKey, serviceSecretKey, ""),
       Secure: true,
   })
   ```

4. **Gateway checks permissions in database**
   ```sql
   SELECT permission FROM file_permissions
   WHERE file_path = 'my-bucket/file.txt'
   AND user_id = $1
   ```

5. **Gateway proxies request to MinIO**
   ```go
   object, err := minioClient.GetObject(ctx, "my-bucket", "file.txt", minio.GetObjectOptions{})

   // Stream to client
   io.Copy(responseWriter, object)
   ```

#### Benefits:
- ✅ Simple to implement
- ✅ Full control over permissions
- ✅ Can enforce real-time policy changes
- ❌ Gateway is bottleneck for large files
- ❌ Higher bandwidth costs

---

## Recommended Architecture for DarkStorage

### **Hybrid Approach** (Best of All Worlds)

```yaml
authentication_flow:
  step1:
    - User logs in via Authentik OAuth2
    - CLI/Console stores JWT access token
    - Token refresh handled automatically

  step2:
    - Every API request includes JWT
    - Gateway validates with Authentik
    - User identity extracted from JWT

  step3_permissions:
    - Load user permissions from PostgreSQL
    - Check file permissions
    - Check compartment access
    - Check bucket policies
    - Evaluate global policies

  step4_minio_access:
    small_files: # < 10 MB
      method: proxy
      reason: "Full control, audit every access"

    large_files: # > 10 MB
      method: presigned_url
      reason: "Direct download, faster"

    admin_operations:
      method: sts_credentials
      reason: "Temporary credentials, better security"
```

---

## Implementation Details

### Backend API Gateway (Go)

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

// Middleware to validate Authentik JWT
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get JWT from Authorization header
        token := c.GetHeader("Authorization")
        token = strings.TrimPrefix(token, "Bearer ")

        // Validate JWT with Authentik
        claims, err := validateAuthentikJWT(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }

        // Store user info in context
        c.Set("user_email", claims.Email)
        c.Set("user_id", claims.Sub)
        c.Next()
    }
}

// Download file handler
func DownloadFile(c *gin.Context) {
    bucket := c.Param("bucket")
    path := c.Param("path")
    userID := c.GetString("user_id")

    // Check permissions in database
    hasPermission, err := checkFilePermission(userID, bucket, path, "read")
    if err != nil || !hasPermission {
        c.JSON(403, gin.H{"error": "Access denied"})
        return
    }

    // Get or create STS credentials for this user
    stsCredentials, err := getSTSCredentials(userID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to get credentials"})
        return
    }

    // Create MinIO client with STS credentials
    minioClient, err := minio.New(minioEndpoint, &minio.Options{
        Creds:  stsCredentials,
        Secure: true,
    })

    // Download file
    object, err := minioClient.GetObject(c.Request.Context(), bucket, path, minio.GetObjectOptions{})
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to download"})
        return
    }
    defer object.Close()

    // Log audit event
    logAuditEvent(userID, "FILE_DOWNLOAD", bucket+"/"+path)

    // Stream to client
    c.Stream(func(w io.Writer) bool {
        _, err := io.Copy(w, object)
        return err == nil
    })
}
```

### CLI OAuth Flow

```go
// cmd/login.go (existing)

func performLogin() error {
    // 1. Start local callback server
    server := startCallbackServer()

    // 2. Open browser to Authentik login
    authURL := fmt.Sprintf("%s/auth/login?client_id=%s&redirect_uri=%s",
        consoleURL, clientID, callbackURL)
    openBrowser(authURL)

    // 3. Wait for callback with access token
    token := <-server.TokenChannel

    // 4. Save token to config file
    config := Config{
        AccessToken:  token.AccessToken,
        RefreshToken: token.RefreshToken,
        ExpiresAt:    time.Now().Add(time.Hour),
    }
    saveConfig(config)

    // 5. Test token by calling API
    client := newAPIClient(token.AccessToken)
    user, err := client.GetCurrentUser()
    if err != nil {
        return err
    }

    fmt.Printf("Logged in as: %s\n", user.Email)
    return nil
}
```

### Storage Backend Configuration

```go
// internal/storage/traditional.go

type TraditionalBackend struct {
    client *minio.Client
    config *TraditionalConfig
}

func NewTraditionalBackend(cfg *TraditionalConfig) (*TraditionalBackend, error) {
    // This is used for backend API, NOT direct CLI access
    client, err := minio.New(cfg.Endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
        Secure: cfg.UseSSL,
        Region: cfg.Region,
    })

    return &TraditionalBackend{
        client: client,
        config: cfg,
    }, nil
}
```

---

## Database Schema for Permissions

```sql
-- Store user permissions
CREATE TABLE file_permissions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    file_path TEXT NOT NULL,
    permission VARCHAR(20) NOT NULL,  -- read, write, delete
    granted_by UUID,
    granted_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    UNIQUE(user_id, file_path, permission)
);

-- Store STS credentials cache
CREATE TABLE sts_credentials_cache (
    user_id UUID PRIMARY KEY,
    access_key_id VARCHAR(255) NOT NULL,
    secret_access_key VARCHAR(255) NOT NULL,
    session_token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Bucket policies
CREATE TABLE bucket_policies (
    bucket_name VARCHAR(255) PRIMARY KEY,
    policy JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

---

## Summary

**How OAuth2 + MinIO Works:**

1. ✅ **User authenticates** via Authentik OAuth2 (web browser)
2. ✅ **CLI/Console gets** JWT access token
3. ✅ **Every request includes** JWT in Authorization header
4. ✅ **Backend API Gateway** validates JWT with Authentik
5. ✅ **Gateway generates** temporary MinIO STS credentials per user
6. ✅ **Gateway uses** STS credentials to access MinIO on behalf of user
7. ✅ **Permissions checked** in database before each operation
8. ✅ **Audit logs** record who accessed what
9. ✅ **STS credentials** auto-expire (security)
10. ✅ **MinIO never sees** Authentik tokens directly

**Key Points:**
- MinIO uses STS (temporary credentials)
- Backend API Gateway bridges OAuth2 ↔ MinIO
- Permissions stored in PostgreSQL
- Every action is audited
- Credentials auto-expire for security
