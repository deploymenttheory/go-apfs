package services

import (
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// EncryptionHelper provides encryption-related utilities for APFS
// This includes key derivation and tweak calculation for XTS-AES-128 encryption.
type EncryptionHelper struct{}

// NewEncryptionHelper creates a new encryption helper
func NewEncryptionHelper() *EncryptionHelper {
	return &EncryptionHelper{}
}

// DeriveContainerKeybagKey derives the 256-bit XTS-AES-128 encryption key for a container keybag
// by concatenating the container UUID with itself.
//
// Reference: APFS Advent Challenge Day 15
// Formula: container_keybag_key = container_uuid + container_uuid
//
// This key is used to decrypt the container keybag, which stores the location of each
// encrypted volume's keybag as well as the wrapped VEK (Volume Encryption Key) for each.
func (eh *EncryptionHelper) DeriveContainerKeybagKey(containerUUID types.UUID) [32]byte {
	var key [32]byte

	// First 16 bytes: container UUID
	copy(key[0:16], containerUUID[:])

	// Second 16 bytes: container UUID again
	copy(key[16:32], containerUUID[:])

	return key
}

// DeriveVolumeKeybagKey derives the 256-bit XTS-AES-128 encryption key for a volume keybag
// by concatenating the volume UUID with itself.
//
// Reference: APFS Advent Challenge Day 15
// Formula: volume_keybag_key = volume_uuid + volume_uuid
//
// This key is used to decrypt the volume keybag, which stores the wrapped KEKs
// (Key Encryption Keys) needed to access the VEK.
func (eh *EncryptionHelper) DeriveVolumeKeybagKey(volumeUUID types.UUID) [32]byte {
	var key [32]byte

	// First 16 bytes: volume UUID
	copy(key[0:16], volumeUUID[:])

	// Second 16 bytes: volume UUID again
	copy(key[16:32], volumeUUID[:])

	return key
}

// CalculateFSTreeNodeTweak calculates the initial tweak value for decrypting an encrypted
// File System Tree node.
//
// Reference: APFS Advent Challenge Day 18
// Formula: tweak0 = (ov_paddr * block_size) / 512
//
// For encrypted FS-Tree nodes, the tweak of the first 512 bytes can be determined by
// the physical location of the data. This tweak value is incremented for each subsequent
// 512-byte block.
//
// Parameters:
//   - physicalAddress: The physical block address (ov_paddr from object map)
//   - blockSize: The container's block size in bytes
//
// Returns:
//   - The initial tweak value for the first 512-byte block
func (eh *EncryptionHelper) CalculateFSTreeNodeTweak(physicalAddress types.Paddr, blockSize uint32) uint64 {
	// Calculate byte offset on disk
	byteOffset := uint64(physicalAddress) * uint64(blockSize)

	// Divide by 512 to get the tweak value
	// Each 512-byte block has its own tweak, incrementing from this initial value
	tweak := byteOffset / 512

	return tweak
}

// CalculateExtentTweak returns the tweak value for decrypting a file extent.
//
// Reference: APFS Advent Challenge Day 18
//
// For file extents, the tweak is stored directly in the crypto_id field of the
// j_file_extent_val_t structure. Extent data can be relocated on disk and is not
// guaranteed to be re-encrypted, so the initial tweak value must be preserved.
//
// Parameters:
//   - cryptoID: The crypto_id field from j_file_extent_val_t
//
// Returns:
//   - The initial tweak value for the first 512-byte block of the extent
func (eh *EncryptionHelper) CalculateExtentTweak(cryptoID uint64) uint64 {
	// For file extents, the crypto_id field directly contains the tweak value
	return cryptoID
}

// IncrementTweak increments a tweak value for the next 512-byte block.
//
// XTS-AES-128 encryption in APFS operates on 512-byte blocks. Each block uses
// a different tweak value, incremented from the initial tweak.
//
// Parameters:
//   - currentTweak: The tweak value for the current block
//
// Returns:
//   - The tweak value for the next 512-byte block
func (eh *EncryptionHelper) IncrementTweak(currentTweak uint64) uint64 {
	return currentTweak + 1
}

// IsObjectMapValueEncrypted checks if an object map value indicates an encrypted object.
//
// Reference: APFS Advent Challenge Day 18
//
// If the OMAP_VAL_ENCRYPTED flag is set in the ov_flags field, then the virtual object
// located at ov_paddr is encrypted.
//
// Parameters:
//   - flags: The ov_flags field from omap_val_t
//
// Returns:
//   - true if the object is encrypted, false otherwise
func (eh *EncryptionHelper) IsObjectMapValueEncrypted(flags uint32) bool {
	const omapValEncrypted uint32 = 0x00000001 // OMAP_VAL_ENCRYPTED
	return (flags & omapValEncrypted) != 0
}

// NOTE: Actual XTS-AES-128 encryption/decryption implementation would require
// integration with a cryptographic library such as:
//   - crypto/aes (Go standard library)
//   - github.com/gtank/cryptopasta or similar for XTS mode
//
// The functions above provide the correct key derivation and tweak calculation
// logic as specified in the APFS documentation, but the actual encryption/decryption
// operations are not implemented. To add full encryption support:
//
// 1. Import a crypto library with XTS-AES support
// 2. Implement DecryptFSTreeNode(data, key, tweak) function
// 3. Implement DecryptExtent(data, key, tweak) function
// 4. Implement EncryptFSTreeNode(data, key, tweak) function (for write support)
// 5. Implement EncryptExtent(data, key, tweak) function (for write support)
//
// Example signature for decryption:
//   func (eh *EncryptionHelper) DecryptBlock(
//       ciphertext []byte,
//       key [32]byte,
//       tweak uint64,
//   ) ([]byte, error)

// ValidateEncryptionKey validates that an encryption key has the correct length
func (eh *EncryptionHelper) ValidateEncryptionKey(key []byte) error {
	if len(key) != 32 {
		return fmt.Errorf("invalid key length: expected 32 bytes (256 bits), got %d", len(key))
	}
	return nil
}

// GetBlockSizeForTweak returns the block size used for tweak calculations (always 512)
func (eh *EncryptionHelper) GetBlockSizeForTweak() uint32 {
	return 512
}
