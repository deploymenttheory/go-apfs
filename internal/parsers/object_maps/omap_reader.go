package objectmaps

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// OmapReader parses and provides access to object map data
type OmapReader struct {
	omap   *types.OmapPhysT
	endian binary.ByteOrder
}

// NewOmapReader creates a new OmapReader from raw bytes
func NewOmapReader(data []byte, endian binary.ByteOrder) (*OmapReader, error) {
	if endian == nil {
		endian = binary.LittleEndian
	}

	omap, err := parseOmapPhys(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse object map: %w", err)
	}

	return &OmapReader{
		omap:   omap,
		endian: endian,
	}, nil
}

// parseOmapPhys parses raw bytes into an OmapPhysT structure
func parseOmapPhys(data []byte, endian binary.ByteOrder) (*types.OmapPhysT, error) {
	if len(data) < 72 { // Minimum size for omap_phys_t
		return nil, fmt.Errorf("insufficient data for object map: need at least 72 bytes, got %d", len(data))
	}

	omap := &types.OmapPhysT{}

	// Parse object header (first 32 bytes)
	copy(omap.OmO.OChecksum[:], data[0:8])
	omap.OmO.OOid = types.OidT(endian.Uint64(data[8:16]))
	omap.OmO.OXid = types.XidT(endian.Uint64(data[16:24]))
	omap.OmO.OType = endian.Uint32(data[24:28])
	omap.OmO.OSubtype = endian.Uint32(data[28:32])

	offset := 32

	// Parse object map specific fields
	omap.OmFlags = endian.Uint32(data[offset : offset+4])
	offset += 4

	omap.OmSnapCount = endian.Uint32(data[offset : offset+4])
	offset += 4

	omap.OmTreeType = endian.Uint32(data[offset : offset+4])
	offset += 4

	omap.OmSnapshotTreeType = endian.Uint32(data[offset : offset+4])
	offset += 4

	omap.OmTreeOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8

	omap.OmSnapshotTreeOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8

	omap.OmMostRecentSnap = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8

	omap.OmPendingRevertMin = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8

	omap.OmPendingRevertMax = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8

	return omap, nil
}

// GetOmap returns the parsed object map
func (or *OmapReader) GetOmap() *types.OmapPhysT {
	return or.omap
}

// GetTreeOid returns the OID of the B-tree containing the actual object mappings
func (or *OmapReader) GetTreeOid() types.OidT {
	return or.omap.OmTreeOid
}

// GetMostRecentSnap returns the most recent snapshot transaction ID
func (or *OmapReader) GetMostRecentSnap() types.XidT {
	return or.omap.OmMostRecentSnap
}

// Validate checks if the object map is valid
func (or *OmapReader) Validate() (bool, []string) {
	issues := []string{}

	// Check that tree OID is not zero
	if or.omap.OmTreeOid == 0 {
		issues = append(issues, "object map tree OID is zero")
	}

	// Check flags are within valid range
	if or.omap.OmFlags&^types.OmapValidFlags != 0 {
		issues = append(issues, fmt.Sprintf("invalid object map flags: 0x%08X", or.omap.OmFlags))
	}

	return len(issues) == 0, issues
}