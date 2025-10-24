package snapshot

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// snapNameReader implements the SnapNameReader interface
type snapNameReader struct {
	key    *types.JSnapNameKeyT
	value  *types.JSnapNameValT
	endian binary.ByteOrder
}

// NewSnapNameReader creates a new SnapNameReader implementation
func NewSnapNameReader(keyData, valueData []byte, endian binary.ByteOrder) (interfaces.SnapNameReader, error) {
	if len(keyData) < 10 { // JKeyT (8 bytes) + NameLen (2 bytes) minimum
		return nil, fmt.Errorf("key data too small for snapshot name key: %d bytes", len(keyData))
	}

	if len(valueData) < 8 { // SnapXid (8 bytes)
		return nil, fmt.Errorf("value data too small for snapshot name value: %d bytes", len(valueData))
	}

	key, err := parseSnapNameKey(keyData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snapshot name key: %w", err)
	}

	value, err := parseSnapNameValue(valueData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snapshot name value: %w", err)
	}

	return &snapNameReader{
		key:    key,
		value:  value,
		endian: endian,
	}, nil
}

// parseSnapNameKey parses raw bytes into a JSnapNameKeyT structure
func parseSnapNameKey(data []byte, endian binary.ByteOrder) (*types.JSnapNameKeyT, error) {
	if len(data) < 10 {
		return nil, fmt.Errorf("insufficient data for snapshot name key")
	}

	key := &types.JSnapNameKeyT{}
	key.Hdr.ObjIdAndType = endian.Uint64(data[0:8])
	key.NameLen = endian.Uint16(data[8:10])

	// Extract the name (variable length, includes null terminator)
	if len(data) < 10+int(key.NameLen) {
		return nil, fmt.Errorf("insufficient data for snapshot name in key: have %d bytes, need %d", len(data)-10, key.NameLen)
	}

	key.Name = make([]byte, key.NameLen)
	copy(key.Name, data[10:10+key.NameLen])

	return key, nil
}

// parseSnapNameValue parses raw bytes into a JSnapNameValT structure
func parseSnapNameValue(data []byte, endian binary.ByteOrder) (*types.JSnapNameValT, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("insufficient data for snapshot name value")
	}

	value := &types.JSnapNameValT{}
	value.SnapXid = types.XidT(endian.Uint64(data[0:8]))

	return value, nil
}

// Name returns the snapshot's name.
func (snr *snapNameReader) Name() string {
	return strings.TrimRight(string(snr.key.Name), "\x00")
}

// SnapXID returns the last transaction identifier included in the snapshot.
func (snr *snapNameReader) SnapXID() uint64 {
	return uint64(snr.value.SnapXid)
}
