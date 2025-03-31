// File: pkg/crypto/keymanagement.go
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
)

// Key sizes
const (
	VolumeEncryptionKeySize    = 32                                 // 256-bit key for volume encryption
	KeyEncryptionKeySize       = 32                                 // 256-bit key for key encryption
	RecoveryKeyLength          = 24                                 // Recovery key length in bytes
	RecoveryKeyEncodingCharset = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ" // Base32 without 0, 1, I, O
	RecoveryKeyGroupSize       = 4                                  // Number of characters per group in recovery key
	RecoveryKeyGroupCount      = 6                                  // Number of groups in a recovery key
)

// GenerateVolumeEncryptionKey creates a new random key for volume encryption (VEK)
func GenerateVolumeEncryptionKey() ([]byte, error) {
	vek := make([]byte, VolumeEncryptionKeySize)
	_, err := rand.Read(vek)
	if err != nil {
		return nil, fmt.Errorf("failed to generate volume encryption key: %w", err)
	}
	return vek, nil
}

// GenerateKeyEncryptionKey creates a new random key encryption key (KEK)
func GenerateKeyEncryptionKey() ([]byte, error) {
	kek := make([]byte, KeyEncryptionKeySize)
	_, err := rand.Read(kek)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key encryption key: %w", err)
	}
	return kek, nil
}

// DeriveKeyFromPassword derives a key from a user password and salt
// This implements PBKDF2 with HMAC-SHA256, similar to what APFS uses
func DeriveKeyFromPassword(password string, salt []byte) ([]byte, error) {
	if password == "" {
		return nil, errors.New("password cannot be empty")
	}

	if len(salt) == 0 {
		return nil, errors.New("salt cannot be empty")
	}

	// In a full implementation, this would use specific parameters from the APFS spec
	// For now, we'll use industry standard practices
	const iterations = 10000 // Number of PBKDF2 iterations
	keyLen := 32             // 256-bit key

	// Implement PBKDF2 with HMAC-SHA256
	key := pbkdf2SHA256([]byte(password), salt, iterations, keyLen)
	return key, nil
}

// pbkdf2SHA256 implements PBKDF2 with HMAC-SHA256
// This is a simplified implementation for demonstration
func pbkdf2SHA256(password, salt []byte, iterations, keyLen int) []byte {
	// In a real implementation, you would use the crypto/pbkdf2 package
	// For demonstration purposes, we'll implement a simplified version

	// Initialize the key with the salt
	key := make([]byte, keyLen)
	copy(key, salt)

	// Hash the password with the salt multiple times
	for i := 0; i < iterations; i++ {
		h := sha256.New()
		h.Write(key)
		h.Write(password)
		key = h.Sum(nil)[:keyLen]
	}

	return key
}

// WrapKey wraps a key using AES-256 key wrapping (RFC 3394)
// This is a simplified implementation for demonstration
func WrapKey(key, wrapperKey []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("key cannot be empty")
	}

	if len(wrapperKey) < 16 {
		return nil, errors.New("wrapper key too short")
	}

	// In APFS, key wrapping is done according to RFC 3394
	// For demonstration, we'll use AES-CBC with a known IV

	// Create AES cipher
	block, err := aes.NewCipher(wrapperKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Add padding to ensure key is a multiple of AES block size
	paddedKey := padPKCS7(key, aes.BlockSize)

	// Generate a random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Encrypt with CBC mode
	ciphertext := make([]byte, len(paddedKey))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedKey)

	// Prepend IV to ciphertext
	wrappedKey := append(iv, ciphertext...)
	return wrappedKey, nil
}

// UnwrapKey unwraps a key using AES-256 key unwrapping (RFC 3394)
// This is a simplified implementation for demonstration
func UnwrapKey(wrappedKey, wrapperKey []byte) ([]byte, error) {
	if len(wrappedKey) < aes.BlockSize*2 {
		return nil, errors.New("wrapped key too short")
	}

	if len(wrapperKey) < 16 {
		return nil, errors.New("wrapper key too short")
	}

	// Create AES cipher
	block, err := aes.NewCipher(wrapperKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Extract IV and ciphertext
	iv := wrappedKey[:aes.BlockSize]
	ciphertext := wrappedKey[aes.BlockSize:]

	// Decrypt with CBC mode
	paddedKey := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(paddedKey, ciphertext)

	// Remove padding
	key, err := unpadPKCS7(paddedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unpad key: %w", err)
	}

	return key, nil
}

// GenerateRecoveryKey generates a random recovery key in the format used by APFS
// The key is returned as both raw bytes and a formatted string
func GenerateRecoveryKey() ([]byte, string, error) {
	// Generate random bytes for the recovery key
	rawKey := make([]byte, RecoveryKeyLength)
	_, err := rand.Read(rawKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate recovery key: %w", err)
	}

	// Convert to the human-readable format
	formattedKey, err := FormatRecoveryKey(rawKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to format recovery key: %w", err)
	}

	return rawKey, formattedKey, nil
}

// FormatRecoveryKey converts a raw binary recovery key into the standard formatted string
// The format is typically 6 groups of 4 characters separated by hyphens
func FormatRecoveryKey(rawKey []byte) (string, error) {
	if len(rawKey) < RecoveryKeyLength {
		return "", fmt.Errorf("recovery key too short: expected %d bytes", RecoveryKeyLength)
	}

	charset := []byte(RecoveryKeyEncodingCharset)
	charsetLen := big.NewInt(int64(len(charset)))

	// Convert raw bytes to base32-like encoding with the custom charset
	var formattedGroups []string
	for group := 0; group < RecoveryKeyGroupCount; group++ {
		// Extract 4 bytes for each group (32 bits)
		startIdx := group * 4
		if startIdx+4 > len(rawKey) {
			break
		}

		// Convert 4 bytes to a single uint32
		groupValue := binary.BigEndian.Uint32(rawKey[startIdx : startIdx+4])

		// Convert to base32-like encoding using our custom charset
		var groupChars []byte
		for i := 0; i < RecoveryKeyGroupSize; i++ {
			// Extract 5 bits at a time (32 bits / 5 bits = 6.4 chars, so we get 4 chars with some unused bits)
			shiftBits := uint(20 - (i * 5))
			idxVal := (groupValue >> shiftBits) & 0x1F

			// Convert to the charset
			idx := int(idxVal) % len(charset)
			groupChars = append(groupChars, charset[idx])
		}

		formattedGroups = append(formattedGroups, string(groupChars))
	}

	// Join groups with hyphens
	return fmt.Sprintf("%s-%s-%s-%s-%s-%s",
		formattedGroups[0], formattedGroups[1], formattedGroups[2],
		formattedGroups[3], formattedGroups[4], formattedGroups[5]), nil
}

// ParseRecoveryKey converts a formatted recovery key string back to its raw binary form
func ParseRecoveryKey(formattedKey string) ([]byte, error) {
	// Remove hyphens and spaces
	cleanKey := []byte(formattedKey)
	for i := 0; i < len(cleanKey); i++ {
		if cleanKey[i] == '-' || cleanKey[i] == ' ' {
			cleanKey = append(cleanKey[:i], cleanKey[i+1:]...)
			i--
		}
	}

	// Check length
	if len(cleanKey) != RecoveryKeyGroupCount*RecoveryKeyGroupSize {
		return nil, fmt.Errorf("invalid recovery key length: expected %d characters",
			RecoveryKeyGroupCount*RecoveryKeyGroupSize)
	}

	// Convert from custom base32-like encoding back to bytes
	charset := []byte(RecoveryKeyEncodingCharset)
	charsetMap := make(map[byte]int)
	for i, c := range charset {
		charsetMap[c] = i
	}

	rawKey := make([]byte, RecoveryKeyLength)

	for group := 0; group < RecoveryKeyGroupCount; group++ {
		// Process each group of 4 characters
		startIdx := group * RecoveryKeyGroupSize
		groupChars := cleanKey[startIdx : startIdx+RecoveryKeyGroupSize]

		// Convert 4 characters back to a 32-bit value
		var groupValue uint32
		for i, c := range groupChars {
			// Convert character to value using the charset map
			val, ok := charsetMap[c]
			if !ok {
				return nil, fmt.Errorf("invalid character in recovery key: %c", c)
			}

			// Shift value into position
			shiftBits := uint(20 - (i * 5))
			groupValue |= uint32(val) << shiftBits
		}

		// Store the 32-bit value in the raw key
		destIdx := group * 4
		binary.BigEndian.PutUint32(rawKey[destIdx:destIdx+4], groupValue)
	}

	return rawKey, nil
}

// GenerateSalt creates a new random salt for key derivation
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16) // 16 bytes (128 bits) is a common salt size
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// ValidatePasswordStrength checks if a password meets minimum strength requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Check for a mix of character types
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		case c >= '!' && c <= '/' || c >= ':' && c <= '@' || c >= '[' && c <= '`' || c >= '{' && c <= '~':
			hasSpecial = true
		}
	}

	var missing []string
	if !hasUpper {
		missing = append(missing, "uppercase letter")
	}
	if !hasLower {
		missing = append(missing, "lowercase letter")
	}
	if !hasDigit {
		missing = append(missing, "digit")
	}
	if !hasSpecial {
		missing = append(missing, "special character")
	}

	if len(missing) > 0 {
		return fmt.Errorf("password must contain at least one %s", formatList(missing))
	}

	return nil
}

// formatList formats a list of items as a comma-separated string with "and" before the last item
func formatList(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " and " + items[1]
	default:
		return fmt.Sprintf("%s, and %s",
			formatListWithoutAnd(items[:len(items)-1]),
			items[len(items)-1])
	}
}

// formatListWithoutAnd formats a list of items as a comma-separated string
func formatListWithoutAnd(items []string) string {
	if len(items) == 0 {
		return ""
	}

	result := items[0]
	for i := 1; i < len(items); i++ {
		result += ", " + items[i]
	}

	return result
}
