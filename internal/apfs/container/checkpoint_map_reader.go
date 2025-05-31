package container

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// checkpointMapReader implements the CheckpointMapReader interface
type checkpointMapReader struct {
	checkpointMap *types.CheckpointMapPhysT
	mappings      []interfaces.CheckpointMappingReader
	data          []byte
	endian        binary.ByteOrder
}

// NewCheckpointMapReader creates a new CheckpointMapReader implementation
func NewCheckpointMapReader(data []byte, endian binary.ByteOrder) (interfaces.CheckpointMapReader, error) {
	if len(data) < 40 { // Minimum size: ObjPhysT (32) + CpmFlags (4) + CpmCount (4)
		return nil, fmt.Errorf("data too small for checkpoint map: %d bytes", len(data))
	}

	checkpointMap, err := parseCheckpointMap(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse checkpoint map: %w", err)
	}

	// Parse individual mappings
	mappings, err := parseCheckpointMappings(data[40:], checkpointMap.CpmCount, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse checkpoint mappings: %w", err)
	}

	return &checkpointMapReader{
		checkpointMap: checkpointMap,
		mappings:      mappings,
		data:          data,
		endian:        endian,
	}, nil
}

// parseCheckpointMap parses raw bytes into a CheckpointMapPhysT structure
func parseCheckpointMap(data []byte, endian binary.ByteOrder) (*types.CheckpointMapPhysT, error) {
	if len(data) < 40 {
		return nil, fmt.Errorf("insufficient data for checkpoint map")
	}

	cm := &types.CheckpointMapPhysT{}

	// Parse object header (32 bytes)
	copy(cm.CpmO.OChecksum[:], data[0:8])
	cm.CpmO.OOid = types.OidT(endian.Uint64(data[8:16]))
	cm.CpmO.OXid = types.XidT(endian.Uint64(data[16:24]))
	cm.CpmO.OType = endian.Uint32(data[24:28])
	cm.CpmO.OSubtype = endian.Uint32(data[28:32])

	// Parse checkpoint map specific fields
	cm.CpmFlags = endian.Uint32(data[32:36])
	cm.CpmCount = endian.Uint32(data[36:40])

	return cm, nil
}

// parseCheckpointMappings parses the array of checkpoint mappings
func parseCheckpointMappings(data []byte, count uint32, endian binary.ByteOrder) ([]interfaces.CheckpointMappingReader, error) {
	expectedSize := int(count) * 40 // Each mapping is 40 bytes
	if len(data) < expectedSize {
		return nil, fmt.Errorf("insufficient data for %d checkpoint mappings: got %d bytes, need %d", count, len(data), expectedSize)
	}

	mappings := make([]interfaces.CheckpointMappingReader, count)
	for i := uint32(0); i < count; i++ {
		offset := int(i) * 40
		mappingData := data[offset : offset+40]

		mapping, err := NewCheckpointMappingReader(mappingData, endian)
		if err != nil {
			return nil, fmt.Errorf("failed to parse checkpoint mapping %d: %w", i, err)
		}

		mappings[i] = mapping
	}

	return mappings, nil
}

// Flags returns the checkpoint map flags
func (cmpr *checkpointMapReader) Flags() uint32 {
	return cmpr.checkpointMap.CpmFlags
}

// Count returns the number of checkpoint mappings in the array
func (cmpr *checkpointMapReader) Count() uint32 {
	return cmpr.checkpointMap.CpmCount
}

// Mappings returns the array of checkpoint mappings
func (cmpr *checkpointMapReader) Mappings() []interfaces.CheckpointMappingReader {
	return cmpr.mappings
}

// IsLast checks if this is the last checkpoint-mapping block in a given checkpoint
func (cmpr *checkpointMapReader) IsLast() bool {
	// Check for checkpoint flags that indicate this is the last block
	// The exact flag value would need to be defined in types package
	// For now, we'll implement a basic check
	const CHECKPOINT_MAP_LAST uint32 = 0x00000001 // This would need to be defined in types
	return cmpr.checkpointMap.CpmFlags&CHECKPOINT_MAP_LAST != 0
}
