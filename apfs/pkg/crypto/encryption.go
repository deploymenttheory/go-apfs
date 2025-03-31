// File: pkg/crypto/encryption.go
package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/types"
)

// EncryptionMode defines the type of encryption algorithm and mode to use
type EncryptionMode uint8

const (
	// EncryptionModeAESXTS is the standard mode used for file data encryption in APFS
	EncryptionModeAESXTS EncryptionMode = iota

	// EncryptionModeAESCBC is used for keybags and metadata encryption in APFS
	EncryptionModeAESCBC
)

// EncryptData encrypts data using the specified key and mode
// For file data, APFS uses AES-XTS with a tweak value derived from
// the file's crypto_id and the logical block address
func EncryptData(data []byte, key []byte, mode EncryptionMode, tweak ...uint64) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("data cannot be empty")
	}

	if len(key) < 16 {
		return nil, errors.New("key too short")
	}

	switch mode {
	case EncryptionModeAESXTS:
		return encryptAESXTS(data, key, tweak...)

	case EncryptionModeAESCBC:
		return encryptAESCBC(data, key)

	default:
		return nil, fmt.Errorf("unsupported encryption mode: %d", mode)
	}
}

// DecryptData decrypts data using the specified key and mode
func DecryptData(encryptedData []byte, key []byte, mode EncryptionMode, tweak ...uint64) ([]byte, error) {
	if len(encryptedData) == 0 {
		return nil, errors.New("encrypted data cannot be empty")
	}

	if len(key) < 16 {
		return nil, errors.New("key too short")
	}

	switch mode {
	case EncryptionModeAESXTS:
		return decryptAESXTS(encryptedData, key, tweak...)

	case EncryptionModeAESCBC:
		return decryptAESCBC(encryptedData, key)

	default:
		return nil, fmt.Errorf("unsupported encryption mode: %d", mode)
	}
}

// EncryptMetadata encrypts file system metadata using the volume encryption key (VEK)
// In APFS, metadata is typically encrypted using AES-CBC
func EncryptMetadata(metadata []byte, vek []byte) ([]byte, error) {
	if len(metadata) == 0 {
		return nil, errors.New("metadata cannot be empty")
	}

	if len(vek) < 16 {
		return nil, errors.New("volume encryption key too short")
	}

	// Encrypt using AES-CBC
	return encryptAESCBC(metadata, vek)
}

// DecryptMetadata decrypts file system metadata using the volume encryption key (VEK)
func DecryptMetadata(encryptedMetadata []byte, vek []byte) ([]byte, error) {
	if len(encryptedMetadata) == 0 {
		return nil, errors.New("encrypted metadata cannot be empty")
	}

	if len(vek) < 16 {
		return nil, errors.New("volume encryption key too short")
	}

	// Decrypt using AES-CBC
	return decryptAESCBC(encryptedMetadata, vek)
}

// EncryptFileBlock encrypts a single file block using AES-XTS
// The tweak value is derived from the file's crypto_id and the logical block address
func EncryptFileBlock(blockData []byte, vek []byte, cryptoID uint64, logicalAddr uint64) ([]byte, error) {
	if len(blockData) == 0 {
		return nil, errors.New("block data cannot be empty")
	}

	if len(vek) != 32 {
		return nil, fmt.Errorf("volume encryption key must be 32 bytes, got %d", len(vek))
	}

	// In APFS, the tweak is calculated from the file's crypto_id and the logical block address
	// For demonstration, we'll use both values combined
	tweak := cryptoID ^ logicalAddr

	return encryptAESXTS(blockData, vek, tweak)
}

// DecryptFileBlock decrypts a single file block using AES-XTS
// The tweak value is derived from the file's crypto_id and the logical block address
func DecryptFileBlock(encryptedBlock []byte, vek []byte, cryptoID uint64, logicalAddr uint64) ([]byte, error) {
	if len(encryptedBlock) == 0 {
		return nil, errors.New("encrypted block data cannot be empty")
	}

	if len(vek) != 32 {
		return nil, fmt.Errorf("volume encryption key must be 32 bytes, got %d", len(vek))
	}

	// Calculate the tweak as in encryption
	tweak := cryptoID ^ logicalAddr

	return decryptAESXTS(encryptedBlock, vek, tweak)
}

// Private helper functions

// encryptAESXTS encrypts data using AES-XTS mode
// This is a simplified implementation for demonstration purposes
func encryptAESXTS(data []byte, key []byte, tweak ...uint64) ([]byte, error) {
	// In a full implementation, this would use a proper XTS implementation
	// For demonstration purposes, we'll use a simplified approach

	// In APFS, AES-XTS uses two keys: one for encryption and one for the tweak
	// The full key is split into two halves
	if len(key) < 32 {
		return nil, errors.New("key must be at least 32 bytes for AES-XTS")
	}

	// Pad data to AES block size
	paddedData := padPKCS7(data, aes.BlockSize)

	// Create a seed from the tweak
	var tweakValue uint64
	if len(tweak) > 0 {
		tweakValue = tweak[0]
	}

	// Simulate XTS mode by encrypting with AES-CBC using a tweak-derived IV
	encKey := key[:16]     // First half for encryption
	tweakKey := key[16:32] // Second half for tweak

	// Create an IV based on the tweak key and tweak value
	iv := make([]byte, aes.BlockSize)
	for i := 0; i < 8; i++ {
		iv[i] = tweakKey[i] ^ byte(tweakValue>>(i*8))
	}

	// Encrypt using AES-CBC with the derived IV
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	ciphertext := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedData)

	// Store the tweak value with the ciphertext
	result := make([]byte, 8+len(ciphertext))
	for i := 0; i < 8; i++ {
		result[i] = byte(tweakValue >> (i * 8))
	}
	copy(result[8:], ciphertext)

	return result, nil
}

// decryptAESXTS decrypts data using AES-XTS mode
func decryptAESXTS(encryptedData []byte, key []byte, tweak ...uint64) ([]byte, error) {
	if len(encryptedData) < 8+aes.BlockSize {
		return nil, errors.New("encrypted data too short")
	}

	if len(key) < 32 {
		return nil, errors.New("key must be at least 32 bytes for AES-XTS")
	}

	// Extract the tweak value
	var storedTweak uint64
	for i := 0; i < 8; i++ {
		storedTweak |= uint64(encryptedData[i]) << (i * 8)
	}

	// Use provided tweak if available, otherwise use the stored value
	tweakValue := storedTweak
	if len(tweak) > 0 {
		tweakValue = tweak[0]
	}

	// Split the key
	encKey := key[:16]     // First half for encryption
	tweakKey := key[16:32] // Second half for tweak

	// Create an IV based on the tweak key and tweak value
	iv := make([]byte, aes.BlockSize)
	for i := 0; i < 8; i++ {
		iv[i] = tweakKey[i] ^ byte(tweakValue>>(i*8))
	}

	// Decrypt using AES-CBC with the derived IV
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	ciphertext := encryptedData[8:]
	paddedPlaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(paddedPlaintext, ciphertext)

	// Unpad the plaintext
	plaintext, err := unpadPKCS7(paddedPlaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to unpad plaintext: %w", err)
	}

	return plaintext, nil
}

// encryptAESCBC encrypts data using AES-CBC mode
func encryptAESCBC(data []byte, key []byte) ([]byte, error) {
	// Pad data to AES block size
	paddedData := padPKCS7(data, aes.BlockSize)

	// Generate a random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Encrypt with CBC mode
	ciphertext := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedData)

	// Prepend IV to ciphertext
	result := append(iv, ciphertext...)
	return result, nil
}

// decryptAESCBC decrypts data using AES-CBC mode
func decryptAESCBC(encryptedData []byte, key []byte) ([]byte, error) {
	if len(encryptedData) < aes.BlockSize*2 {
		return nil, errors.New("encrypted data too short")
	}

	// Extract IV and ciphertext
	iv := encryptedData[:aes.BlockSize]
	ciphertext := encryptedData[aes.BlockSize:]

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Decrypt with CBC mode
	paddedPlaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(paddedPlaintext, ciphertext)

	// Unpad the plaintext
	plaintext, err := unpadPKCS7(paddedPlaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to unpad plaintext: %w", err)
	}

	return plaintext, nil
}

// padPKCS7 adds PKCS#7 padding to data
func padPKCS7(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// unpadPKCS7 removes PKCS#7 padding from data
func unpadPKCS7(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("empty data")
	}

	padLength := int(data[length-1])
	if padLength > length {
		return nil, errors.New("invalid padding length")
	}

	// Verify padding
	for i := length - padLength; i < length; i++ {
		if data[i] != byte(padLength) {
			return nil, errors.New("invalid padding")
		}
	}

	return data[:length-padLength], nil
}

// EncryptFileExtent encrypts an entire file extent
// This would be used for file data encryption in APFS
func EncryptFileExtent(extentData []byte, vek []byte, fileExtent *types.JFileExtentVal, blockSize uint32) ([]byte, error) {
	if len(extentData) == 0 {
		return nil, errors.New("extent data cannot be empty")
	}

	if fileExtent == nil {
		return nil, errors.New("file extent information is nil")
	}

	if blockSize == 0 {
		return nil, errors.New("block size cannot be zero")
	}

	// Get the crypto ID from the file extent
	cryptoID := fileExtent.CryptoID

	// Encrypt each block in the extent
	encryptedData := make([]byte, len(extentData))

	for blockOffset := 0; blockOffset < len(extentData); blockOffset += int(blockSize) {
		// Calculate end of current block (or end of data)
		endOffset := blockOffset + int(blockSize)
		if endOffset > len(extentData) {
			endOffset = len(extentData)
		}

		// Calculate logical block address relative to the start of the extent
		logicalAddr := uint64(blockOffset / int(blockSize))

		// Get the block data
		blockData := extentData[blockOffset:endOffset]

		// Encrypt the block
		encryptedBlock, err := EncryptFileBlock(blockData, vek, cryptoID, logicalAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt block at offset %d: %w", blockOffset, err)
		}

		// Copy the encrypted block to the result
		copy(encryptedData[blockOffset:], encryptedBlock)
	}

	return encryptedData, nil
}

// DecryptFileExtent decrypts an entire file extent
func DecryptFileExtent(encryptedExtent []byte, vek []byte, fileExtent *types.JFileExtentVal, blockSize uint32) ([]byte, error) {
	if len(encryptedExtent) == 0 {
		return nil, errors.New("encrypted extent data cannot be empty")
	}

	if fileExtent == nil {
		return nil, errors.New("file extent information is nil")
	}

	if blockSize == 0 {
		return nil, errors.New("block size cannot be zero")
	}

	// Get the crypto ID from the file extent
	cryptoID := fileExtent.CryptoID

	// Decrypt each block in the extent
	decryptedData := make([]byte, len(encryptedExtent))

	for blockOffset := 0; blockOffset < len(encryptedExtent); blockOffset += int(blockSize) {
		// Calculate end of current block (or end of data)
		endOffset := blockOffset + int(blockSize)
		if endOffset > len(encryptedExtent) {
			endOffset = len(encryptedExtent)
		}

		// Calculate logical block address relative to the start of the extent
		logicalAddr := uint64(blockOffset / int(blockSize))

		// Get the encrypted block data
		encryptedBlock := encryptedExtent[blockOffset:endOffset]

		// Decrypt the block
		decryptedBlock, err := DecryptFileBlock(encryptedBlock, vek, cryptoID, logicalAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt block at offset %d: %w", blockOffset, err)
		}

		// Copy the decrypted block to the result
		copy(decryptedData[blockOffset:], decryptedBlock)
	}

	return decryptedData, nil
}
