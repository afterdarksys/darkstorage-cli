package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// Encryptor handles AES-256-GCM encryption
type Encryptor struct {
	keySet *KeySet
}

// NewEncryptor creates a new encryptor with the given key set
func NewEncryptor(keySet *KeySet) *Encryptor {
	return &Encryptor{keySet: keySet}
}

// Encrypt encrypts data using the active key
func (e *Encryptor) Encrypt(plaintext []byte) (*EncryptedData, error) {
	if e.keySet.Active == nil {
		return nil, fmt.Errorf("no active encryption key")
	}

	// Create AES cipher
	block, err := aes.NewCipher(e.keySet.Active.KeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	return &EncryptedData{
		KeyID:      e.keySet.Active.ID,
		Algorithm:  "AES-256-GCM",
		Nonce:      nonce,
		Ciphertext: ciphertext,
		AuthTag:    nil, // GCM includes auth tag in ciphertext
	}, nil
}

// Decrypt decrypts data using the appropriate key from the key set
func (e *Encryptor) Decrypt(data *EncryptedData) ([]byte, error) {
	// Find the key that was used
	var key *EncryptionKey

	if e.keySet.Active.ID == data.KeyID {
		key = e.keySet.Active
	} else {
		// Check backup keys
		for _, backup := range e.keySet.Backups {
			if backup.ID == data.KeyID {
				key = backup
				break
			}
		}
	}

	if key == nil {
		return nil, fmt.Errorf("encryption key %s not found", data.KeyID)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key.KeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, data.Nonce, data.Ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// EncryptStream encrypts a stream of data (for large files)
func (e *Encryptor) EncryptStream(reader io.Reader, writer io.Writer) (*EncryptedData, error) {
	if e.keySet.Active == nil {
		return nil, fmt.Errorf("no active encryption key")
	}

	// Create AES cipher
	block, err := aes.NewCipher(e.keySet.Active.KeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create CTR mode for streaming
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	stream := cipher.NewCTR(block, iv)
	streamWriter := &cipher.StreamWriter{S: stream, W: writer}

	// Copy and encrypt
	if _, err := io.Copy(streamWriter, reader); err != nil {
		return nil, fmt.Errorf("stream encryption failed: %w", err)
	}

	return &EncryptedData{
		KeyID:     e.keySet.Active.ID,
		Algorithm: "AES-256-CTR",
		Nonce:     iv,
	}, nil
}

// DecryptStream decrypts a stream of data
func (e *Encryptor) DecryptStream(reader io.Reader, writer io.Writer, data *EncryptedData) error {
	// Find the key
	var key *EncryptionKey

	if e.keySet.Active.ID == data.KeyID {
		key = e.keySet.Active
	} else {
		for _, backup := range e.keySet.Backups {
			if backup.ID == data.KeyID {
				key = backup
				break
			}
		}
	}

	if key == nil {
		return fmt.Errorf("encryption key %s not found", data.KeyID)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key.KeyData)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create CTR mode for streaming
	stream := cipher.NewCTR(block, data.Nonce)
	streamReader := &cipher.StreamReader{S: stream, R: reader}

	// Copy and decrypt
	if _, err := io.Copy(writer, streamReader); err != nil {
		return fmt.Errorf("stream decryption failed: %w", err)
	}

	return nil
}
