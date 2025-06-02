package datastreams

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// physicalExtentReader implements the PhysicalExtentReader interface
type physicalExtentReader struct {
	key    *types.JPhysExtKeyT
	value  *types.JPhysExtValT
	data   []byte
	endian binary.ByteOrder
}

// NewPhysicalExtentReader creates a new PhysicalExtentReader implementation
func NewPhysicalExtentReader(keyData, valueData []byte, endian binary.ByteOrder) (interfaces.PhysicalExtentReader, error) {
	if len(keyData) < 8 { // JKeyT is 8 bytes
		return nil, fmt.Errorf("key data too small for physical extent key: %d bytes", len(keyData))
	}

	if len(valueData) < 20 { // JPhysExtValT is 20 bytes (8 + 8 + 4)
		return nil, fmt.Errorf("value data too small for physical extent value: %d bytes", len(valueData))
	}

	key, err := parsePhysicalExtentKey(keyData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse physical extent key: %w", err)
	}

	value, err := parsePhysicalExtentValue(valueData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse physical extent value: %w", err)
	}

	return &physicalExtentReader{
		key:    key,
		value:  value,
		data:   append(keyData, valueData...),
		endian: endian,
	}, nil
}

// parsePhysicalExtentKey parses raw bytes into a JPhysExtKeyT structure
func parsePhysicalExtentKey(data []byte, endian binary.ByteOrder) (*types.JPhysExtKeyT, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("insufficient data for physical extent key")
	}

	key := &types.JPhysExtKeyT{}
	key.Hdr.ObjIdAndType = endian.Uint64(data[0:8])

	return key, nil
}

// parsePhysicalExtentValue parses raw bytes into a JPhysExtValT structure
func parsePhysicalExtentValue(data []byte, endian binary.ByteOrder) (*types.JPhysExtValT, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("insufficient data for physical extent value")
	}

	value := &types.JPhysExtValT{}
	value.LenAndKind = endian.Uint64(data[0:8])
	value.OwningObjId = endian.Uint64(data[8:16])
	value.Refcnt = int32(endian.Uint32(data[16:20]))

	return value, nil
}

// PhysicalBlockAddress returns the physical block address of the start of the extent
func (per *physicalExtentReader) PhysicalBlockAddress() uint64 {
	// The object identifier in the header is the physical block address of the start of the extent
	return per.key.Hdr.ObjIdAndType & types.ObjIdMask
}

// Length returns the length of the extent in blocks
func (per *physicalExtentReader) Length() uint64 {
	// The extent's length is accessed as len_and_kind & PEXT_LEN_MASK
	return per.value.LenAndKind & types.PextLenMask
}

// Kind returns the kind of the extent (e.g., APFS_KIND_NEW)
func (per *physicalExtentReader) Kind() uint8 {
	// The extent's kind is accessed as (len_and_kind & PEXT_KIND_MASK) >> PEXT_KIND_SHIFT
	return uint8((per.value.LenAndKind & types.PextKindMask) >> types.PextKindShift)
}

// OwningObjectID returns the identifier of the file system record using this extent
func (per *physicalExtentReader) OwningObjectID() uint64 {
	return per.value.OwningObjId
}

// ReferenceCount returns the reference count for this extent
func (per *physicalExtentReader) ReferenceCount() int32 {
	return per.value.Refcnt
}

// IsKindNew checks if the extent kind is APFS_KIND_NEW
func (per *physicalExtentReader) IsKindNew() bool {
	return per.Kind() == uint8(types.ApfsKindNew)
}

// IsKindUpdate checks if the extent kind is APFS_KIND_UPDATE
func (per *physicalExtentReader) IsKindUpdate() bool {
	return per.Kind() == uint8(types.ApfsKindUpdate)
}

// IsKindDead checks if the extent kind is APFS_KIND_DEAD
func (per *physicalExtentReader) IsKindDead() bool {
	return per.Kind() == uint8(types.ApfsKindDead)
}

// IsValidKind checks if the extent kind is valid for on-disk storage
func (per *physicalExtentReader) IsValidKind() bool {
	kind := per.Kind()
	// Only APFS_KIND_NEW and APFS_KIND_UPDATE are valid on disk
	return kind == uint8(types.ApfsKindNew) || kind == uint8(types.ApfsKindUpdate)
}

// GetKindDescription returns a human-readable description of the extent kind
func (per *physicalExtentReader) GetKindDescription() string {
	switch per.Kind() {
	case uint8(types.ApfsKindAny):
		return "Any (invalid on disk)"
	case uint8(types.ApfsKindNew):
		return "New"
	case uint8(types.ApfsKindUpdate):
		return "Update"
	case uint8(types.ApfsKindDead):
		return "Dead (invalid on disk)"
	case uint8(types.ApfsKindUpdateRefcnt):
		return "Update Reference Count (invalid on disk)"
	case uint8(types.ApfsKindInvalid):
		return "Invalid"
	default:
		return fmt.Sprintf("Unknown (%d)", per.Kind())
	}
}

// CanBeDeleted checks if the extent can be deleted (reference count is zero)
func (per *physicalExtentReader) CanBeDeleted() bool {
	return per.value.Refcnt <= 0
}

// IsShared checks if the extent is shared (reference count > 1)
func (per *physicalExtentReader) IsShared() bool {
	return per.value.Refcnt > 1
}

// SizeInBytes returns the extent size in bytes given a block size
func (per *physicalExtentReader) SizeInBytes(blockSize uint32) uint64 {
	return per.Length() * uint64(blockSize)
}

// EndBlockAddress returns the physical block address of the end of the extent (exclusive)
func (per *physicalExtentReader) EndBlockAddress() uint64 {
	return per.PhysicalBlockAddress() + per.Length()
}

// ContainsBlock checks if the extent contains the given physical block address
func (per *physicalExtentReader) ContainsBlock(blockAddress uint64) bool {
	start := per.PhysicalBlockAddress()
	end := per.EndBlockAddress()
	return blockAddress >= start && blockAddress < end
}

// Validate performs validation checks on the physical extent
func (per *physicalExtentReader) Validate() error {
	// Check if the extent has zero length
	if per.Length() == 0 {
		return fmt.Errorf("physical extent has zero length")
	}

	// Check if the kind is valid for on-disk storage
	if !per.IsValidKind() {
		return fmt.Errorf("physical extent has invalid kind for disk storage: %s", per.GetKindDescription())
	}

	// Check if reference count is valid
	if per.ReferenceCount() < 0 {
		return fmt.Errorf("physical extent has negative reference count: %d", per.ReferenceCount())
	}

	// Check for potential overflow in end address calculation
	start := per.PhysicalBlockAddress()
	length := per.Length()
	if start > ^uint64(0)-length {
		return fmt.Errorf("physical extent end address would overflow: start=%d, length=%d", start, length)
	}

	return nil
}
