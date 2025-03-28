package container

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/pkg/types"
)

// BTreeNodePhysSize defines the fixed-size portion of a B-tree node structure
const BTreeNodePhysSize = 64

// ReadBTreeNodePhys reads and parses a BTreeNodePhys structure from a block device.
func ReadBTreeNodePhys(device types.BlockDevice, addr types.PAddr) (*types.BTreeNodePhys, error) {
	blockSize := device.GetBlockSize()
	if BTreeNodePhysSize > int(blockSize) {
		return nil, fmt.Errorf("BTreeNodePhys size (%d) exceeds device block size (%d)", BTreeNodePhysSize, blockSize)
	}

	data, err := device.ReadBlock(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to read block at addr %d: %w", addr, err)
	}

	if len(data) < BTreeNodePhysSize {
		return nil, fmt.Errorf("data too short: expected %d bytes, got %d", BTreeNodePhysSize, len(data))
	}

	var node types.BTreeNodePhys
	reader := binary.LittleEndian

	node.Header.Cksum = types.Checksum(data[0:8])
	node.Header.OID = types.OID(reader.Uint64(data[8:16]))
	node.Header.XID = types.XID(reader.Uint64(data[16:24]))
	node.Header.Type = reader.Uint32(data[24:28])
	node.Header.Subtype = reader.Uint32(data[28:32])

	node.Flags = reader.Uint16(data[32:34])
	node.Level = reader.Uint16(data[34:36])
	node.NKeys = reader.Uint32(data[36:40])
	node.TableSpace = reader.Uint16(data[40:42])
	node.FreeSpace = reader.Uint16(data[42:44])
	node.KeyFreeListOffset = reader.Uint16(data[44:46])
	node.ValFreeListOffset = reader.Uint16(data[46:48])
	node.DataOffset = reader.Uint16(data[48:50])

	return &node, nil
}

// Validate performs basic validation on the BTreeNodePhys structure.
func (node *types.BTreeNodePhys) Validate() error {
	if node.NKeys == 0 {
		return fmt.Errorf("invalid number of keys: %d", node.NKeys)
	}
	if node.Level > 16 { // arbitrary sanity check
		return fmt.Errorf("unreasonable B-tree level: %d", node.Level)
	}

	// Additional checks can be implemented here
	return nil
}
