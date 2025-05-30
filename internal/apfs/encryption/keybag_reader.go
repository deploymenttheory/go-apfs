package encryption

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// keybagReader implements the KeybagReader interface
type keybagReader struct {
	keybag *types.KbLockerT
	endian binary.ByteOrder
}

// keybagEntryReader implements the KeybagEntryReader interface
type keybagEntryReader struct {
	entry  *types.KeybagEntryT
	endian binary.ByteOrder
}

// Ensure implementations match interfaces
var _ interfaces.KeybagReader = (*keybagReader)(nil)
var _ interfaces.KeybagEntryReader = (*keybagEntryReader)(nil)

// NewKeybagReader creates a new KeybagReader from raw keybag data
func NewKeybagReader(data []byte, endian binary.ByteOrder) (interfaces.KeybagReader, error) {
	if endian == nil {
		endian = binary.LittleEndian
	}

	keybag, err := parseKeybag(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse keybag: %w", err)
	}

	return &keybagReader{
		keybag: keybag,
		endian: endian,
	}, nil
}

// parseKeybag parses raw bytes into a KbLockerT structure
func parseKeybag(data []byte, endian binary.ByteOrder) (*types.KbLockerT, error) {
	// Minimum size: version(2) + nkeys(2) + nbytes(4) + padding(8) = 16 bytes
	if len(data) < 16 {
		return nil, fmt.Errorf("insufficient data for keybag: need at least 16 bytes, got %d", len(data))
	}

	keybag := &types.KbLockerT{}
	offset := 0

	// Parse version
	keybag.KlVersion = endian.Uint16(data[offset : offset+2])
	offset += 2

	// Parse number of keys
	keybag.KlNkeys = endian.Uint16(data[offset : offset+2])
	offset += 2

	// Parse number of bytes
	keybag.KlNbytes = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Skip padding
	copy(keybag.Padding[:], data[offset:offset+8])
	offset += 8

	// Validate version
	if keybag.KlVersion < types.ApfsKeybagVersion {
		return nil, fmt.Errorf("unsupported keybag version: %d (minimum supported: %d)",
			keybag.KlVersion, types.ApfsKeybagVersion)
	}

	// Parse entries if we have enough data
	if len(data) >= offset+int(keybag.KlNbytes) {
		entries, err := parseKeybagEntries(data[offset:], keybag.KlNkeys, endian)
		if err != nil {
			return nil, fmt.Errorf("failed to parse keybag entries: %w", err)
		}
		keybag.KlEntries = entries
	}

	return keybag, nil
}

// parseKeybagEntries parses keybag entries from raw data
func parseKeybagEntries(data []byte, count uint16, endian binary.ByteOrder) ([]types.KeybagEntryT, error) {
	entries := make([]types.KeybagEntryT, 0, count)
	offset := 0

	for i := uint16(0); i < count; i++ {
		// Check minimum size for entry header: UUID(16) + tag(2) + keylen(2) + padding(4) = 24 bytes
		if offset+24 > len(data) {
			return nil, fmt.Errorf("insufficient data for keybag entry %d", i)
		}

		entry := types.KeybagEntryT{}

		// Parse UUID
		copy(entry.KeUuid[:], data[offset:offset+16])
		offset += 16

		// Parse tag
		entry.KeTag = endian.Uint16(data[offset : offset+2])
		offset += 2

		// Parse key length
		entry.KeKeylen = endian.Uint16(data[offset : offset+2])
		offset += 2

		// Skip padding
		copy(entry.Padding[:], data[offset:offset+4])
		offset += 4

		// Validate key length
		if entry.KeKeylen > types.ApfsVolKeybagEntryMaxSize {
			return nil, fmt.Errorf("keybag entry %d key length %d exceeds maximum %d",
				i, entry.KeKeylen, types.ApfsVolKeybagEntryMaxSize)
		}

		// Parse key data if available
		if offset+int(entry.KeKeylen) <= len(data) {
			entry.KeKeydata = make([]byte, entry.KeKeylen)
			copy(entry.KeKeydata, data[offset:offset+int(entry.KeKeylen)])
			offset += int(entry.KeKeylen)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// Version returns the keybag version
func (kr *keybagReader) Version() uint16 {
	return kr.keybag.KlVersion
}

// EntryCount returns the number of entries in the keybag
func (kr *keybagReader) EntryCount() uint16 {
	return kr.keybag.KlNkeys
}

// TotalDataSize returns the total size of keybag entries in bytes
func (kr *keybagReader) TotalDataSize() uint32 {
	return kr.keybag.KlNbytes
}

// ListEntries returns all keybag entries
func (kr *keybagReader) ListEntries() []interfaces.KeybagEntryReader {
	entries := make([]interfaces.KeybagEntryReader, len(kr.keybag.KlEntries))
	for i, entry := range kr.keybag.KlEntries {
		// Create a copy to avoid pointer issues
		entryCopy := entry
		entries[i] = &keybagEntryReader{
			entry:  &entryCopy,
			endian: kr.endian,
		}
	}
	return entries
}

// IsValid checks if the keybag structure is valid
func (kr *keybagReader) IsValid() bool {
	// Check version
	if kr.keybag.KlVersion < types.ApfsKeybagVersion {
		return false
	}

	// Check if entry count matches actual entries
	if len(kr.keybag.KlEntries) != int(kr.keybag.KlNkeys) {
		return false
	}

	// Validate each entry
	for _, entry := range kr.keybag.KlEntries {
		if entry.KeKeylen > types.ApfsVolKeybagEntryMaxSize {
			return false
		}
		if len(entry.KeKeydata) != int(entry.KeKeylen) {
			return false
		}
	}

	return true
}

// KeybagEntryReader implementation

// UUID returns the UUID associated with the entry
func (ker *keybagEntryReader) UUID() types.UUID {
	return ker.entry.KeUuid
}

// Tag returns the keybag entry tag
func (ker *keybagEntryReader) Tag() types.KbTag {
	return types.KbTag(ker.entry.KeTag)
}

// TagDescription returns a human-readable description of the tag
func (ker *keybagEntryReader) TagDescription() string {
	tag := types.KbTag(ker.entry.KeTag)
	switch tag {
	case types.KbTagUnknown:
		return "Unknown"
	case types.KbTagReserved1:
		return "Reserved (1)"
	case types.KbTagVolumeKey:
		return "Volume Key (VEK)"
	case types.KbTagVolumeUnlockRecords:
		return "Volume Unlock Records"
	case types.KbTagVolumePassphraseHint:
		return "Volume Passphrase Hint"
	case types.KbTagWrappingMKey:
		return "Wrapping Media Key"
	case types.KbTagVolumeMKey:
		return "Volume Media Key"
	case types.KbTagReservedF8:
		return "Reserved (0xF8)"
	default:
		return fmt.Sprintf("Unknown Tag (0x%04X)", uint16(tag))
	}
}

// KeyLength returns the length of the entry's key data
func (ker *keybagEntryReader) KeyLength() uint16 {
	return ker.entry.KeKeylen
}

// KeyData returns the raw key data
func (ker *keybagEntryReader) KeyData() []byte {
	result := make([]byte, len(ker.entry.KeKeydata))
	copy(result, ker.entry.KeKeydata)
	return result
}

// IsPersonalRecoveryKey checks if this entry contains a personal recovery key
func (ker *keybagEntryReader) IsPersonalRecoveryKey() bool {
	return ker.entry.KeUuid == types.ApfsFvPersonalRecoveryKeyUuid
}

// IsVolumeKey checks if this entry contains a volume encryption key
func (ker *keybagEntryReader) IsVolumeKey() bool {
	return types.KbTag(ker.entry.KeTag) == types.KbTagVolumeKey
}

// IsUnlockRecord checks if this entry contains volume unlock records
func (ker *keybagEntryReader) IsUnlockRecord() bool {
	return types.KbTag(ker.entry.KeTag) == types.KbTagVolumeUnlockRecords
}
