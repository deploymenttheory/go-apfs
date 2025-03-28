// File: pkg/container/omap.go
package container

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/types"
)

// OMapPhysSize defines the fixed size of an OMapPhys structure.
const OMapPhysSize = 56

// ReadOMapPhys reads and parses an OMapPhys structure from a block device.
func ReadOMapPhys(device types.BlockDevice, addr types.PAddr) (*types.OMapPhys, error) {
	blockSize := device.GetBlockSize()
	if OMapPhysSize > int(blockSize) {
		return nil, fmt.Errorf("OMapPhys size (%d) exceeds device block size (%d)", OMapPhysSize, blockSize)
	}

	data, err := device.ReadBlock(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to read block at addr %d: %w", addr, err)
	}

	if len(data) < OMapPhysSize {
		return nil, fmt.Errorf("data too short: expected %d bytes, got %d", OMapPhysSize, len(data))
	}

	var omap types.OMapPhys
	reader := binary.LittleEndian

	omap.Header.Cksum = types.Checksum(data[0:8])
	omap.Header.OID = types.OID(reader.Uint64(data[8:16]))
	omap.Header.XID = types.XID(reader.Uint64(data[16:24]))
	omap.Header.Type = reader.Uint32(data[24:28])
	omap.Header.Subtype = reader.Uint32(data[28:32])

	omap.Flags = reader.Uint32(data[32:36])
	omap.SnapCount = reader.Uint32(data[36:40])
	omap.TreeType = reader.Uint32(data[40:44])
	omap.SnapshotTreeType = reader.Uint32(data[44:48])
	omap.TreeOID = types.OID(reader.Uint64(data[48:56]))
	// SnapshotTreeOID and other dynamic fields would be parsed separately as needed

	return &omap, nil
}

// Validate performs basic validation on the OMapPhys structure.
func ValidateOMapPhys(omap *types.OMapPhys) error {
	if omap.TreeOID == types.OIDInvalid {
		return fmt.Errorf("invalid TreeOID: cannot be OIDInvalid (0)")
	}
	return nil
}
