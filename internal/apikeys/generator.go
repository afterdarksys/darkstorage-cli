package apikeys

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Generator handles API key generation
type Generator struct {
	environment string // "live" or "test"
}

// NewGenerator creates a new API key generator
func NewGenerator(env string) *Generator {
	if env != "live" && env != "test" {
		env = "live"
	}
	return &Generator{environment: env}
}

// Generate creates a new API key with all metadata
func (g *Generator) Generate(req *CreateAPIKeyRequest, userID string) (*APIKey, string, error) {
	// Generate the actual API key
	apiKey, err := g.generateAPIKey()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate API key: %w", err)
	}

	// Hash the key for storage
	keyHash, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash API key: %w", err)
	}

	// Generate S3 credentials
	s3AccessKey, s3SecretKey, err := g.generateS3Credentials()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate S3 credentials: %w", err)
	}

	// Calculate expiration
	var expiresAt *time.Time
	if req.ExpiresInDays > 0 {
		expiry := time.Now().Add(time.Duration(req.ExpiresInDays) * 24 * time.Hour)
		expiresAt = &expiry
	}

	// Create API key object
	key := &APIKey{
		ID:             generateID(),
		UserID:         userID,
		Name:           req.Name,
		KeyPrefix:      g.getPrefix() + apiKey[:20], // First 20 chars for display
		KeyHash:        string(keyHash),
		Permissions:    req.Permissions,
		IPRestrictions: req.IPRestrictions,
		Status:         StatusActive,
		CreatedAt:      time.Now(),
		ExpiresAt:      expiresAt,
		Metadata:       req.Metadata,
		S3AccessKey:    s3AccessKey,
		S3SecretKey:    s3SecretKey,
	}

	return key, apiKey, nil
}

// generateAPIKey creates a cryptographically secure random API key
func (g *Generator) generateAPIKey() (string, error) {
	// Generate 48 random bytes (will be 64 chars when base64 encoded)
	b := make([]byte, 48)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Encode to base64 and remove padding
	encoded := base64.RawURLEncoding.EncodeToString(b)

	// Add prefix
	prefix := g.getPrefix()
	return prefix + encoded, nil
}

// generateS3Credentials creates S3-compatible access and secret keys
func (g *Generator) generateS3Credentials() (string, string, error) {
	// Access Key: DKSA + 16 random chars
	accessBytes := make([]byte, 12)
	if _, err := rand.Read(accessBytes); err != nil {
		return "", "", err
	}
	accessKey := "DKSA" + base64.RawURLEncoding.EncodeToString(accessBytes)

	// Secret Key: 40 random chars
	secretBytes := make([]byte, 30)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", "", err
	}
	secretKey := base64.RawURLEncoding.EncodeToString(secretBytes)

	return accessKey, secretKey, nil
}

// getPrefix returns the appropriate key prefix based on environment
func (g *Generator) getPrefix() string {
	if g.environment == "test" {
		return PrefixTest
	}
	return PrefixLive
}

// generateID creates a unique ID for the API key
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "key_" + base64.RawURLEncoding.EncodeToString(b)
}

// ValidateKey validates an API key format
func ValidateKey(key string) error {
	if !strings.HasPrefix(key, PrefixLive) && !strings.HasPrefix(key, PrefixTest) {
		return fmt.Errorf("invalid key prefix")
	}

	if len(key) < 70 { // prefix (8) + encoded data (64+)
		return fmt.Errorf("key too short")
	}

	return nil
}

// VerifyKey compares a plain API key with its hash
func VerifyKey(plainKey, hashedKey string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedKey), []byte(plainKey))
}
