package services

import (
	"crypto/hmac"
	"crypto/sha256"
)

// CryptoService provides cryptographic utilities for APFS
type CryptoService struct{}

// NewCryptoService creates a new crypto service
func NewCryptoService() *CryptoService {
	return &CryptoService{}
}

// Pbkdf2 derives a key from a password using PBKDF2 with SHA-256
// This matches the libfsapfs implementation for APFS encryption key derivation
func (cs *CryptoService) Pbkdf2(password, salt []byte, iterations int, keyLen int) []byte {
	return cs.pbkdf2(password, salt, iterations, keyLen)
}

// pbkdf2 implements PBKDF2-HMAC-SHA256 key derivation per RFC 2898
func (cs *CryptoService) pbkdf2(password, salt []byte, iterations, keyLen int) []byte {
	result := make([]byte, keyLen)
	hashLen := sha256.Size
	blockCount := (keyLen + hashLen - 1) / hashLen

	for block := 1; block <= blockCount; block++ {
		blockStart := (block - 1) * hashLen
		blockEnd := blockStart + hashLen
		if blockEnd > keyLen {
			blockEnd = keyLen
		}

		blockSize := blockEnd - blockStart
		hash := cs.pbkdf2Block(password, salt, iterations, block)
		copy(result[blockStart:blockEnd], hash[:blockSize])
	}

	return result
}

// pbkdf2Block computes a single PBKDF2 block
func (cs *CryptoService) pbkdf2Block(password, salt []byte, iterations, block int) [sha256.Size]byte {
	h := hmac.New(sha256.New, password)
	h.Write(salt)

	// Write block number (4 bytes, big-endian)
	blockBytes := [4]byte{
		byte(block >> 24),
		byte(block >> 16),
		byte(block >> 8),
		byte(block),
	}
	h.Write(blockBytes[:])

	var u [sha256.Size]byte
	h.Sum(u[:0])

	result := u

	for i := 1; i < iterations; i++ {
		h = hmac.New(sha256.New, password)
		h.Write(u[:])
		var uNext [sha256.Size]byte
		h.Sum(uNext[:0])

		for j := 0; j < sha256.Size; j++ {
			result[j] ^= uNext[j]
		}
		u = uNext
	}

	return result
}
