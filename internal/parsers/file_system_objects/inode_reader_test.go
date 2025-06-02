package file_system_objects

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestNewInodeReader(t *testing.T) {
	// Create test data
	keyData := make([]byte, 8)
	binary.LittleEndian.PutUint64(keyData[0:8], uint64(types.ApfsTypeInode)<<types.ObjTypeShift|123)

	valueData := make([]byte, 98) // Minimum size
	offset := 0

	// Parent ID
	binary.LittleEndian.PutUint64(valueData[offset:offset+8], 456)
	offset += 8

	// Private ID
	binary.LittleEndian.PutUint64(valueData[offset:offset+8], 789)
	offset += 8

	// Create time (current time in nanoseconds)
	createTime := uint64(time.Now().UnixNano())
	binary.LittleEndian.PutUint64(valueData[offset:offset+8], createTime)
	offset += 8

	// Mod time
	binary.LittleEndian.PutUint64(valueData[offset:offset+8], createTime+1000)
	offset += 8

	// Change time
	binary.LittleEndian.PutUint64(valueData[offset:offset+8], createTime+2000)
	offset += 8

	// Access time
	binary.LittleEndian.PutUint64(valueData[offset:offset+8], createTime+3000)
	offset += 8

	// Internal flags
	binary.LittleEndian.PutUint64(valueData[offset:offset+8], 0)
	offset += 8

	// Nchildren/nlink
	binary.LittleEndian.PutUint32(valueData[offset:offset+4], 5)
	offset += 4

	// Default protection class
	binary.LittleEndian.PutUint32(valueData[offset:offset+4], 0)
	offset += 4

	// Write generation counter
	binary.LittleEndian.PutUint32(valueData[offset:offset+4], 1)
	offset += 4

	// BSD flags
	binary.LittleEndian.PutUint32(valueData[offset:offset+4], 0)
	offset += 4

	// Owner
	binary.LittleEndian.PutUint32(valueData[offset:offset+4], 501)
	offset += 4

	// Group
	binary.LittleEndian.PutUint32(valueData[offset:offset+4], 20)
	offset += 4

	// Mode (directory: 0x4000 | 0755)
	binary.LittleEndian.PutUint16(valueData[offset:offset+2], 0x4000|0755)
	offset += 2

	// Pad1
	binary.LittleEndian.PutUint16(valueData[offset:offset+2], 0)
	offset += 2

	// Uncompressed size
	binary.LittleEndian.PutUint64(valueData[offset:offset+8], 0)

	reader, err := NewInodeReader(keyData, valueData, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewInodeReader failed: %v", err)
	}

	// Test basic properties
	if reader.ObjectIdentifier() != 123 {
		t.Errorf("Expected object identifier 123, got %d", reader.ObjectIdentifier())
	}

	if reader.ObjectType() != types.ApfsTypeInode {
		t.Errorf("Expected object type %d, got %d", types.ApfsTypeInode, reader.ObjectType())
	}

	if reader.ParentID() != 456 {
		t.Errorf("Expected parent ID 456, got %d", reader.ParentID())
	}

	if reader.PrivateID() != 789 {
		t.Errorf("Expected private ID 789, got %d", reader.PrivateID())
	}

	if !reader.IsDirectory() {
		t.Error("Expected inode to be a directory")
	}

	if reader.NumberOfChildren() != 5 {
		t.Errorf("Expected 5 children, got %d", reader.NumberOfChildren())
	}

	if reader.Owner() != 501 {
		t.Errorf("Expected owner 501, got %d", reader.Owner())
	}

	if reader.Group() != 20 {
		t.Errorf("Expected group 20, got %d", reader.Group())
	}

	// Test timestamps
	expectedCreateTime := time.Unix(0, int64(createTime))
	if !reader.CreationTime().Equal(expectedCreateTime) {
		t.Errorf("Expected create time %v, got %v", expectedCreateTime, reader.CreationTime())
	}
}

func TestNewInodeReaderInvalidKey(t *testing.T) {
	keyData := make([]byte, 4) // Too small
	valueData := make([]byte, 98)

	_, err := NewInodeReader(keyData, valueData, binary.LittleEndian)
	if err == nil {
		t.Error("Expected error for invalid key data")
	}
}

func TestNewInodeReaderInvalidValue(t *testing.T) {
	keyData := make([]byte, 8)
	binary.LittleEndian.PutUint64(keyData[0:8], uint64(types.ApfsTypeInode)<<types.ObjTypeShift|123)

	valueData := make([]byte, 50) // Too small

	_, err := NewInodeReader(keyData, valueData, binary.LittleEndian)
	if err == nil {
		t.Error("Expected error for invalid value data")
	}
}
