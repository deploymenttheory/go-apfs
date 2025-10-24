package spacemanager

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestNewCibAddrBlockReader(t *testing.T) {
	// Create CIB address block data with 3 addresses
	addrCount := uint32(3)
	addrSize := 8 // Size of paddr_t
	dataSize := 40 + int(addrCount)*addrSize // Header + addresses
	data := make([]byte, dataSize)
	endian := binary.LittleEndian

	// Fill object header
	endian.PutUint64(data[8:16], 789)                             // OOid
	endian.PutUint64(data[16:24], 101112)                         // OXid
	endian.PutUint32(data[24:28], types.ObjectTypeSpacemanCab)    // OType
	endian.PutUint32(data[28:32], 0)                              // OSubtype

	// Fill CIB address block specific fields
	endian.PutUint32(data[32:36], 7)         // CabIndex
	endian.PutUint32(data[36:40], addrCount) // CabCibCount

	// Fill addresses at offset 40
	addresses := []uint64{0x1000, 0x2000, 0x3000}
	offset := 40
	for i, addr := range addresses {
		endian.PutUint64(data[offset+i*8:offset+(i+1)*8], addr)
	}

	reader, err := NewCibAddrBlockReader(data, endian)
	if err != nil {
		t.Fatalf("NewCibAddrBlockReader() failed: %v", err)
	}

	if reader.Index() != 7 {
		t.Errorf("Index() = %d, want 7", reader.Index())
	}

	if reader.CibCount() != addrCount {
		t.Errorf("CibCount() = %d, want %d", reader.CibCount(), addrCount)
	}

	// Test individual addresses
	for i, expectedAddr := range addresses {
		addr, err := reader.GetCibAddress(uint32(i))
		if err != nil {
			t.Fatalf("GetCibAddress(%d) failed: %v", i, err)
		}

		if addr != types.Paddr(expectedAddr) {
			t.Errorf("GetCibAddress(%d) = 0x%x, want 0x%x", i, addr, expectedAddr)
		}
	}
}

func TestCibAddrBlockReader_InvalidType(t *testing.T) {
	data := make([]byte, 64)
	endian := binary.LittleEndian

	// Set invalid object type
	endian.PutUint32(data[24:28], types.ObjectTypeNxSuperblock) // Wrong type

	_, err := NewCibAddrBlockReader(data, endian)
	if err == nil {
		t.Error("NewCibAddrBlockReader() should have failed with invalid object type")
	}
}

func TestCibAddrBlockReader_TooSmall(t *testing.T) {
	data := make([]byte, 35) // Too small (need at least 40)

	_, err := NewCibAddrBlockReader(data, binary.LittleEndian)
	if err == nil {
		t.Error("NewCibAddrBlockReader() should have failed with too small data")
	}
}

func TestCibAddrBlockReader_InsufficientDataForAddresses(t *testing.T) {
	data := make([]byte, 50) // Header + partial addresses
	endian := binary.LittleEndian

	// Set object header
	endian.PutUint32(data[24:28], types.ObjectTypeSpacemanCab)

	// Set count that requires more data than available
	endian.PutUint32(data[36:40], 5) // CabCibCount = 5, but not enough data

	_, err := NewCibAddrBlockReader(data, endian)
	if err == nil {
		t.Error("NewCibAddrBlockReader() should have failed with insufficient data for addresses")
	}
}

func TestCibAddrBlockReader_GetCibAddressOutOfRange(t *testing.T) {
	data := make([]byte, 48) // Header + 1 address
	endian := binary.LittleEndian

	// Set object header
	endian.PutUint32(data[24:28], types.ObjectTypeSpacemanCab)
	endian.PutUint32(data[36:40], 1) // CabCibCount = 1

	reader, err := NewCibAddrBlockReader(data, endian)
	if err != nil {
		t.Fatalf("NewCibAddrBlockReader() failed: %v", err)
	}

	// Try to access index 5 when only 1 address exists
	_, err = reader.GetCibAddress(5)
	if err == nil {
		t.Error("GetCibAddress(5) should have failed with out of range index")
	}
}

func TestCibAddrBlockReader_GetAllCibAddresses(t *testing.T) {
	addrCount := uint32(4)
	dataSize := 40 + int(addrCount)*8
	data := make([]byte, dataSize)
	endian := binary.LittleEndian

	// Set object header
	endian.PutUint32(data[24:28], types.ObjectTypeSpacemanCab)
	endian.PutUint32(data[36:40], addrCount) // CabCibCount

	// Fill addresses
	addresses := []uint64{0x1000, 0x2000, 0x3000, 0x4000}
	offset := 40
	for i, addr := range addresses {
		endian.PutUint64(data[offset+i*8:offset+(i+1)*8], addr)
	}

	reader, err := NewCibAddrBlockReader(data, endian)
	if err != nil {
		t.Fatalf("NewCibAddrBlockReader() failed: %v", err)
	}

	allAddresses := reader.GetAllCibAddresses()
	if len(allAddresses) != int(addrCount) {
		t.Errorf("GetAllCibAddresses() returned %d addresses, want %d", len(allAddresses), addrCount)
	}

	for i, expectedAddr := range addresses {
		if allAddresses[i] != types.Paddr(expectedAddr) {
			t.Errorf("GetAllCibAddresses()[%d] = 0x%x, want 0x%x", i, allAddresses[i], expectedAddr)
		}
	}
}

func TestCibAddrBlockReader_HasValidAddresses(t *testing.T) {
	tests := []struct {
		name      string
		addrCount uint32
		expected  bool
	}{
		{"With addresses", 3, true},
		{"Without addresses", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataSize := 40 + int(tt.addrCount)*8
			data := make([]byte, dataSize)
			endian := binary.LittleEndian

			// Set object header
			endian.PutUint32(data[24:28], types.ObjectTypeSpacemanCab)
			endian.PutUint32(data[36:40], tt.addrCount) // CabCibCount

			reader, err := NewCibAddrBlockReader(data, endian)
			if err != nil {
				t.Fatalf("NewCibAddrBlockReader() failed: %v", err)
			}

			if reader.HasValidAddresses() != tt.expected {
				t.Errorf("HasValidAddresses() = %v, want %v", reader.HasValidAddresses(), tt.expected)
			}
		})
	}
}

func TestCibAddrBlockReader_FindAddressIndex(t *testing.T) {
	addrCount := uint32(3)
	dataSize := 40 + int(addrCount)*8
	data := make([]byte, dataSize)
	endian := binary.LittleEndian

	// Set object header
	endian.PutUint32(data[24:28], types.ObjectTypeSpacemanCab)
	endian.PutUint32(data[36:40], addrCount) // CabCibCount

	// Fill addresses
	addresses := []uint64{0x1000, 0x2000, 0x3000}
	offset := 40
	for i, addr := range addresses {
		endian.PutUint64(data[offset+i*8:offset+(i+1)*8], addr)
	}

	reader, err := NewCibAddrBlockReader(data, endian)
	if err != nil {
		t.Fatalf("NewCibAddrBlockReader() failed: %v", err)
	}

	// Test finding existing addresses
	for i, addr := range addresses {
		index := reader.FindAddressIndex(types.Paddr(addr))
		if index != i {
			t.Errorf("FindAddressIndex(0x%x) = %d, want %d", addr, index, i)
		}
	}

	// Test finding non-existing address
	index := reader.FindAddressIndex(types.Paddr(0x9999))
	if index != -1 {
		t.Errorf("FindAddressIndex(0x9999) = %d, want -1", index)
	}
}

func TestCibAddrBlockReader_ValidateAddresses(t *testing.T) {
	addrCount := uint32(4)
	dataSize := 40 + int(addrCount)*8
	data := make([]byte, dataSize)
	endian := binary.LittleEndian

	// Set object header
	endian.PutUint32(data[24:28], types.ObjectTypeSpacemanCab)
	endian.PutUint32(data[36:40], addrCount) // CabCibCount

	// Fill addresses with some zeros (invalid)
	addresses := []uint64{0x1000, 0, 0x3000, 0} // 2 valid, 2 invalid
	offset := 40
	for i, addr := range addresses {
		endian.PutUint64(data[offset+i*8:offset+(i+1)*8], addr)
	}

	reader, err := NewCibAddrBlockReader(data, endian)
	if err != nil {
		t.Fatalf("NewCibAddrBlockReader() failed: %v", err)
	}

	validCount, invalidAddresses := reader.ValidateAddresses()
	expectedValidCount := uint32(2)
	expectedInvalidCount := 2

	if validCount != expectedValidCount {
		t.Errorf("ValidateAddresses() valid count = %d, want %d", validCount, expectedValidCount)
	}

	if len(invalidAddresses) != expectedInvalidCount {
		t.Errorf("ValidateAddresses() invalid count = %d, want %d", len(invalidAddresses), expectedInvalidCount)
	}
}

func TestCibAddrBlockReader_GetAddressRange(t *testing.T) {
	addrCount := uint32(5)
	dataSize := 40 + int(addrCount)*8
	data := make([]byte, dataSize)
	endian := binary.LittleEndian

	// Set object header
	endian.PutUint32(data[24:28], types.ObjectTypeSpacemanCab)
	endian.PutUint32(data[36:40], addrCount) // CabCibCount

	// Fill addresses
	addresses := []uint64{0x1000, 0x2000, 0x3000, 0x4000, 0x5000}
	offset := 40
	for i, addr := range addresses {
		endian.PutUint64(data[offset+i*8:offset+(i+1)*8], addr)
	}

	reader, err := NewCibAddrBlockReader(data, endian)
	if err != nil {
		t.Fatalf("NewCibAddrBlockReader() failed: %v", err)
	}

	// Test getting range [1:3]
	rangeAddrs, err := reader.GetAddressRange(1, 2)
	if err != nil {
		t.Fatalf("GetAddressRange(1, 2) failed: %v", err)
	}

	expectedRange := []types.Paddr{0x2000, 0x3000}
	if len(rangeAddrs) != len(expectedRange) {
		t.Errorf("GetAddressRange(1, 2) returned %d addresses, want %d", len(rangeAddrs), len(expectedRange))
	}

	for i, expected := range expectedRange {
		if rangeAddrs[i] != expected {
			t.Errorf("GetAddressRange(1, 2)[%d] = 0x%x, want 0x%x", i, rangeAddrs[i], expected)
		}
	}

	// Test out of range start index
	_, err = reader.GetAddressRange(10, 1)
	if err == nil {
		t.Error("GetAddressRange(10, 1) should have failed with out of range start index")
	}
}

func TestCibAddrBlockReader_IsEmptyAndIsFull(t *testing.T) {
	// Test empty block
	emptyData := make([]byte, 40)
	endian := binary.LittleEndian
	endian.PutUint32(emptyData[24:28], types.ObjectTypeSpacemanCab)
	endian.PutUint32(emptyData[36:40], 0) // CabCibCount = 0

	emptyReader, err := NewCibAddrBlockReader(emptyData, endian)
	if err != nil {
		t.Fatalf("NewCibAddrBlockReader() failed: %v", err)
	}

	if !emptyReader.IsEmpty() {
		t.Error("IsEmpty() should be true for empty block")
	}

	// Test block with maximum addresses
	maxAddrs := (len(emptyData) - 40) / 8 // Calculate max possible addresses
	fullDataSize := 40 + maxAddrs*8
	fullData := make([]byte, fullDataSize)
	endian.PutUint32(fullData[24:28], types.ObjectTypeSpacemanCab)
	endian.PutUint32(fullData[36:40], uint32(maxAddrs)) // CabCibCount = max

	fullReader, err := NewCibAddrBlockReader(fullData, endian)
	if err != nil {
		t.Fatalf("NewCibAddrBlockReader() failed: %v", err)
	}

	if !fullReader.IsFull() {
		t.Error("IsFull() should be true for full block")
	}
}

func TestCibAddrBlockReader_CalculateUtilization(t *testing.T) {
	dataSize := 40 + 4*8 // Header + space for 4 addresses
	data := make([]byte, dataSize)
	endian := binary.LittleEndian

	// Set object header with 2 addresses (50% utilization)
	endian.PutUint32(data[24:28], types.ObjectTypeSpacemanCab)
	endian.PutUint32(data[36:40], 2) // CabCibCount = 2 out of 4 possible

	reader, err := NewCibAddrBlockReader(data, endian)
	if err != nil {
		t.Fatalf("NewCibAddrBlockReader() failed: %v", err)
	}

	utilization := reader.CalculateUtilization()
	expected := 50.0 // 2/4 * 100 = 50%
	if utilization != expected {
		t.Errorf("CalculateUtilization() = %.1f, want %.1f", utilization, expected)
	}
}