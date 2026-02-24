package encryption

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

// KeyGenerator handles generation of encryption keys
type KeyGenerator struct{}

// NewKeyGenerator creates a new key generator
func NewKeyGenerator() *KeyGenerator {
	return &KeyGenerator{}
}

// GenerateKeySet creates a new 3+1 key set
func (kg *KeyGenerator) GenerateKeySet() (*KeySet, error) {
	// Generate active key
	active, err := kg.GenerateKey(KeyTypeActive)
	if err != nil {
		return nil, fmt.Errorf("failed to generate active key: %w", err)
	}

	// Generate 3 backup keys
	backups := make([]*EncryptionKey, 3)
	backupTypes := []KeyType{KeyTypeBackup1, KeyTypeBackup2, KeyTypeBackup3}

	for i, keyType := range backupTypes {
		key, err := kg.GenerateKey(keyType)
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup key %d: %w", i+1, err)
		}
		backups[i] = key
	}

	return &KeySet{
		Active:  active,
		Backups: backups,
	}, nil
}

// GenerateKey creates a single AES-256 key (32 bytes)
func (kg *KeyGenerator) GenerateKey(keyType KeyType) (*EncryptionKey, error) {
	// Generate 32 random bytes for AES-256
	keyData := make([]byte, 32)
	if _, err := rand.Read(keyData); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	// Generate unique ID
	id := kg.generateKeyID(keyType)

	return &EncryptionKey{
		ID:        id,
		Type:      keyType,
		KeyData:   keyData,
		CreatedAt: time.Now(),
		Status:    "active",
	}, nil
}

// generateKeyID creates a unique identifier for the key
func (kg *KeyGenerator) generateKeyID(keyType KeyType) string {
	// Generate 16 random bytes
	b := make([]byte, 16)
	rand.Read(b)

	// Encode and add prefix based on type
	encoded := base64.RawURLEncoding.EncodeToString(b)

	switch keyType {
	case KeyTypeActive:
		return "ek_act_" + encoded
	case KeyTypeBackup1:
		return "ek_bk1_" + encoded
	case KeyTypeBackup2:
		return "ek_bk2_" + encoded
	case KeyTypeBackup3:
		return "ek_bk3_" + encoded
	default:
		return "ek_unk_" + encoded
	}
}

// RotateKeySet rotates the active key to backup and generates a new active key
func (kg *KeyGenerator) RotateKeySet(current *KeySet) (*KeySet, error) {
	// Mark current active key as rotated
	current.Active.Status = "rotated"

	// Generate new active key
	newActive, err := kg.GenerateKey(KeyTypeActive)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new active key: %w", err)
	}

	// Shift backups: active → backup1, backup1 → backup2, backup2 → backup3
	// backup3 is discarded
	newBackups := []*EncryptionKey{
		current.Active,    // Old active becomes backup1
		current.Backups[0], // Old backup1 becomes backup2
		current.Backups[1], // Old backup2 becomes backup3
	}

	// Update types
	newBackups[0].Type = KeyTypeBackup1
	newBackups[1].Type = KeyTypeBackup2
	newBackups[2].Type = KeyTypeBackup3

	return &KeySet{
		Active:  newActive,
		Backups: newBackups,
	}, nil
}

// ExportKeySet exports a key set for secure storage (encrypted with master key)
func (kg *KeyGenerator) ExportKeySet(ks *KeySet) (string, error) {
	// In production, this would encrypt the keys with a master key
	// For now, just base64 encode (NOT SECURE - placeholder)

	// TODO: Implement proper master key encryption
	// This should use a hardware security module or key management service

	return "PLACEHOLDER_ENCRYPTED_KEYSET", nil
}
