// File: pkg/crypto/utils.go
package crypto

import (
	"crypto/sha256"
	"fmt"
	"regexp"
)

// ComputeKeyDerivationHash computes a hash suitable for key derivation
func ComputeKeyDerivationHash(input []byte, salt []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("input cannot be empty")
	}

	if len(salt) == 0 {
		return nil, fmt.Errorf("salt cannot be empty")
	}

	// Create a SHA-256 hash
	h := sha256.New()

	// Add salt to the hash
	h.Write(salt)

	// Add input to the hash
	h.Write(input)

	// Return the hash
	return h.Sum(nil), nil
}

// HasPasswordStrength returns a map of password strength characteristics
func HasPasswordStrength(password string) map[string]bool {
	result := make(map[string]bool)

	result["min_length"] = len(password) >= 8
	result["has_uppercase"] = regexp.MustCompile(`[A-Z]`).MatchString(password)
	result["has_lowercase"] = regexp.MustCompile(`[a-z]`).MatchString(password)
	result["has_digit"] = regexp.MustCompile(`[0-9]`).MatchString(password)
	result["has_special"] = regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password)

	return result
}

// IsStrongPassword returns true if the password meets all strength requirements
func IsStrongPassword(password string) bool {
	strength := HasPasswordStrength(password)

	return strength["min_length"] &&
		strength["has_uppercase"] &&
		strength["has_lowercase"] &&
		strength["has_digit"] &&
		strength["has_special"]
}

// GetPasswordStrengthSuggestions returns suggestions for improving password strength
func GetPasswordStrengthSuggestions(password string) []string {
	var suggestions []string
	strength := HasPasswordStrength(password)

	if !strength["min_length"] {
		suggestions = append(suggestions, "Password should be at least 8 characters long")
	}

	if !strength["has_uppercase"] {
		suggestions = append(suggestions, "Include at least one uppercase letter")
	}

	if !strength["has_lowercase"] {
		suggestions = append(suggestions, "Include at least one lowercase letter")
	}

	if !strength["has_digit"] {
		suggestions = append(suggestions, "Include at least one digit")
	}

	if !strength["has_special"] {
		suggestions = append(suggestions, "Include at least one special character")
	}

	return suggestions
}

// BackupEncryptionKeys creates a backup of encryption keys
func BackupEncryptionKeys(keybag *Keybag, backupPassword string) ([]byte, error) {
	if keybag == nil {
		return nil, fmt.Errorf("keybag is nil")
	}

	// Serialize the keybag
	serializedKeybag, err := SerializeKeybag(keybag)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize keybag: %w", err)
	}

	// Generate a salt for the backup password
	salt, err := GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive a key from the backup password
	backupKey, err := DeriveKeyFromPassword(backupPassword, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive backup key: %w", err)
	}

	// Encrypt the serialized keybag with the backup key
	encryptedKeybag, err := EncryptData(serializedKeybag, backupKey, EncryptionModeAESCBC)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt keybag: %w", err)
	}

	// Create a backup structure with salt and encrypted keybag
	backup := append(salt, encryptedKeybag...)

	return backup, nil
}

// RestoreEncryptionKeys restores encryption keys from a backup
func RestoreEncryptionKeys(backupData []byte, backupPassword string) (*Keybag, error) {
	if len(backupData) < 16 {
		return nil, fmt.Errorf("backup data too short")
	}

	// Extract salt from the backup
	salt := backupData[:16]
	encryptedKeybag := backupData[16:]

	// Derive the key from the backup password
	backupKey, err := DeriveKeyFromPassword(backupPassword, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive backup key: %w", err)
	}

	// Decrypt the keybag
	serializedKeybag, err := DecryptData(encryptedKeybag, backupKey, EncryptionModeAESCBC)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt keybag (wrong password?): %w", err)
	}

	// Deserialize the keybag
	keybag, err := DeserializeKeybag(serializedKeybag)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize keybag: %w", err)
	}

	return keybag, nil
}
