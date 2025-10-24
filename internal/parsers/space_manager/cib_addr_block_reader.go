package spacemanager

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// CibAddrBlockReader provides parsing capabilities for cib_addr_block_t structures
// A CIB address block contains an array of chunk-info block addresses for hierarchical management
type CibAddrBlockReader struct {
	cibAddrBlock *types.CibAddrBlockT
	data         []byte
	endian       binary.ByteOrder
}

// NewCibAddrBlockReader creates a new CIB address block reader
// CIB address blocks provide the top level of the chunk management hierarchy
func NewCibAddrBlockReader(data []byte, endian binary.ByteOrder) (*CibAddrBlockReader, error) {
	// Minimum size: obj_phys_t (32) + index (4) + count (4) = 40 bytes + variable address array
	if len(data) < 40 {
		return nil, fmt.Errorf("data too small for CIB address block: %d bytes, need at least 40", len(data))
	}

	cibAddrBlock, err := parseCibAddrBlock(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CIB address block: %w", err)
	}

	// Validate the object type
	objectType := cibAddrBlock.CabO.OType & types.ObjectTypeMask
	if objectType != types.ObjectTypeSpacemanCab {
		return nil, fmt.Errorf("invalid CIB address block object type: 0x%x", objectType)
	}

	return &CibAddrBlockReader{
		cibAddrBlock: cibAddrBlock,
		data:         data,
		endian:       endian,
	}, nil
}

// parseCibAddrBlock parses raw bytes into a CibAddrBlockT structure
// This follows the exact layout of cib_addr_block_t from Apple File System Reference
func parseCibAddrBlock(data []byte, endian binary.ByteOrder) (*types.CibAddrBlockT, error) {
	if len(data) < 40 {
		return nil, fmt.Errorf("insufficient data for CIB address block")
	}

	cab := &types.CibAddrBlockT{}
	offset := 0

	// Parse object header (obj_phys_t): 32 bytes
	// Contains checksum, object ID, transaction ID, type, and subtype
	copy(cab.CabO.OChecksum[:], data[offset:offset+8])
	offset += 8
	cab.CabO.OOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	cab.CabO.OXid = types.XidT(endian.Uint64(data[offset : offset+8]))
	offset += 8
	cab.CabO.OType = endian.Uint32(data[offset : offset+4])
	offset += 4
	cab.CabO.OSubtype = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse CIB address block specific fields
	// uint32_t cab_index - Index of this chunk-info address block
	cab.CabIndex = endian.Uint32(data[offset : offset+4])
	offset += 4

	// uint32_t cab_cib_count - Number of chunk-info blocks referenced by this address block
	cab.CabCibCount = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse paddr_t array: cab_cib_addr[]
	// Each paddr_t is 8 bytes
	if cab.CabCibCount > 0 {
		addrSize := 8 // Size of paddr_t
		requiredSize := offset + int(cab.CabCibCount)*addrSize
		if len(data) < requiredSize {
			return nil, fmt.Errorf("insufficient data for CIB addresses: need %d bytes, have %d",
				requiredSize, len(data))
		}

		cab.CabCibAddr = make([]types.Paddr, cab.CabCibCount)
		for i := uint32(0); i < cab.CabCibCount; i++ {
			cab.CabCibAddr[i] = types.Paddr(endian.Uint64(data[offset : offset+8]))
			offset += 8
		}
	}

	return cab, nil
}

// GetCibAddrBlock returns the CIB address block structure
// Provides direct access to the underlying cib_addr_block_t for advanced operations
func (cabr *CibAddrBlockReader) GetCibAddrBlock() *types.CibAddrBlockT {
	return cabr.cibAddrBlock
}

// ObjectHeader returns the object header
// Provides access to object ID, transaction ID, and other metadata
func (cabr *CibAddrBlockReader) ObjectHeader() *types.ObjPhysT {
	return &cabr.cibAddrBlock.CabO
}

// Index returns the index of this CIB address block
// Used to identify the position of this block in the space manager hierarchy
func (cabr *CibAddrBlockReader) Index() uint32 {
	return cabr.cibAddrBlock.CabIndex
}

// CibCount returns the number of chunk-info blocks referenced by this address block
// Indicates how many CIB blocks are managed by this address block
func (cabr *CibAddrBlockReader) CibCount() uint32 {
	return cabr.cibAddrBlock.CabCibCount
}

// GetCibAddress returns a specific chunk-info block address by index
// Provides access to individual CIB addresses within this block
func (cabr *CibAddrBlockReader) GetCibAddress(index uint32) (types.Paddr, error) {
	if index >= cabr.cibAddrBlock.CabCibCount {
		return 0, fmt.Errorf("CIB address index %d out of range (have %d addresses)",
			index, cabr.cibAddrBlock.CabCibCount)
	}
	return cabr.cibAddrBlock.CabCibAddr[index], nil
}

// GetAllCibAddresses returns all chunk-info block addresses in this block
// Provides bulk access to all CIB addresses
func (cabr *CibAddrBlockReader) GetAllCibAddresses() []types.Paddr {
	return cabr.cibAddrBlock.CabCibAddr
}

// HasValidAddresses returns true if this block contains any CIB addresses
// Indicates whether this address block is actively managing any CIBs
func (cabr *CibAddrBlockReader) HasValidAddresses() bool {
	return cabr.cibAddrBlock.CabCibCount > 0
}

// FindAddressIndex returns the index of a specific address, or -1 if not found
// Useful for locating a specific CIB within this address block
func (cabr *CibAddrBlockReader) FindAddressIndex(targetAddr types.Paddr) int {
	for i, addr := range cabr.cibAddrBlock.CabCibAddr {
		if addr == targetAddr {
			return i
		}
	}
	return -1
}

// ValidateAddresses checks if all addresses in the block are non-zero
// Returns the count of valid (non-zero) addresses
func (cabr *CibAddrBlockReader) ValidateAddresses() (uint32, []types.Paddr) {
	var validCount uint32
	var invalidAddresses []types.Paddr

	for _, addr := range cabr.cibAddrBlock.CabCibAddr {
		if addr == 0 {
			invalidAddresses = append(invalidAddresses, addr)
		} else {
			validCount++
		}
	}

	return validCount, invalidAddresses
}

// GetAddressRange returns a slice of addresses within the specified range
// Provides efficient access to a subset of CIB addresses
func (cabr *CibAddrBlockReader) GetAddressRange(startIndex, count uint32) ([]types.Paddr, error) {
	if startIndex >= cabr.cibAddrBlock.CabCibCount {
		return nil, fmt.Errorf("start index %d out of range (have %d addresses)",
			startIndex, cabr.cibAddrBlock.CabCibCount)
	}

	endIndex := startIndex + count
	if endIndex > cabr.cibAddrBlock.CabCibCount {
		endIndex = cabr.cibAddrBlock.CabCibCount
	}

	return cabr.cibAddrBlock.CabCibAddr[startIndex:endIndex], nil
}

// IsEmpty returns true if this address block contains no CIB addresses
// Indicates an unused or uninitialized address block
func (cabr *CibAddrBlockReader) IsEmpty() bool {
	return cabr.cibAddrBlock.CabCibCount == 0
}

// IsFull returns true if this address block has reached its maximum capacity
// The maximum capacity depends on the block size and structure overhead
func (cabr *CibAddrBlockReader) IsFull() bool {
	// Calculate maximum possible addresses based on data size
	// obj_phys_t (32) + index (4) + count (4) = 40 bytes overhead
	// Each address is 8 bytes (paddr_t)
	availableSpace := len(cabr.data) - 40
	maxAddresses := availableSpace / 8
	return int(cabr.cibAddrBlock.CabCibCount) >= maxAddresses
}

// CalculateUtilization returns the percentage of address slots used
// Returns a value between 0.0 (empty) and 100.0 (full)
func (cabr *CibAddrBlockReader) CalculateUtilization() float64 {
	if len(cabr.data) <= 40 {
		return 0.0
	}

	availableSpace := len(cabr.data) - 40
	maxAddresses := availableSpace / 8
	if maxAddresses == 0 {
		return 0.0
	}

	return (float64(cabr.cibAddrBlock.CabCibCount) / float64(maxAddresses)) * 100.0
}