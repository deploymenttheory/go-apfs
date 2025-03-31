// Package crypto provides encryption and key management functionality for APFS.
//
// This package implements the cryptographic operations required to read and write
// encrypted APFS volumes, including:
//
// - Volume encryption key (VEK) management
// - Password-based key derivation
// - Recovery key generation and parsing
// - File-level encryption
// - Metadata encryption
// - Key rolling operations
//
// APFS uses AES-XTS for file data encryption and AES-CBC for metadata encryption.
// Keys are typically derived from user passwords using PBKDF2, and stored in keybags
// that are themselves encrypted.
//
// The key hierarchy in APFS consists of:
// 1. User password or recovery key
// 2. Key encryption key (KEK) derived from password/recovery key
// 3. Volume encryption key (VEK) encrypted with the KEK
// 4. File keys (optional) encrypted with the VEK
//
// Basic usage:
//
//	import (
//		"github.com/yourusername/apfs/common"
//		"github.com/yourusername/apfs/crypto"
//	)
//
//	// Create a crypto manager
//	cryptoManager := crypto.NewCryptoManager()
//
//	// Unlock a volume with a password
//	volumeUUID, _ := common.UUIDFromString("11111111-2222-3333-4444-555555555555")
//	vek, err := cryptoManager.UnlockVolumeWithPassword(volumeUUID, "mypassword")
//	if err != nil {
//		panic(err)
//	}
//
//	// Decrypt a file block
//	decryptedBlock, err := cryptoManager.DecryptFileBlock(encryptedBlock, vek, cryptoID, logicalAddr)
//	if err != nil {
//		panic(err)
//	}
//
// The crypto package is designed to work with both the container and filesystem
// packages, providing encryption services for both layers of the APFS stack.
package crypto
