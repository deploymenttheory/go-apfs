package container

import (
	"encoding/binary"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"

	parser "github.com/deploymenttheory/go-apfs/internal/parsers/container"
)

// CreateTestContainerSuperblockData creates test container superblock data for testing
// This is a copy of the function from the parser package to avoid circular imports
func CreateTestContainerSuperblockData(magic uint32, blockSize uint32, blockCount uint64, features uint64, uuid types.UUID, nextOID types.OidT, nextXID types.XidT, maxFileSystems uint32, endian binary.ByteOrder) []byte {
	data := make([]byte, 1500) // Conservative size for container superblock

	// Object header (32 bytes)
	copy(data[0:8], []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) // Checksum
	endian.PutUint64(data[8:16], uint64(1000))                              // OID
	endian.PutUint64(data[16:24], uint64(2000))                             // XID
	endian.PutUint32(data[24:28], types.ObjectTypeNxSuperblock)             // Type
	endian.PutUint32(data[28:32], 0)                                        // Subtype

	// Container superblock fields
	endian.PutUint32(data[32:36], magic)
	endian.PutUint32(data[36:40], blockSize)
	endian.PutUint64(data[40:48], blockCount)
	endian.PutUint64(data[48:56], features)
	endian.PutUint64(data[56:64], 0) // readonly compatible features
	endian.PutUint64(data[64:72], 0) // incompatible features

	// UUID
	copy(data[72:88], uuid[:])

	// Object and transaction IDs
	endian.PutUint64(data[88:96], uint64(nextOID))
	endian.PutUint64(data[96:104], uint64(nextXID))

	// Checkpoint fields
	endian.PutUint32(data[104:108], 100)  // desc blocks
	endian.PutUint32(data[108:112], 200)  // data blocks
	endian.PutUint64(data[112:120], 1000) // desc base
	endian.PutUint64(data[120:128], 2000) // data base
	endian.PutUint32(data[128:132], 10)   // desc next
	endian.PutUint32(data[132:136], 20)   // data next
	endian.PutUint32(data[136:140], 5)    // desc index
	endian.PutUint32(data[140:144], 50)   // desc len
	endian.PutUint32(data[144:148], 15)   // data index
	endian.PutUint32(data[148:152], 150)  // data len

	// Critical OIDs
	endian.PutUint64(data[152:160], 3000) // spaceman OID
	endian.PutUint64(data[160:168], 4000) // omap OID
	endian.PutUint64(data[168:176], 5000) // reaper OID

	// Testing and filesystem management
	endian.PutUint32(data[176:180], 0)              // test type
	endian.PutUint32(data[180:184], maxFileSystems) // max filesystems

	// Volume OIDs array (NxMaxFileSystems * 8 bytes)
	offset := 184
	for i := 0; i < types.NxMaxFileSystems; i++ {
		if i < int(maxFileSystems) && i < 3 { // Respect maxFileSystems parameter and limit to 3 for testing
			endian.PutUint64(data[offset:offset+8], uint64(6000+i))
		} else {
			endian.PutUint64(data[offset:offset+8], 0) // Invalid/empty volume
		}
		offset += 8
	}

	// Counters array (NxNumCounters * 8 bytes)
	for i := 0; i < types.NxNumCounters; i++ {
		endian.PutUint64(data[offset:offset+8], uint64(i*100))
		offset += 8
	}

	// Blocked out range
	endian.PutUint64(data[offset:offset+8], 7000)   // start paddr
	endian.PutUint64(data[offset+8:offset+16], 100) // block count
	offset += 16

	// Remaining fields with test values
	endian.PutUint64(data[offset:offset+8], 8000) // evict mapping tree OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], 0x04) // flags (NxCryptoSw)
	offset += 8
	endian.PutUint64(data[offset:offset+8], 9000) // EFI jumpstart
	offset += 8

	// Fusion UUID (16 bytes)
	for i := 0; i < 16; i++ {
		data[offset+i] = byte(0x10 + i)
	}
	offset += 16

	// Keybag location
	endian.PutUint64(data[offset:offset+8], 10000) // start paddr
	endian.PutUint64(data[offset+8:offset+16], 50) // block count
	offset += 16

	// Ephemeral info array
	for i := 0; i < types.NxEphInfoCount; i++ {
		endian.PutUint64(data[offset:offset+8], uint64(11000+i*100))
		offset += 8
	}

	// Test OID, Fusion OIDs, and cache range
	endian.PutUint64(data[offset:offset+8], 12000) // test OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], 13000) // fusion MT OID
	offset += 8
	endian.PutUint64(data[offset:offset+8], 14000) // fusion WBC OID
	offset += 8

	// Fusion WBC range
	endian.PutUint64(data[offset:offset+8], 15000) // start paddr
	endian.PutUint64(data[offset+8:offset+16], 25) // block count
	offset += 16

	// Final fields
	endian.PutUint64(data[offset:offset+8], 2) // newest mounted version
	offset += 8

	// Media key locker location
	endian.PutUint64(data[offset:offset+8], 16000) // start paddr
	endian.PutUint64(data[offset+8:offset+16], 75) // block count

	return data
}

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

	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
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

	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
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

	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
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

func TestContainerManager_VolumeOperations(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	manager := NewContainerManager(superblockReader, blockReader, objectMapReader)

	// Test ListVolumes
	volumes, err := manager.ListVolumes()
	if err != nil {
		t.Fatalf("ListVolumes() failed: %v", err)
	}

	if len(volumes) == 0 {
		t.Error("ListVolumes() returned no volumes")
	}

	// Test FindVolumeByName with non-existent volume
	_, err = manager.FindVolumeByName("NonExistent")
	if err == nil {
		t.Error("FindVolumeByName() should have failed for non-existent volume")
	}

	// Test FindVolumeByUUID with non-existent UUID
	testUUID := types.UUID{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	_, err = manager.FindVolumeByUUID(testUUID)
	if err == nil {
		t.Error("FindVolumeByUUID() should have failed for non-existent volume")
	}

	// Test FindVolumesByRole
	roleVolumes, err := manager.FindVolumesByRole(0)
	if err != nil {
		t.Fatalf("FindVolumesByRole() failed: %v", err)
	}

	// Should return all volumes since they have default role 0
	if len(roleVolumes) != len(volumes) {
		t.Errorf("FindVolumesByRole() returned %d volumes, want %d", len(roleVolumes), len(volumes))
	}
}

func TestContainerManager_SpaceManagement(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	manager := NewContainerManager(superblockReader, blockReader, objectMapReader)

	// Test space calculations
	totalSize := manager.TotalSize()
	usedSpace := manager.UsedSpace()
	freeSpace := manager.FreeSpace()
	utilization := manager.SpaceUtilization()

	expectedTotalSize := uint64(1000000 * 4096)
	if totalSize != expectedTotalSize {
		t.Errorf("TotalSize() = %d, want %d", totalSize, expectedTotalSize)
	}

	if usedSpace+freeSpace != totalSize {
		t.Errorf("UsedSpace(%d) + FreeSpace(%d) != TotalSize(%d)", usedSpace, freeSpace, totalSize)
	}

	if utilization < 0 || utilization > 100 {
		t.Errorf("SpaceUtilization() = %f, want 0-100", utilization)
	}

	expectedUtilization := (float64(usedSpace) / float64(totalSize)) * 100.0
	if utilization != expectedUtilization {
		t.Errorf("SpaceUtilization() = %f, want %f", utilization, expectedUtilization)
	}
}

func TestContainerManager_EncryptionAndSecurity(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	manager := NewContainerManager(superblockReader, blockReader, objectMapReader)

	// Test encryption detection
	isEncrypted := manager.IsEncrypted()
	// Based on test data with keybag location block count > 0
	if !isEncrypted {
		t.Error("IsEncrypted() should return true for test data with keybag")
	}

	// Test crypto type
	cryptoType := manager.CryptoType()
	// Should return 0 for placeholder implementation
	if cryptoType != 0 {
		t.Errorf("CryptoType() = %d, want 0", cryptoType)
	}
}

func TestContainerManager_FeatureCompatibility(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	manager := NewContainerManager(superblockReader, blockReader, objectMapReader)

	// Test feature methods (should return 0 due to nil feature manager)
	features := manager.Features()
	incompatibleFeatures := manager.IncompatibleFeatures()
	readonlyFeatures := manager.ReadonlyCompatibleFeatures()

	if features != 0 {
		t.Errorf("Features() = %d, want 0", features)
	}

	if incompatibleFeatures != 0 {
		t.Errorf("IncompatibleFeatures() = %d, want 0", incompatibleFeatures)
	}

	if readonlyFeatures != 0 {
		t.Errorf("ReadonlyCompatibleFeatures() = %d, want 0", readonlyFeatures)
	}
}

func TestContainerManager_SnapshotsAndVersioning(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	manager := NewContainerManager(superblockReader, blockReader, objectMapReader)

	// Test snapshot methods (placeholder implementations)
	totalSnapshots := manager.TotalSnapshots()
	latestSnapshotXID := manager.LatestSnapshotXID()

	if totalSnapshots != 0 {
		t.Errorf("TotalSnapshots() = %d, want 0", totalSnapshots)
	}

	if latestSnapshotXID != 0 {
		t.Errorf("LatestSnapshotXID() = %d, want 0", latestSnapshotXID)
	}
}

func TestContainerManager_BlockedSpaceAndRanges(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	data := CreateTestContainerSuperblockData(types.NxMagic, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	manager := NewContainerManager(superblockReader, blockReader, objectMapReader)

	// Test blocked out range
	blockedRange := manager.BlockedOutRange()
	expectedStartAddr := types.Paddr(7000)
	expectedBlockCount := uint64(100)

	if blockedRange.PrStartPaddr != expectedStartAddr {
		t.Errorf("BlockedOutRange start = %d, want %d", blockedRange.PrStartPaddr, expectedStartAddr)
	}

	if blockedRange.PrBlockCount != expectedBlockCount {
		t.Errorf("BlockedOutRange block count = %d, want %d", blockedRange.PrBlockCount, expectedBlockCount)
	}

	// Test evict mapping tree OID
	evictMappingOID := manager.EvictMappingTreeOID()
	expectedEvictOID := types.OidT(8000)

	if evictMappingOID != expectedEvictOID {
		t.Errorf("EvictMappingTreeOID() = %d, want %d", evictMappingOID, expectedEvictOID)
	}
}

func TestContainerManager_CheckIntegrityFailures(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	endian := binary.LittleEndian

	// Test with invalid magic number - the parser will reject this, so we expect an error
	data := CreateTestContainerSuperblockData(0xBADC0DE, 4096, 1000000, 0, uuid, 5000, 6000, 10, endian)
	_, err := parser.NewContainerSuperblockReader(data, endian)
	if err == nil {
		t.Error("NewContainerSuperblockReader() should have failed with invalid magic number")
	}

	// Test with valid magic but other issues
	data = CreateTestContainerSuperblockData(types.NxMagic, 0, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	manager := NewContainerManager(superblockReader, blockReader, objectMapReader)

	isHealthy, issues := manager.CheckIntegrity()
	if isHealthy {
		t.Error("CheckIntegrity() should return false for zero block size")
	}

	if len(issues) == 0 {
		t.Error("CheckIntegrity() should return issues for zero block size")
	}

	foundBlockSizeIssue := false
	for _, issue := range issues {
		if strings.Contains(issue, "Invalid block size") {
			foundBlockSizeIssue = true
			break
		}
	}

	if !foundBlockSizeIssue {
		t.Error("CheckIntegrity() should report invalid block size issue")
	}

	if manager.IsHealthy() {
		t.Error("IsHealthy() should return false for container with issues")
	}
}

func TestContainerManager_ErrorCases(t *testing.T) {
	blockReader := NewMockBlockDeviceReader()
	uuid := types.UUID{}
	endian := binary.LittleEndian

	// Test with zero block size
	data := CreateTestContainerSuperblockData(types.NxMagic, 0, 1000000, 0, uuid, 5000, 6000, 10, endian)
	superblockReader, err := parser.NewContainerSuperblockReader(data, endian)
	if err != nil {
		t.Fatalf("NewContainerSuperblockReader() failed: %v", err)
	}

	objectMapReader := NewMockObjectMapReader()
	manager := NewContainerManager(superblockReader, blockReader, objectMapReader)

	isHealthy, issues := manager.CheckIntegrity()
	if isHealthy {
		t.Error("CheckIntegrity() should return false for zero block size")
	}

	foundBlockSizeIssue := false
	for _, issue := range issues {
		if strings.Contains(issue, "Invalid block size") {
			foundBlockSizeIssue = true
			break
		}
	}

	if !foundBlockSizeIssue {
		t.Error("CheckIntegrity() should report invalid block size issue")
	}
}
