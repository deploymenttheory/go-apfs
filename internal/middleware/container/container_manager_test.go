package container

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// MockBlockDeviceReader implements the BlockDeviceReader interface for testing
type MockBlockDeviceReader struct {
	blocks map[types.Paddr][]byte
}

func NewMockBlockDeviceReader() *MockBlockDeviceReader {
	return &MockBlockDeviceReader{
		blocks: make(map[types.Paddr][]byte),
	}
}

func (m *MockBlockDeviceReader) ReadBlock(address types.Paddr) ([]byte, error) {
	if data, exists := m.blocks[address]; exists {
		return data, nil
	}
	return nil, fmt.Errorf("block not found at address %d", address)
}

func (m *MockBlockDeviceReader) ReadBlockRange(start types.Paddr, count uint32) ([]byte, error) {
	var result []byte
	for i := uint32(0); i < count; i++ {
		block, err := m.ReadBlock(start + types.Paddr(i))
		if err != nil {
			return nil, err
		}
		result = append(result, block...)
	}
	return result, nil
}

func (m *MockBlockDeviceReader) ReadBytes(address types.Paddr, offset uint32, length uint32) ([]byte, error) {
	block, err := m.ReadBlock(address)
	if err != nil {
		return nil, err
	}
	if int(offset+length) > len(block) {
		return nil, fmt.Errorf("read beyond block boundary")
	}
	return block[offset : offset+length], nil
}

func (m *MockBlockDeviceReader) BlockSize() uint32 {
	return 4096
}

func (m *MockBlockDeviceReader) TotalBlocks() uint64 {
	return uint64(len(m.blocks))
}

func (m *MockBlockDeviceReader) TotalSize() uint64 {
	return uint64(len(m.blocks)) * uint64(m.BlockSize())
}

func (m *MockBlockDeviceReader) IsValidAddress(address types.Paddr) bool {
	_, exists := m.blocks[address]
	return exists
}

func (m *MockBlockDeviceReader) CanReadRange(start types.Paddr, count uint32) bool {
	for i := uint32(0); i < count; i++ {
		if !m.IsValidAddress(start + types.Paddr(i)) {
			return false
		}
	}
	return true
}

func (m *MockBlockDeviceReader) SetBlock(address types.Paddr, data []byte) {
	m.blocks[address] = data
}

// MockObjectMapReader for testing
type MockObjectMapReader struct {
	flags             uint32
	treeOID           types.OidT
	snapshotTreeOID   types.OidT
	mostRecentSnapXID types.XidT
}

func NewMockObjectMapReader() *MockObjectMapReader {
	return &MockObjectMapReader{
		flags:             0,
		treeOID:           types.OidT(1000),
		snapshotTreeOID:   types.OidT(2000),
		mostRecentSnapXID: types.XidT(100),
	}
}

func (m *MockObjectMapReader) Flags() uint32                     { return m.flags }
func (m *MockObjectMapReader) SnapshotCount() uint32             { return 5 }
func (m *MockObjectMapReader) TreeType() uint32                  { return types.ObjectTypeBtree }
func (m *MockObjectMapReader) SnapshotTreeType() uint32          { return types.ObjectTypeBtree }
func (m *MockObjectMapReader) TreeOID() types.OidT               { return m.treeOID }
func (m *MockObjectMapReader) SnapshotTreeOID() types.OidT       { return m.snapshotTreeOID }
func (m *MockObjectMapReader) MostRecentSnapshotXID() types.XidT { return m.mostRecentSnapXID }
func (m *MockObjectMapReader) PendingRevertMinXID() types.XidT   { return 0 }
func (m *MockObjectMapReader) PendingRevertMaxXID() types.XidT   { return 0 }

func TestNewContainerManager(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	data := createTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()

	manager := NewContainerManager(superblockReader, blockReader, objectMapReader)
	if manager == nil {
		t.Fatal("NewContainerManager() returned nil")
	}
}

func TestContainerManager_ObjectMapOperations(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{}
	endian := binary.LittleEndian

	data := createTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	manager := NewContainerManager(superblockReader, blockReader, objectMapReader)

	// Test GetObjectMap
	omap, err := manager.GetObjectMap()
	if err != nil {
		t.Fatalf("GetObjectMap() failed: %v", err)
	}
	if omap == nil {
		t.Fatal("GetObjectMap() returned nil object map")
	}

	// Test ResolveVirtualObject (should return not implemented error for now)
	_, err = manager.ResolveVirtualObject(1234, 5678)
	if err == nil {
		t.Error("ResolveVirtualObject() should return error for unimplemented functionality")
	}
}

func TestContainerManager_BasicFunctionality(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	data := createTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	manager := NewContainerManager(superblockReader, blockReader, objectMapReader)

	// Test container metadata
	if containerUUID := manager.UUID(); containerUUID != uuid {
		t.Errorf("UUID() = %v, want %v", containerUUID, uuid)
	}

	if nextOID := manager.NextObjectID(); nextOID != 5000 {
		t.Errorf("NextObjectID() = %d, want 5000", nextOID)
	}

	if nextXID := manager.NextTransactionID(); nextXID != 6000 {
		t.Errorf("NextTransactionID() = %d, want 6000", nextXID)
	}

	// Test space management
	expectedTotalSize := uint64(1000000 * 4096)
	if totalSize := manager.TotalSize(); totalSize != expectedTotalSize {
		t.Errorf("TotalSize() = %d, want %d", totalSize, expectedTotalSize)
	}

	// Test health check
	if !manager.IsHealthy() {
		t.Error("IsHealthy() should return true for valid container")
	}
}
