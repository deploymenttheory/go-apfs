package btrees

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// MockBTreeNode is a mock implementation of BTreeNodeReader for testing
type MockBTreeNode struct {
	flags     uint16
	level     uint16
	keyCount  uint32
	data      []byte
	tableOff  uint16
	tableLen  uint16
	freeOff   uint16
	freeLen   uint16
	keyFreOff uint16
	keyFreLen uint16
	valFreOff uint16
	valFreLen uint16
}

func (m *MockBTreeNode) Flags() uint16    { return m.flags }
func (m *MockBTreeNode) Level() uint16    { return m.level }
func (m *MockBTreeNode) KeyCount() uint32 { return m.keyCount }
func (m *MockBTreeNode) TableSpace() types.NlocT {
	return types.NlocT{Off: m.tableOff, Len: m.tableLen}
}
func (m *MockBTreeNode) FreeSpace() types.NlocT { return types.NlocT{Off: m.freeOff, Len: m.freeLen} }
func (m *MockBTreeNode) KeyFreeList() types.NlocT {
	return types.NlocT{Off: m.keyFreOff, Len: m.keyFreLen}
}
func (m *MockBTreeNode) ValueFreeList() types.NlocT {
	return types.NlocT{Off: m.valFreOff, Len: m.valFreLen}
}
func (m *MockBTreeNode) Data() []byte         { return m.data }
func (m *MockBTreeNode) IsRoot() bool         { return m.flags&types.BtnodeRoot != 0 }
func (m *MockBTreeNode) IsLeaf() bool         { return m.flags&types.BtnodeLeaf != 0 }
func (m *MockBTreeNode) HasFixedKVSize() bool { return m.flags&types.BtnodeFixedKvSize != 0 }
func (m *MockBTreeNode) IsHashed() bool       { return m.flags&types.BtnodeHashed != 0 }
func (m *MockBTreeNode) HasHeader() bool      { return m.flags&types.BtnodeNoheader == 0 }

func TestBinarySearcherFixedSizeOID(t *testing.T) {
	// Create a mock node with 3 entries with OIDs: 10, 20, 30
	searcher := NewBinarySearcher(binary.LittleEndian)

	// Build node data: header (56 bytes) + TOC (12 bytes) + keys (24 bytes) + values (24 bytes)
	nodeData := make([]byte, 56+12+24+24)

	// Set up TOC (3 entries of kvoff_t: 4 bytes each)
	tocOffset := 56
	// Entry 0: key offset 0, value offset 0
	binary.LittleEndian.PutUint16(nodeData[tocOffset:], 0)
	binary.LittleEndian.PutUint16(nodeData[tocOffset+2:], 0)
	// Entry 1: key offset 8, value offset 8
	binary.LittleEndian.PutUint16(nodeData[tocOffset+4:], 8)
	binary.LittleEndian.PutUint16(nodeData[tocOffset+6:], 8)
	// Entry 2: key offset 16, value offset 16
	binary.LittleEndian.PutUint16(nodeData[tocOffset+8:], 16)
	binary.LittleEndian.PutUint16(nodeData[tocOffset+10:], 16)

	// Set up keys (OIDs: 10, 20, 30)
	keyOffset := tocOffset + 12
	binary.LittleEndian.PutUint64(nodeData[keyOffset:], 10)
	binary.LittleEndian.PutUint64(nodeData[keyOffset+8:], 20)
	binary.LittleEndian.PutUint64(nodeData[keyOffset+16:], 30)

	// Set up values (child OIDs: 100, 200, 300)
	valueOffset := keyOffset + 24
	binary.LittleEndian.PutUint64(nodeData[valueOffset:], 100)
	binary.LittleEndian.PutUint64(nodeData[valueOffset+8:], 200)
	binary.LittleEndian.PutUint64(nodeData[valueOffset+16:], 300)

	// Create mock node
	node := &MockBTreeNode{
		flags:    types.BtnodeFixedKvSize, // Fixed-size, leaf node
		level:    0,
		keyCount: 3,
		data:     nodeData,
		tableOff: 0,
		tableLen: 12,
		freeOff:  0,
		freeLen:  0,
	}

	tests := []struct {
		name      string
		targetOID uint64
		shouldErr bool
		expectOID uint64
	}{
		{"OID at start", 10, false, 10},
		{"OID in middle", 20, false, 20},
		{"OID at end", 30, false, 30},
		{"OID not found but lower", 5, false, 10},  // Should return first entry
		{"OID not found in middle", 15, false, 10}, // Should return closest lower
		{"OID higher than all", 40, true, 0},       // No entry >= target
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := searcher.FindEntryByOID(node, types.OidT(tt.targetOID))
			if tt.shouldErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.shouldErr && entry != nil {
				if len(entry.KeyData) < 8 {
					t.Errorf("key data too short")
				}
				oid := binary.LittleEndian.Uint64(entry.KeyData[0:8])
				if oid != tt.expectOID {
					t.Errorf("expected OID %d, got %d", tt.expectOID, oid)
				}
			}
		})
	}
}

func TestBinarySearcherCompositeOIDXID(t *testing.T) {
	searcher := NewBinarySearcher(binary.LittleEndian)

	// Create node data with composite keys: [(oid=10, xid=5), (oid=10, xid=10), (oid=20, xid=3)]
	nodeData := make([]byte, 56+12+48+24)

	// Set up TOC (3 entries of kvoff_t)
	tocOffset := 56
	binary.LittleEndian.PutUint16(nodeData[tocOffset:], 0)
	binary.LittleEndian.PutUint16(nodeData[tocOffset+2:], 0)
	binary.LittleEndian.PutUint16(nodeData[tocOffset+4:], 16)
	binary.LittleEndian.PutUint16(nodeData[tocOffset+6:], 8)
	binary.LittleEndian.PutUint16(nodeData[tocOffset+8:], 32)
	binary.LittleEndian.PutUint16(nodeData[tocOffset+10:], 16)

	// Set up keys
	keyOffset := tocOffset + 12
	// Key 0: oid=10, xid=5
	binary.LittleEndian.PutUint64(nodeData[keyOffset:], 10)
	binary.LittleEndian.PutUint64(nodeData[keyOffset+8:], 5)
	// Key 1: oid=10, xid=10
	binary.LittleEndian.PutUint64(nodeData[keyOffset+16:], 10)
	binary.LittleEndian.PutUint64(nodeData[keyOffset+24:], 10)
	// Key 2: oid=20, xid=3
	binary.LittleEndian.PutUint64(nodeData[keyOffset+32:], 20)
	binary.LittleEndian.PutUint64(nodeData[keyOffset+40:], 3)

	// Set up values
	valueOffset := keyOffset + 48
	binary.LittleEndian.PutUint64(nodeData[valueOffset:], 100)
	binary.LittleEndian.PutUint64(nodeData[valueOffset+8:], 200)
	binary.LittleEndian.PutUint64(nodeData[valueOffset+16:], 300)

	node := &MockBTreeNode{
		flags:    types.BtnodeFixedKvSize | types.BtnodeLeaf,
		level:    0,
		keyCount: 3,
		data:     nodeData,
		tableOff: 0,
		tableLen: 12,
	}

	tests := []struct {
		name      string
		targetOID uint64
		targetXID uint64
		shouldErr bool
		expectOID uint64
		expectXID uint64
	}{
		{"Exact match: (10,5)", 10, 5, false, 10, 5},
		{"Exact match: (10,10)", 10, 10, false, 10, 10},
		{"Lookup with higher XID: (10,8)", 10, 8, false, 10, 5}, // Should return highest XID <= 8
		{"Lookup with much higher XID: (10,100)", 10, 100, false, 10, 10},
		{"Different OID: (20,5)", 20, 5, false, 20, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := searcher.FindEntryByOIDAndXID(node, types.OidT(tt.targetOID), types.XidT(tt.targetXID))
			if tt.shouldErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.shouldErr && entry != nil {
				if len(entry.KeyData) < 16 {
					t.Errorf("key data too short for composite key")
				}
				oid := binary.LittleEndian.Uint64(entry.KeyData[0:8])
				xid := binary.LittleEndian.Uint64(entry.KeyData[8:16])
				if oid != tt.expectOID || xid != tt.expectXID {
					t.Errorf("expected (%d,%d), got (%d,%d)", tt.expectOID, tt.expectXID, oid, xid)
				}
			}
		})
	}
}

func TestBinarySearcherEdgeCases(t *testing.T) {
	searcher := NewBinarySearcher(binary.LittleEndian)

	// Test empty node
	nodeData := make([]byte, 56+4) // Header + empty TOC
	node := &MockBTreeNode{
		flags:    types.BtnodeFixedKvSize | types.BtnodeLeaf,
		level:    0,
		keyCount: 0,
		data:     nodeData,
		tableOff: 0,
		tableLen: 0,
	}

	_, err := searcher.FindEntryByOID(node, types.OidT(10))
	if err == nil {
		t.Error("expected error for empty node")
	}

	// Test single entry node
	nodeData2 := make([]byte, 56+4+8+8)
	tocOffset := 56
	binary.LittleEndian.PutUint16(nodeData2[tocOffset:], 0)
	binary.LittleEndian.PutUint16(nodeData2[tocOffset+2:], 0)

	keyOffset := tocOffset + 4
	binary.LittleEndian.PutUint64(nodeData2[keyOffset:], 42)

	valueOffset := keyOffset + 8
	binary.LittleEndian.PutUint64(nodeData2[valueOffset:], 999)

	node2 := &MockBTreeNode{
		flags:    types.BtnodeFixedKvSize | types.BtnodeLeaf,
		level:    0,
		keyCount: 1,
		data:     nodeData2,
		tableOff: 0,
		tableLen: 4,
	}

	entry, err := searcher.FindEntryByOID(node2, types.OidT(42))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if entry == nil {
		t.Error("expected entry for single-entry node")
	}
}

func TestExtractOIDFromKey(t *testing.T) {
	searcher := NewBinarySearcher(binary.LittleEndian)

	// Test valid key
	keyData := make([]byte, 8)
	binary.LittleEndian.PutUint64(keyData, 12345)

	oid, err := searcher.extractOIDFromKey(keyData)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if oid != 12345 {
		t.Errorf("expected OID 12345, got %d", oid)
	}

	// Test short key
	shortKey := make([]byte, 4)
	_, err = searcher.extractOIDFromKey(shortKey)
	if err == nil {
		t.Error("expected error for short key")
	}
}

func TestExtractOIDXIDFromKey(t *testing.T) {
	searcher := NewBinarySearcher(binary.LittleEndian)

	// Test valid composite key
	keyData := make([]byte, 16)
	binary.LittleEndian.PutUint64(keyData[0:8], 100)
	binary.LittleEndian.PutUint64(keyData[8:16], 50)

	oid, xid, err := searcher.extractOIDXIDFromKey(keyData)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if oid != 100 || xid != 50 {
		t.Errorf("expected (100,50), got (%d,%d)", oid, xid)
	}

	// Test short key
	shortKey := make([]byte, 8)
	_, _, err = searcher.extractOIDXIDFromKey(shortKey)
	if err == nil {
		t.Error("expected error for short key")
	}
}
