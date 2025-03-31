// File: pkg/crypto/keybag.go
package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/container"
	"github.com/deploymenttheory/go-apfs/apfs/pkg/types"
)

// Constants for keybag implementation
const (
	KBLockerVersion = 2 // Current keybag locker version
)

// KeybagEntryTag represents the keybag entry tag types
type KeybagEntryTag uint16

// Keybag represents an APFS keybag structure (kb_locker_t)
type Keybag struct {
	Version  uint16        // Keybag version (should be 2)
	NumKeys  uint16        // Number of keys in the keybag
	NumBytes uint32        // Number of bytes of key data
	Padding  [8]byte       // Padding
	Entries  []KeybagEntry // Keybag entries
}

// KeybagEntry represents an entry in the keybag
type KeybagEntry struct {
	UUID    types.UUID // UUID for this entry
	Tag     uint16     // Tag type
	KeyLen  uint16     // Length of key data
	Padding [4]byte    // Padding
	KeyData []byte     // Actual key data
}

// CreateKeybag creates a new keybag structure for a volume
func CreateKeybag(volumeUUID types.UUID) *Keybag {
	return &Keybag{
		Version:  KBLockerVersion,
		NumKeys:  0,
		NumBytes: 0,
		Entries:  make([]KeybagEntry, 0),
	}
}

// AddKeyToKeybag adds a key with the specified tag to a keybag
func AddKeyToKeybag(keybag *Keybag, key []byte, tag uint16, uuid types.UUID) error {
	if keybag == nil {
		return errors.New("keybag is nil")
	}

	if len(key) == 0 {
		return errors.New("key data cannot be empty")
	}

	// Check for existing entry with the same UUID and tag
	for i, entry := range keybag.Entries {
		if entry.Tag == tag && bytes.Equal(entry.UUID[:], uuid[:]) {
			// Update existing entry
			keybag.NumBytes = keybag.NumBytes - uint32(keybag.Entries[i].KeyLen) + uint32(len(key))
			keybag.Entries[i].KeyLen = uint16(len(key))
			keybag.Entries[i].KeyData = make([]byte, len(key))
			copy(keybag.Entries[i].KeyData, key)
			return nil
		}
	}

	// Create new entry
	entry := KeybagEntry{
		Tag:     tag,
		KeyLen:  uint16(len(key)),
		KeyData: make([]byte, len(key)),
	}
	copy(entry.UUID[:], uuid[:])
	copy(entry.KeyData, key)

	keybag.Entries = append(keybag.Entries, entry)
	keybag.NumKeys++
	keybag.NumBytes += uint32(len(key))

	return nil
}

// RemoveKeyFromKeybag removes a key with the specified tag and UUID from the keybag
func RemoveKeyFromKeybag(keybag *Keybag, uuid types.UUID, tag uint16) error {
	if keybag == nil {
		return errors.New("keybag is nil")
	}

	for i, entry := range keybag.Entries {
		if entry.Tag == tag && bytes.Equal(entry.UUID[:], uuid[:]) {
			// Update keybag metrics
			keybag.NumBytes -= uint32(entry.KeyLen)
			keybag.NumKeys--

			// Remove the entry
			keybag.Entries = append(keybag.Entries[:i], keybag.Entries[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("key with UUID %x and tag %d not found in keybag", uuid, tag)
}

// RetrieveKeyFromKeybag gets a key from the keybag with the specified UUID and tag
func RetrieveKeyFromKeybag(keybag *Keybag, uuid types.UUID, tag uint16) ([]byte, error) {
	if keybag == nil {
		return nil, errors.New("keybag is nil")
	}

	for _, entry := range keybag.Entries {
		if entry.Tag == tag && bytes.Equal(entry.UUID[:], uuid[:]) {
			// Return a copy of the key data
			keyData := make([]byte, len(entry.KeyData))
			copy(keyData, entry.KeyData)
			return keyData, nil
		}
	}

	return nil, fmt.Errorf("key with UUID %x and tag %d not found in keybag", uuid, tag)
}

// RotateKeybag replaces all keys in the keybag with new ones by re-encrypting with new wrapping keys
func RotateKeybag(keybag *Keybag, oldPassword, newPassword string, volumeUUID types.UUID) error {
	if keybag == nil {
		return errors.New("keybag is nil")
	}

	// Derive a new Key Encryption Key (KEK) using the new password
	newKEK, err := DeriveKeyFromPassword(newPassword, volumeUUID[:])
	if err != nil {
		return fmt.Errorf("failed to derive new KEK: %w", err)
	}

	// Derive the old KEK to unwrap existing keys
	oldKEK, err := DeriveKeyFromPassword(oldPassword, volumeUUID[:])
	if err != nil {
		return fmt.Errorf("failed to derive old KEK: %w", err)
	}

	// Process each keybag entry
	for i, entry := range keybag.Entries {
		switch entry.Tag {
		case types.KBTagVolumeKey:
			// For Volume Key entries, unwrap with old KEK and rewrap with new KEK
			unwrappedKey, err := UnwrapKey(entry.KeyData, oldKEK)
			if err != nil {
				return fmt.Errorf("failed to unwrap volume key: %w", err)
			}

			// Rewrap the VEK with the new KEK
			wrappedKey, err := WrapKey(unwrappedKey, newKEK)
			if err != nil {
				return fmt.Errorf("failed to rewrap volume key: %w", err)
			}

			// Update the entry
			keybag.Entries[i].KeyData = wrappedKey
			keybag.Entries[i].KeyLen = uint16(len(wrappedKey))

		case types.KBTagVolumeUnlockRecords:
			// Update the unlock records for the volume
			// Generate a new wrapped KEK as an unlock record
			wrappedKEK, err := WrapKey(newKEK, newKEK) // Self-wrapped for unlock records
			if err != nil {
				return fmt.Errorf("failed to create new unlock record: %w", err)
			}

			// Update the entry
			keybag.Entries[i].KeyData = wrappedKEK
			keybag.Entries[i].KeyLen = uint16(len(wrappedKEK))

		case types.KBTagVolumePassphraseHint:
			// Password hint doesn't need to be re-encrypted, but may need updating
			// In a real implementation, we would update the hint if provided

		case types.KBTagVolumeMKey:
			// For media keys, re-wrap with the new KEK
			unwrappedKey, err := UnwrapKey(entry.KeyData, oldKEK)
			if err != nil {
				return fmt.Errorf("failed to unwrap media key: %w", err)
			}

			wrappedKey, err := WrapKey(unwrappedKey, newKEK)
			if err != nil {
				return fmt.Errorf("failed to rewrap media key: %w", err)
			}

			// Update the entry
			keybag.Entries[i].KeyData = wrappedKey
			keybag.Entries[i].KeyLen = uint16(len(wrappedKey))

		default:
			// Skip unknown tag types
		}
	}

	// Update keybag metadata
	keybag.NumBytes = 0
	for _, entry := range keybag.Entries {
		keybag.NumBytes += uint32(entry.KeyLen)
	}

	// Set flag to indicate key rolling has occurred
	// This would set APFS_INCOMPAT_ENC_ROLLED flag on the volume in a full implementation

	return nil
}

// SerializeKeybag converts a keybag structure to its binary representation
func SerializeKeybag(keybag *Keybag) ([]byte, error) {
	if keybag == nil {
		return nil, errors.New("keybag is nil")
	}

	// Calculate total size needed
	totalSize := 16 // Fixed header: 2(version) + 2(numKeys) + 4(numBytes) + 8(padding)

	// Add space for each entry's header
	entryHeaderSize := 24 // 16(UUID) + 2(tag) + 2(keylen) + 4(padding)
	totalSize += len(keybag.Entries) * entryHeaderSize

	// Add space for actual key data
	for _, entry := range keybag.Entries {
		totalSize += len(entry.KeyData)
	}

	// Create buffer and writer
	buf := make([]byte, totalSize)
	w := bytes.NewBuffer(buf[:0])

	// Write keybag header
	binary.Write(w, binary.LittleEndian, keybag.Version)
	binary.Write(w, binary.LittleEndian, keybag.NumKeys)
	binary.Write(w, binary.LittleEndian, keybag.NumBytes)
	binary.Write(w, binary.LittleEndian, keybag.Padding)

	// Write each entry
	for _, entry := range keybag.Entries {
		binary.Write(w, binary.LittleEndian, entry.UUID)
		binary.Write(w, binary.LittleEndian, entry.Tag)
		binary.Write(w, binary.LittleEndian, entry.KeyLen)
		binary.Write(w, binary.LittleEndian, entry.Padding)
		binary.Write(w, binary.LittleEndian, entry.KeyData)
	}

	return w.Bytes(), nil
}

// DeserializeKeybag creates a keybag structure from its binary representation
func DeserializeKeybag(data []byte) (*Keybag, error) {
	if len(data) < 16 {
		return nil, errors.New("keybag data too short")
	}

	keybag := &Keybag{}
	r := bytes.NewReader(data)

	// Read keybag header
	binary.Read(r, binary.LittleEndian, &keybag.Version)
	binary.Read(r, binary.LittleEndian, &keybag.NumKeys)
	binary.Read(r, binary.LittleEndian, &keybag.NumBytes)
	binary.Read(r, binary.LittleEndian, &keybag.Padding)

	// Check version
	if keybag.Version != KBLockerVersion {
		return nil, fmt.Errorf("unsupported keybag version: %d", keybag.Version)
	}

	// Allocate entries
	keybag.Entries = make([]KeybagEntry, keybag.NumKeys)

	// Read entries
	bytesRead := int64(16) // Header size
	for i := uint16(0); i < keybag.NumKeys; i++ {
		if bytesRead >= int64(len(data)) {
			return nil, errors.New("unexpected end of keybag data")
		}

		var entry KeybagEntry
		binary.Read(r, binary.LittleEndian, &entry.UUID)
		binary.Read(r, binary.LittleEndian, &entry.Tag)
		binary.Read(r, binary.LittleEndian, &entry.KeyLen)
		binary.Read(r, binary.LittleEndian, &entry.Padding)

		bytesRead += 24 // Size of entry header

		// Bounds check for key data
		if bytesRead+int64(entry.KeyLen) > int64(len(data)) {
			return nil, errors.New("key data extends beyond end of buffer")
		}

		// Read key data
		entry.KeyData = make([]byte, entry.KeyLen)
		n, err := r.Read(entry.KeyData)
		if err != nil || n != int(entry.KeyLen) {
			return nil, fmt.Errorf("failed to read key data: %v", err)
		}
		bytesRead += int64(entry.KeyLen)

		keybag.Entries[i] = entry
	}

	return keybag, nil
}

// EncryptKeybag encrypts a serialized keybag using AES-CBC with the provided key
// For container keybags, this key is derived from the container UUID
// For volume keybags, this key is derived from the volume UUID
func EncryptKeybag(serializedKeybag []byte, uuid types.UUID) ([]byte, error) {
	// Generate encryption key from UUID
	key := deriveKeyFromUUID(uuid)

	// Generate random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %v", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %v", err)
	}

	// Pad data to AES block size
	paddedData := padPKCS7(serializedKeybag, aes.BlockSize)

	// Encrypt data with CBC mode
	encryptedData := make([]byte, len(paddedData))
	cbc := cipher.NewCBCEncrypter(block, iv)
	cbc.CryptBlocks(encryptedData, paddedData)

	// Prepend IV to encrypted data
	result := append(iv, encryptedData...)
	return result, nil
}

// DecryptKeybag decrypts an encrypted keybag using AES-CBC with the provided key
func DecryptKeybag(encryptedKeybag []byte, uuid types.UUID) ([]byte, error) {
	if len(encryptedKeybag) < aes.BlockSize {
		return nil, errors.New("encrypted keybag data too short")
	}

	// Generate decryption key from UUID
	key := deriveKeyFromUUID(uuid)

	// Extract IV (first AES block)
	iv := encryptedKeybag[:aes.BlockSize]
	ciphertext := encryptedKeybag[aes.BlockSize:]

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %v", err)
	}

	// Decrypt data with CBC mode
	decryptedData := make([]byte, len(ciphertext))
	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(decryptedData, ciphertext)

	// Remove padding
	unpaddedData, err := unpadPKCS7(decryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to unpad decrypted data: %v", err)
	}

	return unpaddedData, nil
}

// Internal helper functions

// deriveKeyFromUUID derives a 32-byte key from a UUID (similar to how APFS uses UUIDs for keybag encryption)
func deriveKeyFromUUID(uuid types.UUID) []byte {
	// In APFS, there's a specific key derivation function for this
	// For this implementation, we'll use a simple SHA-256 derivation
	h := sha256.New()
	h.Write(uuid[:])
	return h.Sum(nil)
}

// ReadKeybagFromDevice reads a keybag from a specified location on the device
func ReadKeybagFromDevice(device types.BlockDevice, prange types.PRange) (*Keybag, error) {
	if prange.BlockCount == 0 {
		return nil, errors.New("invalid keybag range: zero blocks")
	}

	// Calculate total size in bytes
	blockSize := device.GetBlockSize()
	totalSize := uint64(blockSize) * prange.BlockCount

	// Allocate buffer for keybag data
	keybagData := make([]byte, totalSize)

	// Read blocks
	var offset uint64
	for i := uint64(0); i < prange.BlockCount; i++ {
		blockData, err := device.ReadBlock(prange.StartAddr + types.PAddr(i))
		if err != nil {
			return nil, fmt.Errorf("failed to read keybag block %d: %w", i, err)
		}

		// Copy block data to buffer
		copy(keybagData[offset:], blockData)
		offset += uint64(blockSize)
	}

	// Deserialize the keybag
	return DeserializeKeybag(keybagData)
}

// WriteKeybagToDevice writes a keybag to a specified location on the device
func WriteKeybagToDevice(device types.BlockDevice, keybag *Keybag, prange types.PRange) error {
	if keybag == nil {
		return errors.New("keybag is nil")
	}

	if prange.BlockCount == 0 {
		return errors.New("invalid keybag range: zero blocks")
	}

	// Serialize the keybag
	keybagData, err := SerializeKeybag(keybag)
	if err != nil {
		return fmt.Errorf("failed to serialize keybag: %w", err)
	}

	// Calculate total space available
	blockSize := device.GetBlockSize()
	totalSpace := uint64(blockSize) * prange.BlockCount

	// Check if keybag fits in the allocated space
	if uint64(len(keybagData)) > totalSpace {
		return fmt.Errorf("keybag size (%d bytes) exceeds allocated space (%d bytes)",
			len(keybagData), totalSpace)
	}

	// Pad the data to fill all blocks
	paddedData := make([]byte, totalSpace)
	copy(paddedData, keybagData)

	// Write blocks
	var offset uint64
	for i := uint64(0); i < prange.BlockCount; i++ {
		blockToWrite := paddedData[offset : offset+uint64(blockSize)]

		if err := device.WriteBlock(prange.StartAddr+types.PAddr(i), blockToWrite); err != nil {
			return fmt.Errorf("failed to write keybag block %d: %w", i, err)
		}

		offset += uint64(blockSize)
	}

	return nil
}

// AllocateKeybagSpace allocates space for a keybag in the container
func AllocateKeybagSpace(device types.BlockDevice, nxsb *types.NXSuperblock, spaceManager *container.SpaceManager, keybagSize uint32) (types.PRange, error) {
	if keybagSize == 0 {
		return types.PRange{}, errors.New("invalid keybag size: zero")
	}

	if nxsb == nil {
		return types.PRange{}, errors.New("container superblock is nil")
	}

	if spaceManager == nil {
		return types.PRange{}, errors.New("space manager is nil")
	}

	// Use the container's actual block size
	blockSize := nxsb.BlockSize

	// Calculate number of blocks needed (rounded up)
	blocksNeeded := (keybagSize + blockSize - 1) / blockSize

	// First try to allocate from the internal pool for critical metadata
	var startAddr types.PAddr
	var err error

	// Try to allocate contiguous blocks
	startAddr, err = spaceManager.GetContiguousBlocks(blocksNeeded)
	if err != nil {
		// Fall back to allocating individual blocks if contiguous allocation fails
		var blocks []types.PAddr
		for i := uint32(0); i < blocksNeeded; i++ {
			addr, blockErr := spaceManager.AllocateBlock()
			if blockErr != nil {
				// Free any blocks we've already allocated
				for _, blockAddr := range blocks {
					spaceManager.FreeBlock(blockAddr)
				}
				return types.PRange{}, fmt.Errorf("failed to allocate space for keybag: %w", blockErr)
			}
			blocks = append(blocks, addr)
		}

		// For non-contiguous allocation, we'd need to create a mapping
		// In a real implementation, we would create an extent list tree
		// For now, report an error since we need contiguous space for proper keybag storage
		for _, blockAddr := range blocks {
			spaceManager.FreeBlock(blockAddr)
		}
		return types.PRange{}, errors.New("could not allocate contiguous space for keybag")
	}

	// Log allocation in transaction if needed
	// This would be part of a transaction system in a complete implementation

	// Update free block count in the superblock if needed
	// In a complete implementation, this would be done by the transaction system

	return types.PRange{
		StartAddr:  startAddr,
		BlockCount: uint64(blocksNeeded),
	}, nil
}
