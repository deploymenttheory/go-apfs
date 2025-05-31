package encryption

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// mediaKeybagReader implements the KeybagReader interface for MediaKeybagT structures
type mediaKeybagReader struct {
	mediaKeybag *types.MediaKeybagT
	endian      binary.ByteOrder
}

// Ensure mediaKeybagReader implements the KeybagReader interface
var _ interfaces.KeybagReader = (*mediaKeybagReader)(nil)

// NewMediaKeybagReader creates a new KeybagReader from raw media keybag data
func NewMediaKeybagReader(data []byte, endian binary.ByteOrder) (interfaces.KeybagReader, error) {
	if endian == nil {
		endian = binary.LittleEndian
	}

	mediaKeybag, err := parseMediaKeybag(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse media keybag: %w", err)
	}

	return &mediaKeybagReader{
		mediaKeybag: mediaKeybag,
		endian:      endian,
	}, nil
}

// parseMediaKeybag parses raw bytes into a MediaKeybagT structure
func parseMediaKeybag(data []byte, endian binary.ByteOrder) (*types.MediaKeybagT, error) {
	// Minimum size: obj_phys_t (32 bytes) + kb_locker_t header (16 bytes) = 48 bytes
	if len(data) < 48 {
		return nil, fmt.Errorf("insufficient data for media keybag: need at least 48 bytes, got %d", len(data))
	}

	mediaKeybag := &types.MediaKeybagT{}
	offset := 0

	// Parse object header (32 bytes)
	copy(mediaKeybag.MkObj.OChecksum[:], data[offset:offset+8])
	offset += 8
	mediaKeybag.MkObj.OOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	mediaKeybag.MkObj.OXid = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	mediaKeybag.MkObj.OType = endian.Uint32(data[offset : offset+4])
	offset += 4
	mediaKeybag.MkObj.OSubtype = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse the embedded keybag
	keybag, err := parseKeybag(data[offset:], endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded keybag: %w", err)
	}

	mediaKeybag.MkLocker = *keybag

	return mediaKeybag, nil
}

// Version returns the keybag version from the embedded keybag
func (mkr *mediaKeybagReader) Version() uint16 {
	return mkr.mediaKeybag.MkLocker.KlVersion
}

// EntryCount returns the number of entries in the embedded keybag
func (mkr *mediaKeybagReader) EntryCount() uint16 {
	return mkr.mediaKeybag.MkLocker.KlNkeys
}

// TotalDataSize returns the total size of keybag entries in bytes from the embedded keybag
func (mkr *mediaKeybagReader) TotalDataSize() uint32 {
	return mkr.mediaKeybag.MkLocker.KlNbytes
}

// ListEntries returns all keybag entries from the embedded keybag
func (mkr *mediaKeybagReader) ListEntries() []interfaces.KeybagEntryReader {
	entries := make([]interfaces.KeybagEntryReader, len(mkr.mediaKeybag.MkLocker.KlEntries))
	for i, entry := range mkr.mediaKeybag.MkLocker.KlEntries {
		// Create a copy to avoid pointer issues
		entryCopy := entry
		entries[i] = &keybagEntryReader{
			entry:  &entryCopy,
			endian: mkr.endian,
		}
	}
	return entries
}

// IsValid checks if the media keybag structure is valid
func (mkr *mediaKeybagReader) IsValid() bool {
	// Check object type - media keybag uses a fourcc type ('mkey') so we don't mask it
	if mkr.mediaKeybag.MkObj.OType != types.ObjectTypeMediaKeybag {
		return false
	}

	// Check embedded keybag version
	if mkr.mediaKeybag.MkLocker.KlVersion < types.ApfsKeybagVersion {
		return false
	}

	// Check if entry count matches actual entries
	if len(mkr.mediaKeybag.MkLocker.KlEntries) != int(mkr.mediaKeybag.MkLocker.KlNkeys) {
		return false
	}

	// Validate each entry
	for _, entry := range mkr.mediaKeybag.MkLocker.KlEntries {
		if entry.KeKeylen > types.ApfsVolKeybagEntryMaxSize {
			return false
		}
		if len(entry.KeKeydata) != int(entry.KeKeylen) {
			return false
		}
	}

	return true
}

// ObjectID returns the object identifier of the media keybag
func (mkr *mediaKeybagReader) ObjectID() types.OidT {
	return mkr.mediaKeybag.MkObj.OOid
}

// TransactionID returns the transaction identifier of the media keybag
func (mkr *mediaKeybagReader) TransactionID() types.XidT {
	return mkr.mediaKeybag.MkObj.OXid
}

// ObjectType returns the object type of the media keybag
func (mkr *mediaKeybagReader) ObjectType() uint32 {
	return mkr.mediaKeybag.MkObj.OType
}

// ObjectSubtype returns the object subtype of the media keybag
func (mkr *mediaKeybagReader) ObjectSubtype() uint32 {
	return mkr.mediaKeybag.MkObj.OSubtype
}
