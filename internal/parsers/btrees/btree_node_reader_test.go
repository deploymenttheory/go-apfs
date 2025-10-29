package btrees

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestBTreeNodeData creates test B-tree node data with valid Fletcher-64 checksum
func createTestBTreeNodeData(oid types.OidT, xid types.XidT, objType, objSubtype uint32, flags uint16, level uint16, nkeys uint32, endian binary.ByteOrder) []byte {
	// Pad to multiple of 4 bytes for Fletcher-64
	size := 100
	if size%4 != 0 {
		size = ((size / 4) + 1) * 4
	}
	data := make([]byte, size)

	// Object header (32 bytes) - start with zero checksum
	// Checksum at 0:8 will be filled later
	endian.PutUint64(data[8:16], uint64(oid))
	endian.PutUint64(data[16:24], uint64(xid))
	endian.PutUint32(data[24:28], objType)
	endian.PutUint32(data[28:32], objSubtype)

	// B-tree node specific fields (24 bytes)
	endian.PutUint16(data[32:34], flags)
	endian.PutUint16(data[34:36], level)
	endian.PutUint32(data[36:40], nkeys)

	// Table space location
	endian.PutUint16(data[40:42], 100) // offset
	endian.PutUint16(data[42:44], 200) // length

	// Free space location
	endian.PutUint16(data[44:46], 300)
	endian.PutUint16(data[46:48], 150)

	// Key free list location
	endian.PutUint16(data[48:50], 400)
	endian.PutUint16(data[50:52], 50)

	// Value free list location
	endian.PutUint16(data[52:54], 500)
	endian.PutUint16(data[54:56], 75)

	// Add some node data
	for i := 56; i < len(data); i++ {
		data[i] = byte(i - 56)
	}

	// Calculate and set the Fletcher-64 checksum
	checksum := calculateFletcherChecksum(data)
	copy(data[0:8], checksum[:])

	return data
}

// calculateFletcherChecksum calculates Fletcher-64 checksum for test data
// This is a simplified version for testing purposes
func calculateFletcherChecksum(data []byte) [8]byte {
	const maxUint32 = uint64(0xFFFFFFFF)
	const chunkSize = 1024

	var sum1, sum2 uint64

	// Make a copy with zeroed checksum field
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	for i := 0; i < 8; i++ {
		dataCopy[i] = 0
	}

	// Process data in chunks of 32-bit words
	for offset := 0; offset < len(dataCopy); offset += chunkSize * 4 {
		chunkEnd := offset + chunkSize*4
		if chunkEnd > len(dataCopy) {
			chunkEnd = len(dataCopy)
		}

		for i := offset; i < chunkEnd; i += 4 {
			if i+4 > len(dataCopy) {
				break
			}
			word := binary.LittleEndian.Uint32(dataCopy[i : i+4])
			sum1 += uint64(word)
			sum2 += sum1
		}

		sum1 %= maxUint32
		sum2 %= maxUint32
	}

	// Calculate final checksum
	ckLow := maxUint32 - ((sum1 + sum2) % maxUint32)
	ckHigh := maxUint32 - ((sum1 + ckLow) % maxUint32)
	result := ckLow | (ckHigh << 32)

	var checksum [8]byte
	binary.LittleEndian.PutUint64(checksum[:], result)
	return checksum
}

// TestBTreeNodeReader tests all B-tree node reader method implementations
func TestBTreeNodeReader(t *testing.T) {
	testCases := []struct {
		name              string
		oid               types.OidT
		xid               types.XidT
		objType           uint32
		objSubtype        uint32
		flags             uint16
		level             uint16
		nkeys             uint32
		expectedIsRoot    bool
		expectedIsLeaf    bool
		expectedFixedKV   bool
		expectedIsHashed  bool
		expectedHasHeader bool
	}{
		{
			name:              "Root Leaf Node",
			oid:               12345,
			xid:               67890,
			objType:           types.ObjectTypeBtreeNode,
			objSubtype:        0,
			flags:             types.BtnodeRoot | types.BtnodeLeaf,
			level:             0,
			nkeys:             10,
			expectedIsRoot:    true,
			expectedIsLeaf:    true,
			expectedFixedKV:   false,
			expectedIsHashed:  false,
			expectedHasHeader: true,
		},
		{
			name:              "Internal Node with Fixed KV",
			oid:               54321,
			xid:               98765,
			objType:           types.ObjectTypeBtreeNode,
			objSubtype:        1,
			flags:             types.BtnodeFixedKvSize,
			level:             2,
			nkeys:             50,
			expectedIsRoot:    false,
			expectedIsLeaf:    false,
			expectedFixedKV:   true,
			expectedIsHashed:  false,
			expectedHasHeader: true,
		},
		{
			name:              "Hashed Node without Header",
			oid:               11111,
			xid:               22222,
			objType:           types.ObjectTypeBtreeNode,
			objSubtype:        2,
			flags:             types.BtnodeHashed | types.BtnodeNoheader,
			level:             1,
			nkeys:             25,
			expectedIsRoot:    false,
			expectedIsLeaf:    false,
			expectedFixedKV:   false,
			expectedIsHashed:  true,
			expectedHasHeader: false,
		},
		{
			name:              "All Flags Node",
			oid:               99999,
			xid:               88888,
			objType:           types.ObjectTypeBtreeNode,
			objSubtype:        3,
			flags:             types.BtnodeRoot | types.BtnodeLeaf | types.BtnodeFixedKvSize | types.BtnodeHashed,
			level:             0,
			nkeys:             1,
			expectedIsRoot:    true,
			expectedIsLeaf:    true,
			expectedFixedKV:   true,
			expectedIsHashed:  true,
			expectedHasHeader: true,
		},
	}

	endian := binary.LittleEndian

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := createTestBTreeNodeData(tc.oid, tc.xid, tc.objType, tc.objSubtype, tc.flags, tc.level, tc.nkeys, endian)
			bnr, err := NewBTreeNodeReader(data, endian)
			if err != nil {
				t.Fatalf("NewBTreeNodeReader() failed: %v", err)
			}

			// Test basic properties
			if flags := bnr.Flags(); flags != tc.flags {
				t.Errorf("Flags() = 0x%04X, want 0x%04X", flags, tc.flags)
			}

			if level := bnr.Level(); level != tc.level {
				t.Errorf("Level() = %d, want %d", level, tc.level)
			}

			if nkeys := bnr.KeyCount(); nkeys != tc.nkeys {
				t.Errorf("KeyCount() = %d, want %d", nkeys, tc.nkeys)
			}

			// Test table space location
			tableSpace := bnr.TableSpace()
			if tableSpace.Off != 100 || tableSpace.Len != 200 {
				t.Errorf("TableSpace() = {Off: %d, Len: %d}, want {Off: 100, Len: 200}", tableSpace.Off, tableSpace.Len)
			}

			// Test free space location
			freeSpace := bnr.FreeSpace()
			if freeSpace.Off != 300 || freeSpace.Len != 150 {
				t.Errorf("FreeSpace() = {Off: %d, Len: %d}, want {Off: 300, Len: 150}", freeSpace.Off, freeSpace.Len)
			}

			// Test key free list location
			keyFreeList := bnr.KeyFreeList()
			if keyFreeList.Off != 400 || keyFreeList.Len != 50 {
				t.Errorf("KeyFreeList() = {Off: %d, Len: %d}, want {Off: 400, Len: 50}", keyFreeList.Off, keyFreeList.Len)
			}

			// Test value free list location
			valueFreeList := bnr.ValueFreeList()
			if valueFreeList.Off != 500 || valueFreeList.Len != 75 {
				t.Errorf("ValueFreeList() = {Off: %d, Len: %d}, want {Off: 500, Len: 75}", valueFreeList.Off, valueFreeList.Len)
			}

			// Test node data
			nodeData := bnr.Data()
			if len(nodeData) != 44 { // 100 - 56 header bytes
				t.Errorf("Data() length = %d, want %d", len(nodeData), 44)
			}

			// Verify data content
			for i := 0; i < len(nodeData) && i < 10; i++ {
				if nodeData[i] != byte(i) {
					t.Errorf("Data()[%d] = %d, want %d", i, nodeData[i], byte(i))
				}
			}

			// Test flag-based properties
			if isRoot := bnr.IsRoot(); isRoot != tc.expectedIsRoot {
				t.Errorf("IsRoot() = %v, want %v", isRoot, tc.expectedIsRoot)
			}

			if isLeaf := bnr.IsLeaf(); isLeaf != tc.expectedIsLeaf {
				t.Errorf("IsLeaf() = %v, want %v", isLeaf, tc.expectedIsLeaf)
			}

			if fixedKV := bnr.HasFixedKVSize(); fixedKV != tc.expectedFixedKV {
				t.Errorf("HasFixedKVSize() = %v, want %v", fixedKV, tc.expectedFixedKV)
			}

			if isHashed := bnr.IsHashed(); isHashed != tc.expectedIsHashed {
				t.Errorf("IsHashed() = %v, want %v", isHashed, tc.expectedIsHashed)
			}

			if hasHeader := bnr.HasHeader(); hasHeader != tc.expectedHasHeader {
				t.Errorf("HasHeader() = %v, want %v", hasHeader, tc.expectedHasHeader)
			}
		})
	}
}

// TestBTreeNodeReader_ErrorCases tests error conditions
func TestBTreeNodeReader_ErrorCases(t *testing.T) {
	endian := binary.LittleEndian

	testCases := []struct {
		name     string
		dataSize int
	}{
		{"Empty data", 0},
		{"Too small data", 20},
		{"Minimum size - 1", 55},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, tc.dataSize)
			_, err := NewBTreeNodeReader(data, endian)
			if err == nil {
				t.Error("NewBTreeNodeReader() should have failed with insufficient data")
			}
		})
	}
}

// TestBTreeNodeReader_MinimumSize tests with minimum valid size
func TestBTreeNodeReader_MinimumSize(t *testing.T) {
	endian := binary.LittleEndian
	data := createTestBTreeNodeData(1, 2, types.ObjectTypeBtreeNode, 0, types.BtnodeLeaf, 0, 5, endian)

	// Trim to minimum size
	data = data[:56]

	// Recalculate checksum for the trimmed data
	checksum := calculateFletcherChecksum(data)
	copy(data[0:8], checksum[:])

	bnr, err := NewBTreeNodeReader(data, endian)
	if err != nil {
		t.Fatalf("NewBTreeNodeReader() failed with minimum size data: %v", err)
	}

	if bnr == nil {
		t.Error("NewBTreeNodeReader() returned nil with valid data")
	}

	// Should have no node data
	if len(bnr.Data()) != 0 {
		t.Errorf("Data() length = %d, want 0 for minimum size", len(bnr.Data()))
	}
}

// TestBTreeNodeReader_EndianHandling tests both little and big endian
func TestBTreeNodeReader_EndianHandling(t *testing.T) {
	testCases := []struct {
		name   string
		endian binary.ByteOrder
	}{
		{"Little Endian", binary.LittleEndian},
		{"Big Endian", binary.BigEndian},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := createTestBTreeNodeData(0x1234567890ABCDEF, 0xFEDCBA0987654321, 0x12345678, 0x87654321, 0xABCD, 42, 1000, tc.endian)
			bnr, err := NewBTreeNodeReader(data, tc.endian)
			if err != nil {
				t.Fatalf("NewBTreeNodeReader() failed: %v", err)
			}

			// Values should be parsed correctly regardless of endianness
			if flags := bnr.Flags(); flags != 0xABCD {
				t.Errorf("Flags() = 0x%04X, want 0xABCD", flags)
			}

			if level := bnr.Level(); level != 42 {
				t.Errorf("Level() = %d, want 42", level)
			}

			if nkeys := bnr.KeyCount(); nkeys != 1000 {
				t.Errorf("KeyCount() = %d, want 1000", nkeys)
			}
		})
	}
}

// Benchmark B-tree node reader methods
func BenchmarkBTreeNodeReader(b *testing.B) {
	endian := binary.LittleEndian
	data := createTestBTreeNodeData(12345, 67890, types.ObjectTypeBtreeNode, 0, types.BtnodeLeaf, 0, 100, endian)
	bnr, _ := NewBTreeNodeReader(data, endian)

	b.Run("Flags", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bnr.Flags()
		}
	})

	b.Run("IsLeaf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bnr.IsLeaf()
		}
	})

	b.Run("KeyCount", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bnr.KeyCount()
		}
	})

	b.Run("Data", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bnr.Data()
		}
	})
}
