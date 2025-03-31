// File: pkg/crypto/aes_xts.go
package crypto

import (
	"crypto/aes"
	"encoding/binary"
	"errors"
	"fmt"
)

// encryptAESXTS encrypts data using XTS-AES mode
// For file data, APFS uses AES-XTS with a tweak value derived from
// the file's crypto_id and the logical block address
func encryptAESXTS(data []byte, key []byte, tweak ...uint64) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("data cannot be empty")
	}

	if len(key) < 32 {
		return nil, errors.New("key must be at least 32 bytes for AES-XTS")
	}

	// In XTS mode, the key is split into two halves
	// First half is used for the AES-ECB cipher, second half for the tweak cipher
	encKey := key[:16]     // First half for data encryption
	tweakKey := key[16:32] // Second half for tweak encryption

	// Initialize both AES ciphers
	dataCipher, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create data AES cipher: %w", err)
	}

	tweakCipher, err := aes.NewCipher(tweakKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create tweak AES cipher: %w", err)
	}

	// Create the tweak value from the provided parameters
	var tweakValue uint64
	if len(tweak) > 0 {
		tweakValue = tweak[0]
	}

	// Convert tweak to a block-sized byte array (16 bytes)
	tweakBlock := make([]byte, 16)
	binary.LittleEndian.PutUint64(tweakBlock, tweakValue)

	// Encrypt the tweak with the tweak key using ECB mode
	tweakCipher.Encrypt(tweakBlock, tweakBlock)

	// Ensure data is padded to block size
	blockSize := dataCipher.BlockSize()
	paddedData := padPKCS7(data, blockSize)
	ciphertext := make([]byte, len(paddedData))

	// Process each block
	for i := 0; i < len(paddedData); i += blockSize {
		// Get current data block
		blockStart := i
		blockEnd := i + blockSize
		if blockEnd > len(paddedData) {
			blockEnd = len(paddedData)
		}
		dataBlock := paddedData[blockStart:blockEnd]

		// XOR data with the tweak
		for j := 0; j < len(dataBlock); j++ {
			dataBlock[j] ^= tweakBlock[j]
		}

		// Encrypt the XOR result
		dataCipher.Encrypt(ciphertext[blockStart:blockEnd], dataBlock)

		// XOR the encrypted data with the tweak again
		for j := 0; j < len(dataBlock); j++ {
			ciphertext[blockStart+j] ^= tweakBlock[j]
		}

		// Multiply the tweak by alpha (x) in GF(2^128) for the next block
		galoisMultiply(tweakBlock)
	}

	// Store the tweak value with the ciphertext for decryption
	result := make([]byte, 8+len(ciphertext))
	binary.LittleEndian.PutUint64(result, tweakValue)
	copy(result[8:], ciphertext)

	return result, nil
}

// decryptAESXTS decrypts data using XTS-AES mode
func decryptAESXTS(encryptedData []byte, key []byte, tweak ...uint64) ([]byte, error) {
	if len(encryptedData) < 8+aes.BlockSize {
		return nil, errors.New("encrypted data too short")
	}

	if len(key) < 32 {
		return nil, errors.New("key must be at least 32 bytes for AES-XTS")
	}

	// Extract the tweak value stored with the ciphertext
	storedTweak := binary.LittleEndian.Uint64(encryptedData[:8])

	// Use provided tweak if available, otherwise use the stored value
	tweakValue := storedTweak
	if len(tweak) > 0 {
		tweakValue = tweak[0]
	}

	// Split the key into two halves as in encryption
	encKey := key[:16]
	tweakKey := key[16:32]

	// Initialize both AES ciphers
	dataCipher, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create data AES cipher: %w", err)
	}

	tweakCipher, err := aes.NewCipher(tweakKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create tweak AES cipher: %w", err)
	}

	// Create and encrypt the initial tweak value
	tweakBlock := make([]byte, 16)
	binary.LittleEndian.PutUint64(tweakBlock, tweakValue)
	tweakCipher.Encrypt(tweakBlock, tweakBlock)

	// Get the ciphertext part
	ciphertext := encryptedData[8:]
	blockSize := dataCipher.BlockSize()

	// Create buffer for the plaintext
	plaintext := make([]byte, len(ciphertext))

	// Process each block
	for i := 0; i < len(ciphertext); i += blockSize {
		// Get current ciphertext block
		blockStart := i
		blockEnd := i + blockSize
		if blockEnd > len(ciphertext) {
			blockEnd = len(ciphertext)
		}
		ctBlock := ciphertext[blockStart:blockEnd]

		// XOR ciphertext with the tweak
		for j := 0; j < len(ctBlock); j++ {
			ctBlock[j] ^= tweakBlock[j]
		}

		// Decrypt the XOR result
		dataCipher.Decrypt(plaintext[blockStart:blockEnd], ctBlock)

		// XOR the decrypted data with the tweak again
		for j := 0; j < len(ctBlock); j++ {
			plaintext[blockStart+j] ^= tweakBlock[j]
		}

		// Multiply the tweak by alpha (x) in GF(2^128) for the next block
		galoisMultiply(tweakBlock)
	}

	// Unpad the plaintext
	unpaddedPlaintext, err := unpadPKCS7(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to unpad plaintext: %w", err)
	}

	return unpaddedPlaintext, nil
}

// XTS-AES implementation as specified in IEEE Std 1619-2007
// XTS is a tweakable block cipher that is designed for disk encryption

// galoisMultiply performs Galois Field multiplication of a 128-bit value
// This is used for the tweak value transformation in XTS mode
func galoisMultiply(x []byte) {
	// Galois field multiplication by 2
	// This is equivalent to a left shift and conditional XOR
	carry := (x[0] & 0x80) != 0

	// Shift left by one bit
	for i := 0; i < len(x)-1; i++ {
		x[i] = (x[i] << 1) | (x[i+1] >> 7)
	}
	x[len(x)-1] <<= 1

	// If carry, XOR with the reduction polynomial
	if carry {
		// The reduction polynomial for GF(2^128) is x^128 + x^7 + x^2 + x + 1
		// In little-endian format, we XOR with 0x87
		x[len(x)-1] ^= 0x87
	}
}

// EncryptFileBlockXTS encrypts a single file block using proper AES-XTS
func EncryptFileBlockXTS(blockData []byte, vek []byte, cryptoID uint64, logicalAddr uint64) ([]byte, error) {
	if len(blockData) == 0 {
		return nil, errors.New("block data cannot be empty")
	}

	if len(vek) != 32 {
		return nil, fmt.Errorf("volume encryption key must be 32 bytes, got %d", len(vek))
	}

	// In APFS, the tweak is calculated from the file's crypto_id and the logical block address
	tweak := cryptoID ^ logicalAddr

	return encryptAESXTS(blockData, vek, tweak)
}

// DecryptFileBlockXTS decrypts a single file block using proper AES-XTS
func DecryptFileBlockXTS(encryptedBlock []byte, vek []byte, cryptoID uint64, logicalAddr uint64) ([]byte, error) {
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
