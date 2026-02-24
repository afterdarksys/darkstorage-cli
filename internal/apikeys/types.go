package apikeys

import (
	"time"
)

// APIKey represents an API key with metadata
type APIKey struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	KeyPrefix   string    `json:"key_prefix"`   // "dk_live_abc123..." for display
	KeyHash     string    `json:"-"`            // bcrypt hash, never exposed
	Permissions []string  `json:"permissions"`
	IPRestrictions []string `json:"ip_restrictions,omitempty"`
	Status      string    `json:"status"`       // active, revoked, expired
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	LastUsedIP  string    `json:"last_used_ip,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`

	// S3 credentials (only populated on creation)
	S3AccessKey string `json:"s3_access_key,omitempty"`
	S3SecretKey string `json:"s3_secret_key,omitempty"`
}

// CreateAPIKeyRequest represents request to create new API key
type CreateAPIKeyRequest struct {
	Name           string            `json:"name"`
	Permissions    []string          `json:"permissions"`
	ExpiresInDays  int               `json:"expires_in_days,omitempty"`
	IPRestrictions []string          `json:"ip_restrictions,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// CreateAPIKeyResponse represents the response when creating an API key
type CreateAPIKeyResponse struct {
	Key         *APIKey `json:"key"`
	APIKeyValue string  `json:"api_key"` // Full key, only shown once
	Warning     string  `json:"warning"`
}

// ListAPIKeysResponse represents list of API keys
type ListAPIKeysResponse struct {
	Keys  []*APIKey `json:"keys"`
	Total int       `json:"total"`
}

// Permission constants
const (
	PermStorageRead   = "storage:read"
	PermStorageWrite  = "storage:write"
	PermStorageDelete = "storage:delete"
	PermBucketCreate  = "bucket:create"
	PermBucketDelete  = "bucket:delete"
	PermShareCreate   = "share:create"
	PermAdminAll      = "admin:*"
)

// Key status constants
const (
	StatusActive  = "active"
	StatusRevoked = "revoked"
	StatusExpired = "expired"
)

// Key prefix constants
const (
	PrefixLive = "dk_live_"
	PrefixTest = "dk_test_"
)

// AllPermissions returns all available permissions
func AllPermissions() []string {
	return []string{
		PermStorageRead,
		PermStorageWrite,
		PermStorageDelete,
		PermBucketCreate,
		PermBucketDelete,
		PermShareCreate,
		PermAdminAll,
	}
}
