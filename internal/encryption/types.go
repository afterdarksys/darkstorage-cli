package encryption

import (
	"time"
)

// KeyType represents the type of encryption key
type KeyType string

const (
	KeyTypeActive  KeyType = "active"  // Currently used for encryption
	KeyTypeBackup1 KeyType = "backup1" // First backup key
	KeyTypeBackup2 KeyType = "backup2" // Second backup key
	KeyTypeBackup3 KeyType = "backup3" // Third backup key
)

// EncryptionKey represents a single encryption key
type EncryptionKey struct {
	ID        string    `json:"id"`
	Type      KeyType   `json:"type"`
	KeyData   []byte    `json:"-"` // Never expose in JSON
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Status    string    `json:"status"` // active, rotated, revoked
}

// KeySet represents the 3+1 key system
type KeySet struct {
	Active  *EncryptionKey   `json:"active"`
	Backups []*EncryptionKey `json:"backups"` // Should always be 3
}

// EncryptedData represents encrypted data with metadata
type EncryptedData struct {
	KeyID      string `json:"key_id"`       // Which key was used
	Algorithm  string `json:"algorithm"`     // AES-256-GCM
	Nonce      []byte `json:"nonce"`        // IV/nonce
	Ciphertext []byte `json:"ciphertext"`   // Encrypted data
	AuthTag    []byte `json:"auth_tag"`     // Authentication tag (GCM)
}

// RotationPolicy defines when keys should be rotated
type RotationPolicy struct {
	MaxAge      time.Duration `json:"max_age"`       // Rotate after this duration
	MaxOperations int64       `json:"max_operations"` // Rotate after N operations
}

// Default rotation policy: 90 days or 1 million operations
var DefaultRotationPolicy = RotationPolicy{
	MaxAge:        90 * 24 * time.Hour,
	MaxOperations: 1000000,
}
